package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/i18n"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/sidecar/service"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// ---------- request / response DTOs ----------

type createQRCodeResponse struct {
	Ticket    string `json:"ticket"`
	QRCodeURL string `json:"qrcode_url"`
	ExpiresIn int    `json:"expires_in"` // seconds
}

type pollQRCodeResponse struct {
	Status     string `json:"status"`
	UserID     int    `json:"user_id,omitempty"`
	Username   string `json:"username,omitempty"`
	Role       int    `json:"role,omitempty"`
	Group      string `json:"group,omitempty"`
	WeChatId   string `json:"wechat_id,omitempty"`
}

type callbackRequest struct {
	Ticket   string `json:"ticket"`
	WeChatId string `json:"wechat_id"`
}

// ---------- helpers ----------

func getWeChatIdByCode(code string) (string, error) {
	if code == "" {
		return "", errors.New("invalid code")
	}
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/wechat/user?code=%s", common.WeChatServerAddress, code), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", common.WeChatServerToken)
	client := http.Client{Timeout: 5 * time.Second}
	httpResponse, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer httpResponse.Body.Close()

	var res struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
		Data    string `json:"data"`
	}
	if err = json.NewDecoder(httpResponse.Body).Decode(&res); err != nil {
		return "", err
	}
	if !res.Success {
		return "", errors.New(res.Message)
	}
	if res.Data == "" {
		return "", errors.New("verification code expired or invalid")
	}
	return res.Data, nil
}

// wechatSetupLogin copies the same logic from the main controller/wechat.go
// (the original is unexported, so we replicate it here per sidecar convention)
func wechatSetupLogin(user *model.User, c *gin.Context) {
	session := sessions.Default(c)
	session.Set("id", user.Id)
	session.Set("username", user.Username)
	session.Set("role", user.Role)
	session.Set("status", user.Status)
	session.Set("group", user.Group)
	err := session.Save()
	if err != nil {
		common.ApiErrorI18n(c, i18n.MsgUserSessionSaveFailed)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "",
		"success": true,
		"data": map[string]any{
			"id":           user.Id,
			"username":     user.Username,
			"display_name": user.DisplayName,
			"role":         user.Role,
			"status":       user.Status,
			"group":        user.Group,
		},
	})
}

// ---------- handlers ----------

// CreateQRCodeLogin initiates a new QR-code login session.
// POST /api/wechat/qrcode/create
func CreateQRCodeLogin(c *gin.Context) {
	if !common.WeChatAuthEnabled {
		common.ApiErrorI18n(c, i18n.MsgFeatureDisabled)
		return
	}

	// Try to obtain a dynamic QR code URL from the external WeChat server.
	qrCodeURL := common.WeChatAccountQRCodeImageURL
	session := service.CreateQRCodeSession()

	// Attempt to generate a scene-based QR code via the external server.
	if common.WeChatServerAddress != "" {
		sceneReqURL := fmt.Sprintf("%s/api/wechat/qrcode/create?scene_id=%s", common.WeChatServerAddress, session.Ticket)
		req, err := http.NewRequest("POST", sceneReqURL, nil)
		if err == nil {
			req.Header.Set("Authorization", common.WeChatServerToken)
			client := http.Client{Timeout: 5 * time.Second}
			if resp, err := client.Do(req); err == nil {
				defer resp.Body.Close()
				var qrResp struct {
					Success   bool   `json:"success"`
					QRCodeURL string `json:"qrcode_url"`
				}
				if json.NewDecoder(resp.Body).Decode(&qrResp) == nil && qrResp.Success && qrResp.QRCodeURL != "" {
					qrCodeURL = qrResp.QRCodeURL
				}
			}
		}
	}

	expiresIn := int(time.Until(session.ExpiresAt).Seconds())
	common.ApiSuccess(c, createQRCodeResponse{
		Ticket:    session.Ticket,
		QRCodeURL: qrCodeURL,
		ExpiresIn: expiresIn,
	})
}

// PollQRCodeLogin polls the status of a QR-code login session.
// GET /api/wechat/qrcode/poll?ticket=xxx
func PollQRCodeLogin(c *gin.Context) {
	if !common.WeChatAuthEnabled {
		common.ApiErrorI18n(c, i18n.MsgFeatureDisabled)
		return
	}

	ticket := c.Query("ticket")
	if ticket == "" {
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}

	session, ok := service.GetQRCodeSession(ticket)
	if !ok {
		common.ApiErrorI18n(c, i18n.MsgNotFound)
		return
	}

	resp := pollQRCodeResponse{
		Status: string(session.Status),
	}

	// When the session is confirmed, set up the login session cookie.
	if session.Status == service.QRCodeStatusConfirmed && session.UserID > 0 {
		user := model.User{Id: session.UserID}
		if err := user.FillUserById(); err == nil {
			resp.UserID = user.Id
			resp.Username = user.Username
			resp.Role = user.Role
			resp.Group = user.Group
			resp.WeChatId = user.WeChatId

			// Set session cookie so the frontend is logged in.
			sess := sessions.Default(c)
			sess.Set("id", user.Id)
			sess.Set("username", user.Username)
			sess.Set("role", user.Role)
			sess.Set("status", user.Status)
			sess.Set("group", user.Group)
			_ = sess.Save()
		}
	}

	common.ApiSuccess(c, resp)
}

// HandleQRCodeCallback receives scan events from the external WeChat bridge server.
// POST /api/wechat/qrcode/callback
func HandleQRCodeCallback(c *gin.Context) {
	if !common.WeChatAuthEnabled {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "WeChat login is disabled",
		})
		return
	}

	var req callbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "invalid request",
		})
		return
	}

	if req.Ticket == "" || req.WeChatId == "" {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "ticket and wechat_id are required",
		})
		return
	}

	// Validate the ticket exists and is still pending.
	session, ok := service.GetQRCodeSession(req.Ticket)
	if !ok {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "ticket not found or expired",
		})
		return
	}
	if session.Status != service.QRCodeStatusPending {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": fmt.Sprintf("ticket already %s", session.Status),
		})
		return
	}

	// Look up existing user or create a new one.
	user := model.User{WeChatId: req.WeChatId}
	if model.IsWeChatIdAlreadyTaken(req.WeChatId) {
		if err := user.FillUserByWeChatId(); err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}
		if user.Id == 0 {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "user has been deleted",
			})
			return
		}
	} else if common.RegisterEnabled {
		user.Username = "wechat_" + strconv.Itoa(model.GetMaxUserId()+1)
		user.DisplayName = "WeChat User"
		user.Role = common.RoleCommonUser
		user.Status = common.UserStatusEnabled

		if err := user.Insert(0); err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}
	} else {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "registration is disabled",
		})
		return
	}

	if user.Status != common.UserStatusEnabled {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "user has been banned",
		})
		return
	}

	// Update the session with the user info.
	service.UpdateQRCodeSessionStatus(req.Ticket, service.QRCodeStatusConfirmed, req.WeChatId, user.Id)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
}

// HandleCodeBasedLogin is for the code-based WeChat login flow (used when
// the external WeChat server provides a code instead of managing QR codes).
// POST /api/wechat/qrcode/exchange
// Body: { code: "xxx" }
func HandleCodeBasedLogin(c *gin.Context) {
	if !common.WeChatAuthEnabled {
		common.ApiErrorI18n(c, i18n.MsgFeatureDisabled)
		return
	}

	var req struct {
		Code string `json:"code"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Code == "" {
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}

	// Use the existing external server's /api/wechat/user endpoint to exchange the code.
	wechatId, err := getWeChatIdByCode(req.Code)
	if err != nil {
		common.ApiErrorI18n(c, i18n.MsgOperationFailed)
		return
	}

	// Look up or create user.
	user := model.User{WeChatId: wechatId}
	if model.IsWeChatIdAlreadyTaken(wechatId) {
		if err := user.FillUserByWeChatId(); err != nil {
			common.ApiErrorI18n(c, i18n.MsgDatabaseError)
			return
		}
		if user.Id == 0 {
			common.ApiErrorI18n(c, i18n.MsgNotFound)
			return
		}
	} else if common.RegisterEnabled {
		user.Username = "wechat_" + strconv.Itoa(model.GetMaxUserId()+1)
		user.DisplayName = "WeChat User"
		user.Role = common.RoleCommonUser
		user.Status = common.UserStatusEnabled
		if err := user.Insert(0); err != nil {
			common.ApiError(c, err)
			return
		}
	} else {
		common.ApiErrorI18n(c, i18n.MsgUserRegisterDisabled)
		return
	}

	if user.Status != common.UserStatusEnabled {
		common.ApiErrorI18n(c, i18n.MsgUserAccountDisabled)
		return
	}

	wechatSetupLogin(&user, c)
}
