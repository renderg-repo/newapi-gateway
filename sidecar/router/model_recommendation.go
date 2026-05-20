package router

import (
	"github.com/QuantumNous/new-api/sidecar/controller"
	"github.com/gin-gonic/gin"
)

func RegisterModelRecommendationRoutes(apiRouter *gin.RouterGroup) {
	apiRouter.GET("/models/recommended", controller.GetModelRecommendations)
}
