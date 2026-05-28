package router

import (
	"github.com/QuantumNous/new-api/sidecar/controller"
	"github.com/gin-gonic/gin"
)

func RegisterWeChatQRCodeRoutes(apiRouter *gin.RouterGroup) {
	// WeChat QR scan login via HPC (智算) system
	weixinRoute := apiRouter.Group("/weixin")
	{
		weixinRoute.GET("/getQrCode", controller.GetQRCode)
		weixinRoute.POST("/checkQrCode", controller.CheckQRCode)
		weixinRoute.POST("/hpcCallback", controller.HpcCallback)
	}
}
