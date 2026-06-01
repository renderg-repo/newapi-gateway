package controller

import (
	"github.com/QuantumNous/new-api/i18n"
	"github.com/QuantumNous/new-api/sidecar/service"

	"github.com/QuantumNous/new-api/common"
	"github.com/gin-gonic/gin"
)

type CaptchaVerifyRequest struct {
	CaptchaId   string `json:"captcha_id" binding:"required"`
	CaptchaCode string `json:"captcha_code" binding:"required"`
}

func GetCaptcha(c *gin.Context) {
	resp, err := service.GenerateCaptchaAuto()
	if err != nil {
		common.ApiErrorI18n(c, i18n.MsgUserVerificationCodeError)
		return
	}

	common.ApiSuccess(c, resp)
}

func VerifyCaptcha(c *gin.Context) {
	var req CaptchaVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}

	valid := service.VerifyCaptchaAuto(req.CaptchaId, req.CaptchaCode)
	if !valid {
		common.ApiErrorI18n(c, i18n.MsgUserVerificationCodeError)
		return
	}

	common.ApiSuccess(c, gin.H{
		"valid": true,
	})
}
