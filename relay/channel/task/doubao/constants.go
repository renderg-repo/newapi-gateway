package doubao

import (
	"github.com/QuantumNous/new-api/setting/video_price_setting"
)

var ModelList = []string{
	"doubao-seedance-1-0-pro-250528",
	"doubao-seedance-1-0-lite-t2v",
	"doubao-seedance-1-0-lite-i2v",
	"doubao-seedance-1-5-pro-251215",
	"doubao-seedance-2-0-260128",
	"doubao-seedance-2-0-fast-260128",
}

var ChannelName = "doubao-video"

// HasVariablePricing 返回指定模型是否配置了按输出分辨率/视频输入变化的多维价格表。
func HasVariablePricing(modelName string) bool {
	return video_price_setting.HasVariablePricing(modelName)
}

// VideoPriceVariant 描述一个模型在特定 (分辨率, 是否含视频输入) 下的原始单价。
// Price 单位与 ModelRatio 一致：元/百万 token。
type VideoPriceVariant = video_price_setting.PriceVariant

// GetVideoPriceVariants 返回指定模型的所有多维定价档位，按分辨率/是否视频排序。
// 若模型无多维定价表，返回 nil。
func GetVideoPriceVariants(modelName string) []VideoPriceVariant {
	return video_price_setting.GetVideoPriceVariants(modelName)
}

// GetVideoInputRatio 返回指定模型在给定输出分辨率/是否含视频输入下，相对基准价的计费倍率。
// 第二个返回值表示该模型是否配置了价格表；倍率为 1.0 时调用方可忽略该 OtherRatio。
func GetVideoInputRatio(modelName, resolution string, hasVideo bool) (float64, bool) {
	return video_price_setting.GetVideoInputRatio(modelName, resolution, hasVideo)
}