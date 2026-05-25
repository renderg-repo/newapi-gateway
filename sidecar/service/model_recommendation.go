package service

import (
	"sort"
	"sync"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
)

const (
	recommendationCacheTTL = 5 * time.Minute
	optionKey              = "ModelRecommendations"
)

// RecommendationConfig 运营配置的原始结构
type RecommendationConfig struct {
	ModelName   string `json:"model_name"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	Order       int    `json:"order"`
}

// ModelRecommendation 对外返回的推荐模型 DTO
type ModelRecommendation struct {
	ModelName   string `json:"model_name"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	VendorName  string `json:"vendor_name"`
	VendorIcon  string `json:"vendor_icon,omitempty"`
	Icon        string `json:"icon,omitempty"`
	Order       int    `json:"order"`
}

type recommendationCacheItem struct {
	expiresAt time.Time
	data      []ModelRecommendation
}

var (
	recommendationCacheMu sync.Mutex
	recommendationCache   recommendationCacheItem
)

// GetModelRecommendations 获取热门模型推荐列表
func GetModelRecommendations() ([]ModelRecommendation, error) {
	now := time.Now()

	recommendationCacheMu.Lock()
	if recommendationCache.expiresAt.After(now) && recommendationCache.data != nil {
		recommendationCacheMu.Unlock()
		return recommendationCache.data, nil
	}
	recommendationCacheMu.Unlock()

	data, err := buildRecommendations()
	if err != nil {
		return nil, err
	}

	recommendationCacheMu.Lock()
	recommendationCache = recommendationCacheItem{
		expiresAt: now.Add(recommendationCacheTTL),
		data:      data,
	}
	recommendationCacheMu.Unlock()

	return data, nil
}

func buildRecommendations() ([]ModelRecommendation, error) {
	// 1. 读取配置
	raw := common.OptionMap[optionKey]
	if raw == "" {
		return []ModelRecommendation{}, nil
	}

	var configs []RecommendationConfig
	if err := common.UnmarshalJsonStr(raw, &configs); err != nil {
		common.SysLog("failed to unmarshal ModelRecommendations: " + err.Error())
		return []ModelRecommendation{}, nil
	}

	if len(configs) == 0 {
		return []ModelRecommendation{}, nil
	}

	// 2. 排序
	sort.Slice(configs, func(i, j int) bool {
		return configs[i].Order < configs[j].Order
	})

	// 3. 获取模型和厂商信息
	pricing := model.GetPricing()
	vendors := model.GetVendors()

	pricingMap := make(map[string]model.Pricing, len(pricing))
	for _, p := range pricing {
		pricingMap[p.ModelName] = p
	}

	vendorMap := make(map[int]model.PricingVendor, len(vendors))
	for _, v := range vendors {
		vendorMap[v.ID] = v
	}

	// 4. 组装结果
	result := make([]ModelRecommendation, 0, len(configs))
	for _, cfg := range configs {
		if cfg.ModelName == "" {
			continue
		}

		item := ModelRecommendation{
			ModelName:   cfg.ModelName,
			DisplayName: cfg.DisplayName,
			Description: cfg.Description,
			Order:       cfg.Order,
		}

		// 关联模型元数据
		if p, ok := pricingMap[cfg.ModelName]; ok {
			item.Icon = p.Icon
			if cfg.Description == "" {
				item.Description = p.Description
			}
			if v, ok := vendorMap[p.VendorID]; ok {
				item.VendorName = v.Name
				item.VendorIcon = v.Icon
			}
		}

		result = append(result, item)
	}

	return result, nil
}
