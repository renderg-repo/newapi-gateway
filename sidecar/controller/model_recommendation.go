package controller

import (
	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/i18n"
	"github.com/QuantumNous/new-api/sidecar/service"
	"github.com/gin-gonic/gin"
)

// GetModelRecommendations 获取热门模型推荐列表（公开接口）
func GetModelRecommendations(c *gin.Context) {
	models, err := service.GetModelRecommendations()
	if err != nil {
		common.ApiErrorI18n(c, i18n.MsgDatabaseError)
		return
	}
	common.ApiSuccess(c, models)
}
