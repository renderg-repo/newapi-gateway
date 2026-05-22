package router

import (
	"github.com/QuantumNous/new-api/middleware"
	"github.com/QuantumNous/new-api/sidecar/controller"
	"github.com/gin-gonic/gin"
)

func RegisterWeChatQRCodeRoutes(apiRouter *gin.RouterGroup) {
	// Direct WeChat Official Account API integration for QR scan login
	weixinRoute := apiRouter.Group("/weixin")
	{
		weixinRoute.GET("/getQrCode", middleware.CriticalRateLimit(), controller.GetQRCode)
		weixinRoute.Any("/receiveMessage", controller.ReceiveMessage)
		weixinRoute.POST("/checkQrCode", middleware.CriticalRateLimit(), controller.CheckQRCode)
	}

	// Legacy code-based login (kept for backward compatibility)
	apiRouter.POST("/wechat/qrcode/exchange", middleware.CriticalRateLimit(), controller.WeChatLoginByCode)
}
