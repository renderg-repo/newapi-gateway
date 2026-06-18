package router

import (
	"github.com/QuantumNous/new-api/middleware"
	"github.com/QuantumNous/new-api/sidecar/controller"
	"github.com/gin-gonic/gin"
)

func RegisterModelPerformanceRoutes(apiRouter *gin.RouterGroup) {
	// 模型性能分析接口（需要登录）
	performanceRoute := apiRouter.Group("/dashboard")
	performanceRoute.Use(middleware.UserAuth())
	{
		performanceRoute.GET("/model-performance", controller.GetModelPerformance)
	}
}
