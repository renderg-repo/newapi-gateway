package controller

import (
	"fmt"
	"net/http"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/i18n"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/sidecar/service"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// ---------- request / response DTOs ----------

type getQRCodeResponse struct {
	Ticket     string `json:"ticket"`
	QRCodeURL  string `json:"qrcode_url"`
	ExpiresIn  int    `json:"expires_in"` // seconds
}

type checkQRCodeRequest struct {
	Ticket string `json:"ticket"`
}

type checkQRCodeResponse struct {
	Status   string `json:"status"`
	OpenID   string `json:"openid,omitempty"`
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

// GetQRCode initiates a new QR-code login session by calling HPC system.
func GetQRCode(c *gin.Context) {
	if common.WeChatHpcServerAddress == "" {
		common.ApiErrorI18n(c, i18n.MsgFeatureDisabled)
		return
	}

	// Create session by calling HPC system API
	session, err := service.CreateQRCodeSession()
	if err != nil {
		common.SysLog(fmt.Sprintf("Failed to create WeChat QR code: %v", err))
		common.ApiErrorI18n(c, i18n.MsgOperationFailed)
		return
	}

	// Construct QR code URL using ticket from HPC
	qrcodeURL := fmt.Sprintf("https://mp.weixin.qq.com/cgi-bin/showqrcode?ticket=%s", session.Ticket)

	expiresIn := 300 // QR code expires in 5 minutes
	common.ApiSuccess(c, getQRCodeResponse{
		Ticket:    session.Ticket,
		QRCodeURL: qrcodeURL,
		ExpiresIn: expiresIn,
	})
}

// CheckQRCode polls the status of a QR-code login session by calling HPC system.
func CheckQRCode(c *gin.Context) {
	if common.WeChatHpcServerAddress == "" {
		common.ApiErrorI18n(c, i18n.MsgFeatureDisabled)
		return
	}

	var req checkQRCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Ticket == "" {
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}

	// Check QR code status by calling HPC system API
	session, err := service.CheckQRCodeStatus(req.Ticket)
	if err != nil {
		common.SysLog(fmt.Sprintf("Failed to check QR code status: %v", err))
		common.ApiErrorI18n(c, i18n.MsgOperationFailed)
		return
	}

	resp := checkQRCodeResponse{
		Status: string(session.Status),
	}

	// When confirmed, check if user is bound
	if session.Status == service.QRCodeStatusConfirmed && session.OpenID != "" {
		// Look up user by openid (WeChatId)
		user := model.User{WeChatId: session.OpenID}
		if model.IsWeChatIdAlreadyTaken(session.OpenID) {
			if err := user.FillUserByWeChatId(); err != nil {
				common.SysLog(fmt.Sprintf("Failed to find user by WeChatId: %v", err))
				common.ApiErrorI18n(c, i18n.MsgDatabaseError)
				return
			} else if user.Id == 0 {
				common.SysLog("User with WeChatId has been deleted")
				common.ApiErrorI18n(c, i18n.MsgOperationFailed)
				return
			}

			if user.Status != common.UserStatusEnabled {
				common.SysLog(fmt.Sprintf("User %d is banned", user.Id))
				common.ApiErrorI18n(c, i18n.MsgUserAccountDisabled)
				return
			}

			// Update session with user id
			session.UserID = user.Id
			service.SaveSession(session)

			resp.UserID = user.Id
			resp.Username = user.Username
			resp.Role = user.Role
			resp.Group = user.Group

			// Set session cookie
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
		} else {
			// OpenID not bound to any user, return need_bind status
			resp.Status = "need_bind"
			resp.OpenID = session.OpenID
		}
	}

	common.ApiSuccess(c, resp)
}

type hpcCallbackRequest struct {
	Ticket string `json:"ticket" binding:"required"`
	Openid string `json:"openid" binding:"required"`
}

// HpcCallback receives callback from HPC system with WeChat openid.
func HpcCallback(c *gin.Context) {
	if common.WeChatHpcServerAddress == "" || common.WeChatHpcCallbackToken == "" {
		common.ApiErrorI18n(c, i18n.MsgFeatureDisabled)
		return
	}

	var req hpcCallbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}

	// Verify token from header
	token := c.GetHeader("API-TOKEN")
	if token != common.WeChatHpcCallbackToken {
		common.SysLog(fmt.Sprintf("Invalid HPC callback token from %s", c.ClientIP()))
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}

	_, err := service.UpdateSessionByCallback(req.Ticket, req.Openid)
	if err != nil {
		common.SysLog(fmt.Sprintf("Failed to update session from HPC callback: %v", err))
		common.ApiErrorI18n(c, i18n.MsgOperationFailed)
		return
	}

	common.ApiSuccess(c, nil)
}
