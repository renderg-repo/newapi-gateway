package service

import (
	"github.com/QuantumNous/new-api/model"
)

// ModelPerformance 模型性能统计数据
type ModelPerformance struct {
	ModelName         string  `json:"model_name"`
	AvgLatencyMs      int64   `json:"avg_latency_ms"`
	AvgLatencySeconds float64 `json:"avg_latency_seconds"`
	InputTokens       int64   `json:"input_tokens"`
	OutputTokens      int64   `json:"output_tokens"`
	TotalTokens       int64   `json:"total_tokens"`
	RequestCount      int64   `json:"request_count"`
}

// ModelPerformanceResult 性能分析结果
type ModelPerformanceResult struct {
	Models []ModelPerformance `json:"models"`
}

// GetModelPerformance 获取模型性能分析数据
// userId: 用户ID
// startTimestamp: 开始时间戳
// endTimestamp: 结束时间戳
func GetModelPerformance(userId int, startTimestamp int64, endTimestamp int64) (*ModelPerformanceResult, error) {
	// 查询 logs 表按 model_name 聚合统计
	// type=2 表示消费日志
	var results []struct {
		ModelName    string  `json:"model_name"`
		AvgLatencyMs float64 `json:"avg_latency_ms"`
		InputTokens  int64   `json:"input_tokens"`
		OutputTokens int64   `json:"output_tokens"`
		TotalTokens  int64   `json:"total_tokens"`
		RequestCount int64   `json:"request_count"`
	}

	err := model.LOG_DB.Table("logs").
		Select("model_name, AVG(use_time) as avg_latency_ms, SUM(prompt_tokens) as input_tokens, SUM(completion_tokens) as output_tokens, SUM(prompt_tokens) + SUM(completion_tokens) as total_tokens, COUNT(*) as request_count").
		Where("type = ? AND user_id = ? AND created_at >= ? AND created_at <= ?", model.LogTypeConsume, userId, startTimestamp, endTimestamp).
		Group("model_name").
		Having("COUNT(*) > 0").
		Order("request_count DESC").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	models := make([]ModelPerformance, 0, len(results))
	for _, r := range results {
		models = append(models, ModelPerformance{
			ModelName:         r.ModelName,
			AvgLatencyMs:      int64(r.AvgLatencyMs),
			AvgLatencySeconds: float64(int64(r.AvgLatencyMs)) / 1000.0,
			InputTokens:       r.InputTokens,
			OutputTokens:      r.OutputTokens,
			TotalTokens:       r.TotalTokens,
			RequestCount:      r.RequestCount,
		})
	}

	return &ModelPerformanceResult{Models: models}, nil
}
