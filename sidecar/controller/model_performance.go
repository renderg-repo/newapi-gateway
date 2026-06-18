package controller

import (
	"fmt"
	"strconv"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/i18n"
	"github.com/QuantumNous/new-api/sidecar/service"
	"github.com/gin-gonic/gin"
)

// GetModelPerformance 获取模型性能分析数据
// GET /api/dashboard/model-performance?start_timestamp=xxx&end_timestamp=xxx
func GetModelPerformance(c *gin.Context) {
	userId := c.GetInt("id")
	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)

	// 判断时间跨度是否超过 1 个月
	if endTimestamp-startTimestamp > 2592000 {
		common.ApiError(c, fmt.Errorf("时间跨度不能超过 1 个月"))
		return
	}

	result, err := service.GetModelPerformance(userId, startTimestamp, endTimestamp)
	if err != nil {
		common.SysLog("failed to get model performance: " + err.Error())
		common.ApiErrorI18n(c, i18n.MsgDatabaseError)
		return
	}

	common.ApiSuccess(c, result)
}
