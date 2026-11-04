package taskcommon

import (
	"testing"

	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/model"
)

func stringPtr(v string) *string {
	return &v
}

func TestResolveVideoFetchMode(t *testing.T) {
	tests := []struct {
		name    string
		channel *model.Channel
		want    model.VideoFetchMode
	}{
		{
			name:    "nil channel falls back to direct",
			channel: nil,
			want:    model.VideoFetchModeDirectURL,
		},
		{
			name: "doubao direct",
			channel: &model.Channel{
				Type:    constant.ChannelTypeDoubaoVideo,
				BaseURL: stringPtr("https://ark.cn-beijing.volces.com/api/v3"),
			},
			want: model.VideoFetchModeDirectURL,
		},
		{
			name: "doubao mobile cloud",
			channel: &model.Channel{
				Type:    constant.ChannelTypeDoubaoVideo,
				BaseURL: stringPtr("https://proxy.example.com/aicc/seedance/api/v3"),
			},
			want: model.VideoFetchModeChannelDownload,
		},
		{
			name:    "openai",
			channel: &model.Channel{Type: constant.ChannelTypeOpenAI},
			want:    model.VideoFetchModeProviderContent,
		},
		{
			name:    "sora",
			channel: &model.Channel{Type: constant.ChannelTypeSora},
			want:    model.VideoFetchModeProviderContent,
		},
		{
			name:    "gemini",
			channel: &model.Channel{Type: constant.ChannelTypeGemini},
			want:    model.VideoFetchModeProviderResolve,
		},
		{
			name:    "vertex",
			channel: &model.Channel{Type: constant.ChannelTypeVertexAi},
			want:    model.VideoFetchModeProviderResolve,
		},
		{
			name:    "other task channel defaults to direct",
			channel: &model.Channel{Type: constant.ChannelTypeKling},
			want:    model.VideoFetchModeDirectURL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ResolveVideoFetchMode(tt.channel); got != tt.want {
				t.Fatalf("ResolveVideoFetchMode() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIsMobileCloudBaseURL(t *testing.T) {
	tests := []struct {
		baseURL string
		want    bool
	}{
		{baseURL: "https://proxy.example.com/aicc/seedance", want: true},
		{baseURL: "https://ark.cn-beijing.volces.com/api/v3", want: false},
		{baseURL: "", want: false},
		{baseURL: "https://example.com/api/v3?next=/aicc/", want: true},
	}

	for _, tt := range tests {
		if got := IsMobileCloudBaseURL(tt.baseURL); got != tt.want {
			t.Fatalf("IsMobileCloudBaseURL(%q) = %v, want %v", tt.baseURL, got, tt.want)
		}
	}
}
