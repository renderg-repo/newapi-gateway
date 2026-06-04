package controller

import (
	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/i18n"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/sidecar/service"
	"github.com/gin-gonic/gin"
)

type SendSMSRequest struct {
	Phone string `json:"phone" binding:"required"`
	Type  string `json:"type" binding:"required,oneof=login register rebind"`
}

func SendSMSCode(c *gin.Context) {
	if !common.SMSEnabled {
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}

	var req SendSMSRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}

	// check user existence based on type
	phoneExists := model.IsPhoneAlreadyTaken(req.Phone)
	switch req.Type {
	case "login":
		if !phoneExists {
			common.ApiErrorI18n(c, i18n.MsgUserNotExists)
			return
		}
	case "register":
		if phoneExists {
			common.ApiErrorI18n(c, i18n.MsgUserExists)
			return
		}
	case "rebind":
		// rebind doesn't need check here, will check in bind handler
	}

	// rate limit: same phone 60s
	if !service.CanSendToPhone(req.Phone) {
		common.ApiErrorI18n(c, i18n.MsgTooManyRequests)
		return
	}

	// generate 6-digit numeric code
	code := common.GenerateNumericVerificationCode(6)
	common.RegisterVerificationCodeWithKey(req.Phone, code, common.SMSVerificationPurpose)

	if err := service.SendSMS(req.Phone, code); err != nil {
		common.SysLog("sms send failed: " + err.Error())
		common.ApiErrorI18n(c, i18n.MsgSMSSendFailed)
		return
	}

	common.ApiSuccess(c, nil)
}
