package router

import (
	"github.com/QuantumNous/new-api/middleware"
	"github.com/QuantumNous/new-api/sidecar/controller"
	"github.com/gin-gonic/gin"
)

func RegisterNativePaymentRoutes(apiRouter *gin.RouterGroup) {
	// Webhooks (no auth)
	apiRouter.POST("/wechatpay/notify", controller.WechatPayNotify)
	apiRouter.POST("/alipay/notify", controller.AlipayNotify)
	apiRouter.GET("/alipay/return", controller.AlipayReturn)

	// User payment endpoints
	userRoute := apiRouter.Group("/user")
	{
		userRoute.POST("/wechatpay/pay", middleware.UserAuth(), middleware.CriticalRateLimit(), controller.RequestWechatPay)
		userRoute.POST("/alipay/pay", middleware.UserAuth(), middleware.CriticalRateLimit(), controller.RequestAlipay)
	}
}
