package video_price_setting

import (
	"encoding/json"
	"sort"
	"strings"

	"github.com/QuantumNous/new-api/setting/config"
)

// PriceVariant 一个模型在特定请求参数下的价格档位。
type PriceVariant struct {
	Resolution string  `json:"resolution"`
	HasVideo   bool    `json:"has_video"`
	Price      float64 `json:"price"`
}

// VideoPriceSetting 是 config.GlobalConfig.Register 管理的配置模块。
// DB key: video_price_table.table
type VideoPriceSetting struct {
	Table string `json:"table"`
}

var videoPriceSetting = VideoPriceSetting{
	Table: defaultTableJSON(),
}

func init() {
	config.GlobalConfig.Register("video_price_table", &videoPriceSetting)
}

// ---------------------------------------------------------------------------
// 默认价格表（与火山引擎官方定价一致，元/百万 token）
// ---------------------------------------------------------------------------

func defaultTableJSON() string {
	table := map[string][]PriceVariant{
		"doubao-seedance-2-0-260128": {
			{Resolution: "480p/720p", HasVideo: false, Price: 46},
			{Resolution: "480p/720p", HasVideo: true,  Price: 28},
			{Resolution: "1080p",     HasVideo: false, Price: 51},
			{Resolution: "1080p",     HasVideo: true,  Price: 31},
			{Resolution: "4k",        HasVideo: false, Price: 26},
			{Resolution: "4k",        HasVideo: true,  Price: 16},
		},
		"doubao-seedance-2-0-fast-260128": {
			{Resolution: "480p/720p", HasVideo: false, Price: 37},
			{Resolution: "480p/720p", HasVideo: true,  Price: 22},
		},
	}
	b, _ := json.Marshal(table)
	return string(b)
}

// ---------------------------------------------------------------------------
// 内部解析
// ---------------------------------------------------------------------------

func parseTable() map[string][]PriceVariant {
	var table map[string][]PriceVariant
	if err := json.Unmarshal([]byte(videoPriceSetting.Table), &table); err != nil {
		return nil
	}
	return table
}

// ---------------------------------------------------------------------------
// 热读访问器（hot path）
// ---------------------------------------------------------------------------

// HasVariablePricing 返回指定模型是否配置了按请求参数变化的多维定价表。
func HasVariablePricing(modelName string) bool {
	table := parseTable()
	_, ok := table[modelName]
	return ok
}

// GetVideoPriceVariants 返回指定模型的所有多维定价档位，按分辨率/是否含视频排序。
func GetVideoPriceVariants(modelName string) []PriceVariant {
	table := parseTable()
	variants, ok := table[modelName]
	if !ok {
		return nil
	}
	sorted := make([]PriceVariant, len(variants))
	copy(sorted, variants)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Resolution != sorted[j].Resolution {
			return sorted[i].Resolution < sorted[j].Resolution
		}
		return !sorted[i].HasVideo && sorted[j].HasVideo
	})
	return sorted
}

// GetVideoInputRatio 返回指定模型在给定输出分辨率/是否含视频输入下，相对基准价的计费倍率。
// 第二个返回值表示该模型是否配置了价格表；倍率为 1.0 时调用方可忽略该 OtherRatio。
//
// 分辨率匹配规则：
//   - 先做精确匹配（忽略大小写、"1080P" 与 "1080p" 等价权）。
//   - 请求分辨率与配置档位均可携带任意标签（如 "1080p"、"720p"、"480p"、"2k"），
//     若请求分辨率没有任何精确匹配，则回退到该模型的不含视频基准档（倍率 1.0）。
func GetVideoInputRatio(modelName, resolution string, hasVideo bool) (float64, bool) {
	table := parseTable()
	variants, ok := table[modelName]
	if !ok {
		return 0, false
	}

	// 基准价：不含视频、价格最低的档位（作为该模型的 1.0 基准）。
	// 若请求分辨率匹配不到任何档位，也按此基准计费。
	var basePrice float64
	for _, v := range variants {
		if !v.HasVideo && (basePrice <= 0 || v.Price < basePrice) {
			basePrice = v.Price
		}
	}
	if basePrice <= 0 {
		return 0, false
	}

	// 精确匹配目标档位（分辨率忽略大小写，hasVideo 精确匹配）
	normResolution := strings.ToLower(strings.TrimSpace(resolution))
	for _, v := range variants {
		if strings.EqualFold(v.Resolution, normResolution) && v.HasVideo == hasVideo {
			return v.Price / basePrice, true
		}
	}

	// 未配置的组合按基准价计费
	return 1.0, true
}