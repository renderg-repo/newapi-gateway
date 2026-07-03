package service

import (
	"math"
	"sort"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/model"
	perfmetrics "github.com/QuantumNous/new-api/pkg/perf_metrics"
	"github.com/QuantumNous/new-api/relay/channel/task/doubao"
	sidecarModel "github.com/QuantumNous/new-api/sidecar/model"
	"github.com/QuantumNous/new-api/setting/ratio_setting"
)

type CatalogModel struct {
	ModelName              string   `json:"model_name"`
	Description            string   `json:"description,omitempty"`
	VendorID               int      `json:"vendor_id,omitempty"`
	VendorName             string   `json:"vendor_name,omitempty"`
	VendorIcon             string   `json:"vendor_icon,omitempty"`
	ContextLength          int      `json:"context_length"`
	MaxOutputTokens        int      `json:"max_output_tokens"`
	Capabilities           []string `json:"capabilities"`
	ModelRatio             float64  `json:"model_ratio"`
	CompletionRatio        float64  `json:"completion_ratio"`
	ModelPrice             float64  `json:"model_price"`
	CacheRatio             *float64 `json:"cache_ratio,omitempty"`
	CreateCacheRatio       *float64 `json:"create_cache_ratio,omitempty"`
	ImageRatio             *float64 `json:"image_ratio,omitempty"`
	AudioRatio             *float64 `json:"audio_ratio,omitempty"`
	AudioCompletionRatio   *float64 `json:"audio_completion_ratio,omitempty"`
	EnableGroup            []string `json:"enable_groups"`
	Tags                   string   `json:"tags,omitempty"`
	SupportedEndpointTypes []string `json:"supported_endpoint_types,omitempty"`
	BillingMode            string   `json:"billing_mode,omitempty"`
	BillingExpr            string   `json:"billing_expr,omitempty"`
	Icon                   string   `json:"icon,omitempty"`
	ReleaseDate            string   `json:"release_date,omitempty"`
	KnowledgeCutoff        string   `json:"knowledge_cutoff,omitempty"`
	ParameterCount         string   `json:"parameter_count,omitempty"`
	InputPrice             float64  `json:"input_price"`
	OutputPrice            float64  `json:"output_price"`
	HasVariablePricing     bool     `json:"has_variable_pricing"`
	PriceVariants          []PriceVariant `json:"price_variants,omitempty"`
	// Runtime metrics
	Status                 string   `json:"status"`
	AvgLatencyMs           int64    `json:"avg_latency_ms"`
	SuccessRate            float64  `json:"success_rate"`
	IsHot                  bool     `json:"is_hot"`
	RequestCount           int64    `json:"request_count"`
}

// PriceVariant 描述模型在特定请求参数下的输入价格档位。
type PriceVariant struct {
	Resolution string  `json:"resolution"`
	HasVideo   bool    `json:"has_video"`
	InputPrice float64 `json:"input_price"`
}

func parseCapabilities(capStr string) []string {
	if capStr == "" {
		return []string{}
	}
	var caps []string
	if err := common.UnmarshalJsonStr(capStr, &caps); err != nil {
		return []string{}
	}
	return caps
}

func overlaySpec(cm *CatalogModel, spec *sidecarModel.ModelSpec) {
	cm.ContextLength = spec.ContextLength
	cm.MaxOutputTokens = spec.MaxOutputTokens
	cm.Capabilities = parseCapabilities(spec.Capabilities)
	cm.ReleaseDate = spec.ReleaseDate
	cm.KnowledgeCutoff = spec.KnowledgeCutoff
	cm.ParameterCount = spec.ParameterCount
}

func minGroupRatio(enableGroups []string) float64 {
	if len(enableGroups) == 0 {
		return 1
	}
	groupRatios := ratio_setting.GetGroupRatioCopy()
	minRatio := math.MaxFloat64
	for _, g := range enableGroups {
		if r, ok := groupRatios[g]; ok && r < minRatio {
			minRatio = r
		}
	}
	if minRatio == math.MaxFloat64 {
		return 1
	}
	return minRatio
}

func buildCatalogModel(p model.Pricing, vendorMap map[int]model.PricingVendor) CatalogModel {
	ratio := minGroupRatio(p.EnableGroup)
	inputPrice := p.ModelRatio * 2 * ratio
	outputPrice := inputPrice * p.CompletionRatio

	cm := CatalogModel{
		ModelName:            p.ModelName,
		Description:          p.Description,
		VendorID:             p.VendorID,
		ModelRatio:           p.ModelRatio,
		CompletionRatio:      p.CompletionRatio,
		ModelPrice:           p.ModelPrice,
		CacheRatio:           p.CacheRatio,
		CreateCacheRatio:     p.CreateCacheRatio,
		ImageRatio:           p.ImageRatio,
		AudioRatio:           p.AudioRatio,
		AudioCompletionRatio: p.AudioCompletionRatio,
		EnableGroup:          p.EnableGroup,
		Tags:                 p.Tags,
		BillingMode:          p.BillingMode,
		BillingExpr:          p.BillingExpr,
		Icon:                 p.Icon,
		InputPrice:           inputPrice,
		OutputPrice:          outputPrice,
		HasVariablePricing:   doubao.HasVariablePricing(p.ModelName),
	}
	if variants := doubao.GetVideoPriceVariants(p.ModelName); len(variants) > 0 {
		cm.PriceVariants = make([]PriceVariant, len(variants))
		for i, v := range variants {
			cm.PriceVariants[i] = PriceVariant{
				Resolution: v.Resolution,
				HasVideo:   v.HasVideo,
				InputPrice: v.Price * ratio,
			}
		}
	}
	if v, ok := vendorMap[p.VendorID]; ok {
		cm.VendorName = v.Name
		cm.VendorIcon = v.Icon
	}
	for _, et := range p.SupportedEndpointTypes {
		cm.SupportedEndpointTypes = append(cm.SupportedEndpointTypes, string(et))
	}
	return cm
}

func statusFromSuccessRate(rate float64, requestCount int64) string {
	if requestCount == 0 {
		return "unknown"
	}
	if rate >= 95 {
		return "running"
	}
	if rate >= 80 {
		return "degraded"
	}
	return "down"
}

func injectRuntimeMetrics(models []CatalogModel) {
	summary, err := perfmetrics.QuerySummaryAll(24)
	if err != nil {
		common.SysLog("failed to query perf metrics summary: " + err.Error())
		for i := range models {
			models[i].Status = "unknown"
		}
		return
	}

	metricMap := make(map[string]perfmetrics.ModelSummary, len(summary.Models))
	for _, m := range summary.Models {
		metricMap[m.ModelName] = m
	}

	type candidate struct {
		idx   int
		count int64
	}
	var candidates []candidate
	for i := range models {
		if m, ok := metricMap[models[i].ModelName]; ok {
			models[i].AvgLatencyMs = m.AvgLatencyMs
			models[i].SuccessRate = m.SuccessRate
			models[i].RequestCount = m.RequestCount
			models[i].Status = statusFromSuccessRate(m.SuccessRate, m.RequestCount)
			if m.RequestCount > 0 {
				candidates = append(candidates, candidate{i, m.RequestCount})
			}
		} else {
			models[i].Status = "unknown"
		}
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].count > candidates[j].count
	})
	hotLimit := 10
	if len(candidates) < hotLimit {
		hotLimit = len(candidates)
	}
	for i := 0; i < hotLimit; i++ {
		models[candidates[i].idx].IsHot = true
	}
}

func injectRuntimeMetricSingle(cm *CatalogModel) {
	summary, err := perfmetrics.QuerySummaryAll(24)
	if err != nil {
		common.SysLog("failed to query perf metrics summary: " + err.Error())
		cm.Status = "unknown"
		return
	}
	for _, m := range summary.Models {
		if m.ModelName == cm.ModelName {
			cm.AvgLatencyMs = m.AvgLatencyMs
			cm.SuccessRate = m.SuccessRate
			cm.RequestCount = m.RequestCount
			cm.Status = statusFromSuccessRate(m.SuccessRate, m.RequestCount)
			return
		}
	}
	cm.Status = "unknown"
}

func GetModelCatalog(vendorFilter string, capabilitiesFilter string) ([]CatalogModel, error) {
	pricing := model.GetPricing()
	vendors := model.GetVendors()
	vendorMap := make(map[int]model.PricingVendor)
	for _, v := range vendors {
		vendorMap[v.ID] = v
	}

	var specs []*sidecarModel.ModelSpec
	err := sidecarModel.DB.Where("status = ?", 1).Find(&specs).Error
	if err != nil {
		return nil, err
	}
	specMap := make(map[string]*sidecarModel.ModelSpec)
	for _, s := range specs {
		specMap[s.ModelName] = s
	}

	// 预加载渠道类型 → 能力标签映射，用于自动注入
	autoCaps, err := buildModelAutoCapabilities()
	if err != nil {
		return nil, err
	}

	// 解析能力过滤条件
	var capFilters []string
	if capabilitiesFilter != "" {
		capFilters = strings.Split(capabilitiesFilter, ",")
		for i := range capFilters {
			capFilters[i] = strings.TrimSpace(capFilters[i])
		}
	}

	result := make([]CatalogModel, 0, len(pricing))
	for _, p := range pricing {
		cm := buildCatalogModel(p, vendorMap)
		if spec, ok := specMap[p.ModelName]; ok {
			overlaySpec(&cm, spec)
		}

		// 自动注入渠道派生的能力标签（video, image-generation, audio 等）
		// model_specs 手动配置的 capabilities 优先级更高，自动注入只做补充
		if autoCaps[p.ModelName] != nil {
			injectAutoCapabilities(&cm, autoCaps[p.ModelName])
		}

		// 按提供商过滤
		if vendorFilter != "" && cm.VendorName != vendorFilter {
			continue
		}

		// 按能力过滤（要求模型具备所有指定的能力）
		if len(capFilters) > 0 {
			capSet := make(map[string]struct{}, len(cm.Capabilities))
			for _, c := range cm.Capabilities {
				capSet[c] = struct{}{}
			}
			match := true
			for _, cf := range capFilters {
				if _, ok := capSet[cf]; !ok {
					match = false
					break
				}
			}
			if !match {
				continue
			}
		}

		result = append(result, cm)
	}
	injectRuntimeMetrics(result)
	return result, nil
}

func GetModelCatalogByName(modelName string) (*CatalogModel, error) {
	pricing := model.GetPricing()
	vendors := model.GetVendors()
	vendorMap := make(map[int]model.PricingVendor)
	for _, v := range vendors {
		vendorMap[v.ID] = v
	}

	var target *model.Pricing
	for i := range pricing {
		if pricing[i].ModelName == modelName {
			target = &pricing[i]
			break
		}
	}
	if target == nil {
		return nil, nil
	}

	cm := buildCatalogModel(*target, vendorMap)
	spec, err := sidecarModel.GetModelSpecByModelName(modelName)
	if err == nil && spec != nil {
		overlaySpec(&cm, spec)
	}

	// 自动注入渠道派生的能力标签
	autoCaps, err := buildModelAutoCapabilities()
	if err == nil && autoCaps[modelName] != nil {
		injectAutoCapabilities(&cm, autoCaps[modelName])
	}

	injectRuntimeMetricSingle(&cm)
	return &cm, nil
}

// channelCapabilityMap 渠道类型 → 能力标签映射。
// 用于根据渠道类型自动推断模型具备的能力。
var channelCapabilityMap = map[int]string{
	constant.ChannelTypeKling:       "video",
	constant.ChannelTypeJimeng:      "video",
	constant.ChannelTypeVidu:        "video",
	constant.ChannelTypeDoubaoVideo: "video",
	constant.ChannelTypeSora:        "video",
	constant.ChannelTypeReplicate:   "video",
	constant.ChannelTypeSunoAPI:     "audio",
}

// buildModelAutoCapabilities 查询所有启用的 abilities，根据渠道类型推断每个模型
// 自动具备的能力标签（如 video、audio 等）。
// 返回 map[modelName] → 该模型自动具备的能力标签集合。
func buildModelAutoCapabilities() (map[string][]string, error) {
	abilities, err := model.GetAllEnableAbilityWithChannels()
	if err != nil {
		return nil, err
	}
	result := make(map[string]map[string]bool)
	for _, a := range abilities {
		capTag, ok := channelCapabilityMap[a.ChannelType]
		if !ok {
			continue
		}
		if result[a.Model] == nil {
			result[a.Model] = make(map[string]bool)
		}
		result[a.Model][capTag] = true
	}
	// 转为切片
	out := make(map[string][]string, len(result))
	for model, capSet := range result {
		for c := range capSet {
			out[model] = append(out[model], c)
		}
	}
	return out, nil
}

// injectAutoCapabilities 将自动推断的能力标签注入到 CatalogModel 的 Capabilities 中。
// 只补充 model_specs 未包含的能力，避免覆盖手动配置。
func injectAutoCapabilities(cm *CatalogModel, autoCaps []string) {
	existing := make(map[string]bool, len(cm.Capabilities))
	for _, c := range cm.Capabilities {
		existing[c] = true
	}
	for _, c := range autoCaps {
		if !existing[c] {
			cm.Capabilities = append(cm.Capabilities, c)
		}
	}
}
