package router

import (
	"github.com/QuantumNous/new-api/sidecar/controller"
	"github.com/gin-gonic/gin"
)

func RegisterCaptchaRoutes(r *gin.RouterGroup) {
	captchaGroup := r.Group("/captcha")
	{
		captchaGroup.GET("/get", controller.GetCaptcha)
		captchaGroup.POST("/verify", controller.VerifyCaptcha)
	}
}
