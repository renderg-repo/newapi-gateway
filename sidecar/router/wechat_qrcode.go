package router

import (
	"github.com/QuantumNous/new-api/middleware"
	"github.com/QuantumNous/new-api/sidecar/controller"
	"github.com/gin-gonic/gin"
)

func RegisterWeChatQRCodeRoutes(apiRouter *gin.RouterGroup) {
	// QR-code-based WeChat scan login endpoints
	qrcodeRoute := apiRouter.Group("/wechat/qrcode")
	{
		qrcodeRoute.POST("/create", middleware.CriticalRateLimit(), controller.CreateQRCodeLogin)
		qrcodeRoute.GET("/poll", middleware.CriticalRateLimit(), controller.PollQRCodeLogin)
		qrcodeRoute.POST("/callback", middleware.CriticalRateLimit(), controller.HandleQRCodeCallback)
		qrcodeRoute.POST("/exchange", middleware.CriticalRateLimit(), controller.HandleCodeBasedLogin)
	}
}
