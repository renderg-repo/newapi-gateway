package controller

import (
	"errors"
	"net/http"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/i18n"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PhoneLoginRequest struct {
	Phone string `json:"phone" binding:"required"`
	Code  string `json:"code" binding:"required"`
}

type PhoneRegisterRequest struct {
	Phone    string `json:"phone" binding:"required"`
	Code     string `json:"code" binding:"required"`
	Password string `json:"password"`
	AffCode  string `json:"aff_code"`
}

type PhoneBindRequest struct {
	Phone string `json:"phone" binding:"required"`
	Code  string `json:"code" binding:"required"`
}

func setupLogin(user *model.User, c *gin.Context) {
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

func PhoneLogin(c *gin.Context) {
	if !common.SMSEnabled {
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}

	var req PhoneLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}

	if !common.VerifyCodeWithKey(req.Phone, req.Code, common.SMSVerificationPurpose) {
		common.ApiErrorI18n(c, i18n.MsgUserVerificationCodeError)
		return
	}
	common.DeleteKey(req.Phone, common.SMSVerificationPurpose)

	user := model.User{Phone: req.Phone}
	if err := user.FillUserByPhone(); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			common.ApiErrorI18n(c, i18n.MsgUserNotExists)
			return
		}
		common.ApiErrorI18n(c, i18n.MsgDatabaseError)
		return
	}

	if user.Status != common.UserStatusEnabled {
		common.ApiErrorI18n(c, i18n.MsgUserAccountDisabled)
		return
	}

	setupLogin(&user, c)
}

func PhoneRegister(c *gin.Context) {
	if !common.SMSEnabled {
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}
	if !common.RegisterEnabled {
		common.ApiErrorI18n(c, i18n.MsgUserRegisterDisabled)
		return
	}

	var req PhoneRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}

	if !common.VerifyCodeWithKey(req.Phone, req.Code, common.SMSVerificationPurpose) {
		common.ApiErrorI18n(c, i18n.MsgUserVerificationCodeError)
		return
	}
	common.DeleteKey(req.Phone, common.SMSVerificationPurpose)

	if model.IsPhoneAlreadyTaken(req.Phone) {
		common.ApiErrorI18n(c, i18n.MsgUserExists)
		return
	}

	username := "phone_" + req.Phone
	exist, _ := model.CheckUserExistOrDeleted(username, "")
	for exist {
		username = "phone_" + req.Phone + "_" + common.GetRandomString(4)
		exist, _ = model.CheckUserExistOrDeleted(username, "")
	}

	password := req.Password
	if password == "" {
		password = common.GetRandomString(16)
	}

	inviterId, _ := model.GetUserIdByAffCode(req.AffCode)

	user := model.User{
		Username:    username,
		Password:    password,
		DisplayName: username,
		Phone:       req.Phone,
		Role:        common.RoleCommonUser,
		InviterId:   inviterId,
	}

	if err := user.Insert(inviterId); err != nil {
		common.ApiError(c, err)
		return
	}

	// re-fetch user to get generated fields
	if err := user.FillUserByPhone(); err != nil {
		common.ApiError(c, err)
		return
	}

	setupLogin(&user, c)
}

func PhoneBind(c *gin.Context) {
	if !common.SMSEnabled {
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}

	var req PhoneBindRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}

	if !common.VerifyCodeWithKey(req.Phone, req.Code, common.SMSVerificationPurpose) {
		common.ApiErrorI18n(c, i18n.MsgUserVerificationCodeError)
		return
	}
	common.DeleteKey(req.Phone, common.SMSVerificationPurpose)

	if model.IsPhoneAlreadyTaken(req.Phone) {
		common.ApiErrorI18n(c, i18n.MsgUserExists)
		return
	}

	userId := c.GetInt("id")
	if err := model.DB.Model(&model.User{}).Where("id = ?", userId).Update("phone", req.Phone).Error; err != nil {
		common.ApiErrorI18n(c, i18n.MsgDatabaseError)
		return
	}

	common.ApiSuccess(c, nil)
}

func PhoneRebind(c *gin.Context) {
	if !common.SMSEnabled {
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}

	var req PhoneBindRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}

	if !common.VerifyCodeWithKey(req.Phone, req.Code, common.SMSVerificationPurpose) {
		common.ApiErrorI18n(c, i18n.MsgUserVerificationCodeError)
		return
	}
	common.DeleteKey(req.Phone, common.SMSVerificationPurpose)

	if model.IsPhoneAlreadyTaken(req.Phone) {
		common.ApiErrorI18n(c, i18n.MsgUserExists)
		return
	}

	userId := c.GetInt("id")
	if err := model.DB.Model(&model.User{}).Where("id = ?", userId).Update("phone", req.Phone).Error; err != nil {
		common.ApiErrorI18n(c, i18n.MsgDatabaseError)
		return
	}

	common.ApiSuccess(c, nil)
}
