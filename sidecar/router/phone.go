package router

import (
	"github.com/QuantumNous/new-api/middleware"
	"github.com/QuantumNous/new-api/sidecar/controller"
	"github.com/gin-gonic/gin"
)

func RegisterPhoneAuthRoutes(apiRouter *gin.RouterGroup) {
	// SMS code sending endpoint
	apiRouter.POST("/sms/send", middleware.CriticalRateLimit(), controller.SendSMSCode)

	// Phone authentication endpoints
	userRoute := apiRouter.Group("/user")
	{
		userRoute.GET("/check-username", middleware.CriticalRateLimit(), controller.CheckUsername)
		userRoute.POST("/login/phone", middleware.CriticalRateLimit(), middleware.TurnstileCheck(), controller.PhoneLogin)
		userRoute.POST("/register/phone", middleware.CriticalRateLimit(), middleware.TurnstileCheck(), controller.PhoneRegister)

		// Bind / rebind phone for logged-in user
		selfRoute := userRoute.Group("/")
		selfRoute.Use(middleware.UserAuth())
		{
			selfRoute.POST("/phone/bind", middleware.CriticalRateLimit(), controller.PhoneBind)
			selfRoute.PUT("/phone/rebind", middleware.CriticalRateLimit(), controller.PhoneRebind)
		}
	}
}
