package taskcommon

import (
	"strings"

	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/model"
)

// IsMobileCloudBaseURL reports whether baseURL points at the mobile-cloud
// encrypted video proxy (the well-known /aicc/ prefix).
func IsMobileCloudBaseURL(baseURL string) bool {
	return strings.Contains(baseURL, "/aicc/")
}

// ResolveVideoFetchMode is the single place that maps a channel's type and
// base URL to a VideoFetchMode. Polling persists the result; VideoProxy falls
// back to this helper for tasks created before video_fetch_mode existed.
func ResolveVideoFetchMode(ch *model.Channel) model.VideoFetchMode {
	if ch == nil {
		return model.VideoFetchModeDirectURL
	}
	switch ch.Type {
	case constant.ChannelTypeDoubaoVideo:
		if IsMobileCloudBaseURL(ch.GetBaseURL()) {
			return model.VideoFetchModeChannelDownload
		}
		return model.VideoFetchModeDirectURL
	case constant.ChannelTypeOpenAI, constant.ChannelTypeSora:
		return model.VideoFetchModeProviderContent
	case constant.ChannelTypeGemini, constant.ChannelTypeVertexAi:
		return model.VideoFetchModeProviderResolve
	default:
		return model.VideoFetchModeDirectURL
	}
}
