package controller

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/i18n"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/sidecar/service"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// ---------- request / response DTOs ----------

type getQRCodeResponse struct {
	Ticket    string `json:"ticket"`
	QRCodeURL string `json:"qrcode_url"`
	ExpiresIn int    `json:"expires_in"` // seconds
}

type checkQRCodeRequest struct {
	Ticket string `json:"ticket"`
}

type checkQRCodeResponse struct {
	Status   string `json:"status"`
	UserID   int    `json:"user_id,omitempty"`
	Username string `json:"username,omitempty"`
	Role     int    `json:"role,omitempty"`
	Group    string `json:"group,omitempty"`
}

// ---------- helpers ----------

// wechatSetupLogin sets up the session cookie after successful login.
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

// GetQRCode initiates a new QR-code login session.
// GET /api/weixin/getQrCode?type=1
func GetQRCode(c *gin.Context) {
	if common.WeChatAppID == "" || common.WeChatAppSecret == "" {
		common.ApiErrorI18n(c, i18n.MsgFeatureDisabled)
		return
	}

	// Generate scene_id and create session.
	sceneID := service.GenerateSceneID()
	session := service.CreateQRCodeSession(sceneID)

	// Create temporary QR code via WeChat API.
	_, qrcodeURL, err := service.CreateTempQRCode(sceneID)
	if err != nil {
		common.SysLog(fmt.Sprintf("Failed to create WeChat QR code: %v", err))
		common.ApiErrorI18n(c, i18n.MsgOperationFailed)
		return
	}

	expiresIn := 600 // QR code expires in 600 seconds per WeChat API
	common.ApiSuccess(c, getQRCodeResponse{
		Ticket:    session.Ticket,
		QRCodeURL: qrcodeURL,
		ExpiresIn: expiresIn,
	})
}

// ReceiveMessage handles WeChat callback for QR code scan events.
// GET/POST /api/weixin/receiveMessage
// GET: signature verification (returns echostr)
// POST: receives scan event XML, updates session status
func ReceiveMessage(c *gin.Context) {
	if common.WeChatAppID == "" || common.WeChatAppSecret == "" || common.WeChatReceiveToken == "" {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "WeChat login not configured",
		})
		return
	}

	signature := c.Query("signature")
	timestamp := c.Query("timestamp")
	nonce := c.Query("nonce")

	// Verify signature.
	if !service.VerifySignature(signature, timestamp, nonce) {
		common.SysLog(fmt.Sprintf("WeChat callback signature verification failed: sig=%s, ts=%s, nonce=%s", signature, timestamp, nonce))
		c.String(http.StatusForbidden, "signature verification failed")
		return
	}

	// GET request: echo back echostr for verification.
	if c.Request.Method == http.MethodGet {
		echostr := c.Query("echostr")
		c.String(http.StatusOK, echostr)
		return
	}

	// POST request: parse XML message.
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		common.SysLog(fmt.Sprintf("Failed to read WeChat callback body: %v", err))
		c.String(http.StatusInternalServerError, "internal error")
		return
	}

	msg, err := service.ParseWeChatMessage(body)
	if err != nil {
		common.SysLog(fmt.Sprintf("Failed to parse WeChat message: %v", err))
		c.String(http.StatusOK, "success") // still return success to WeChat
		return
	}

	// Only handle SCAN or subscribe (SCAN) events.
	event := msg.Event
	if event != "SCAN" && event != "subscribe" {
		c.String(http.StatusOK, "success")
		return
	}

	// Extract scene_id from EventKey.
	// For SCAN event: EventKey is the scene_id directly.
	// For subscribe event with QR code: EventKey is "qrscene_<scene_id>".
	sceneIDStr := msg.EventKey
	if strings.HasPrefix(sceneIDStr, "qrscene_") {
		sceneIDStr = strings.TrimPrefix(sceneIDStr, "qrscene_")
	}
	sceneID, err := strconv.Atoi(sceneIDStr)
	if err != nil {
		common.SysLog(fmt.Sprintf("Invalid scene_id in WeChat event: %s", msg.EventKey))
		c.String(http.StatusOK, "success")
		return
	}

	openID := msg.FromUserName

	// Look up session by scene_id.
	session, ok := service.GetQRCodeSessionBySceneID(sceneID)
	if !ok {
		common.SysLog(fmt.Sprintf("QR code session not found for scene_id: %d", sceneID))
		c.String(http.StatusOK, "success")
		return
	}

	// Check session is still pending.
	if session.Status != service.QRCodeStatusPending {
		common.SysLog(fmt.Sprintf("QR code session already processed: ticket=%s, status=%s", session.Ticket, session.Status))
		c.String(http.StatusOK, "success")
		return
	}

	// Look up or create user by openID (WeChatId).
	user := model.User{WeChatId: openID}
	if model.IsWeChatIdAlreadyTaken(openID) {
		if err := user.FillUserByWeChatId(); err != nil {
			common.SysLog(fmt.Sprintf("Failed to find user by WeChatId: %v", err))
			c.String(http.StatusOK, "success")
			return
		}
		if user.Id == 0 {
			common.SysLog("User with WeChatId has been deleted")
			c.String(http.StatusOK, "success")
			return
		}
	} else if common.RegisterEnabled {
		user.Username = "wechat_" + strconv.Itoa(model.GetMaxUserId()+1)
		user.DisplayName = "WeChat User"
		user.Role = common.RoleCommonUser
		user.Status = common.UserStatusEnabled
		if err := user.Insert(0); err != nil {
			common.SysLog(fmt.Sprintf("Failed to create user for WeChat login: %v", err))
			c.String(http.StatusOK, "success")
			return
		}
	} else {
		// Registration disabled: mark session as needs_bind.
		service.UpdateQRCodeSessionStatus(session.Ticket, service.QRCodeStatusScanned, openID, 0)
		c.String(http.StatusOK, "success")
		return
	}

	if user.Status != common.UserStatusEnabled {
		common.SysLog(fmt.Sprintf("User %d is banned", user.Id))
		c.String(http.StatusOK, "success")
		return
	}

	// Update session to confirmed.
	service.UpdateQRCodeSessionStatus(session.Ticket, service.QRCodeStatusConfirmed, openID, user.Id)

	// Return success to WeChat.
	c.String(http.StatusOK, "success")
}

// CheckQRCode polls the status of a QR-code login session.
// POST /api/weixin/checkQrCode
// Body: { ticket: "xxx" }
func CheckQRCode(c *gin.Context) {
	if common.WeChatAppID == "" || common.WeChatAppSecret == "" {
		common.ApiErrorI18n(c, i18n.MsgFeatureDisabled)
		return
	}

	var req checkQRCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Ticket == "" {
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}

	session, ok := service.GetQRCodeSessionByTicket(req.Ticket)
	if !ok {
		common.ApiErrorI18n(c, i18n.MsgNotFound)
		return
	}

	resp := checkQRCodeResponse{
		Status: string(session.Status),
	}

	// When confirmed, set up login session and return user info.
	if session.Status == service.QRCodeStatusConfirmed && session.UserID > 0 {
		user := model.User{Id: session.UserID}
		if err := user.FillUserById(); err != nil {
			common.ApiErrorI18n(c, i18n.MsgDatabaseError)
			return
		}
		resp.UserID = user.Id
		resp.Username = user.Username
		resp.Role = user.Role
		resp.Group = user.Group

		// Set session cookie.
		sess := sessions.Default(c)
		sess.Set("id", user.Id)
		sess.Set("username", user.Username)
		sess.Set("role", user.Role)
		sess.Set("status", user.Status)
		sess.Set("group", user.Group)
		if err := sess.Save(); err != nil {
			common.ApiErrorI18n(c, i18n.MsgUserSessionSaveFailed)
			return
		}
	}

	common.ApiSuccess(c, resp)
}

// WeChatLoginByCode handles code-based WeChat login (legacy flow).
// POST /api/wechat/qrcode/exchange
// Body: { code: "xxx" }
// This is kept for backward compatibility with the existing code-based login.
func WeChatLoginByCode(c *gin.Context) {
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

	// Exchange code for openID via WeChat OAuth API.
	appID := common.WeChatAppID
	appSecret := common.WeChatAppSecret
	if appID == "" || appSecret == "" {
		common.ApiErrorI18n(c, i18n.MsgFeatureDisabled)
		return
	}

	url := fmt.Sprintf("https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code", appID, appSecret, req.Code)
	resp, err := http.Get(url)
	if err != nil {
		common.ApiErrorI18n(c, i18n.MsgOperationFailed)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		common.ApiErrorI18n(c, i18n.MsgOperationFailed)
		return
	}

	var result struct {
		OpenID     string `json:"openid"`
		Errcode    int    `json:"errcode"`
		Errmsg     string `json:"errmsg"`
	}
	if err := common.UnmarshalJsonStr(string(body), &result); err != nil {
		common.ApiErrorI18n(c, i18n.MsgOperationFailed)
		return
	}
	if result.Errcode != 0 {
		common.SysLog(fmt.Sprintf("WeChat OAuth error: %d %s", result.Errcode, result.Errmsg))
		common.ApiErrorI18n(c, i18n.MsgOperationFailed)
		return
	}
	if result.OpenID == "" {
		common.ApiError(c, errors.New("empty openid"))
		return
	}

	// Look up or create user.
	openID := result.OpenID
	user := model.User{WeChatId: openID}
	if model.IsWeChatIdAlreadyTaken(openID) {
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