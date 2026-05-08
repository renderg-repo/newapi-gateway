package router

import (
	"github.com/QuantumNous/new-api/middleware"
	"github.com/QuantumNous/new-api/sidecar/controller"
	"github.com/gin-gonic/gin"
)

func RegisterModelCatalogRoutes(apiRouter *gin.RouterGroup) {
	// Public catalog endpoints (no auth required)
	catalogRoute := apiRouter.Group("/model-catalog")
	{
		catalogRoute.GET("/", controller.GetModelCatalog)
		catalogRoute.GET("/:model_name", controller.GetModelCatalogByName)
	}

	// Admin spec management endpoints
	adminRoute := apiRouter.Group("/model-catalog/admin/specs")
	adminRoute.Use(middleware.AdminAuth())
	{
		adminRoute.GET("/", controller.AdminListModelSpecs)
		adminRoute.GET("/:id", controller.AdminGetModelSpec)
		adminRoute.POST("/", controller.AdminCreateModelSpec)
		adminRoute.PUT("/", controller.AdminUpdateModelSpec)
		adminRoute.DELETE("/:id", controller.AdminDeleteModelSpec)
	}
}
