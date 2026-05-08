package service

import (
	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	sidecarModel "github.com/QuantumNous/new-api/sidecar/model"
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
	if spec.Description != "" {
		cm.Description = spec.Description
	}
	if spec.Icon != "" {
		cm.Icon = spec.Icon
	}
	cm.ReleaseDate = spec.ReleaseDate
	cm.KnowledgeCutoff = spec.KnowledgeCutoff
	cm.ParameterCount = spec.ParameterCount
}

func buildCatalogModel(p model.Pricing, vendorMap map[int]model.PricingVendor) CatalogModel {
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

func GetModelCatalog() ([]CatalogModel, error) {
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

	result := make([]CatalogModel, 0, len(pricing))
	for _, p := range pricing {
		cm := buildCatalogModel(p, vendorMap)
		if spec, ok := specMap[p.ModelName]; ok {
			overlaySpec(&cm, spec)
		}
		result = append(result, cm)
	}
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
	return &cm, nil
}
