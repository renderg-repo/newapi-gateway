package model

// VideoFetchMode describes how VideoProxy obtains a video task's bytes.
// It is persisted in TaskPrivateData so polling and proxying agree on the
// fetch strategy without inferring it from the stored result URL.
type VideoFetchMode string

const (
	// VideoFetchModeDirectURL means PrivateData.ResultURL holds a real
	// upstream/CDN URL that VideoProxy can HTTP GET.
	VideoFetchModeDirectURL VideoFetchMode = "direct_url"
	// VideoFetchModeChannelDownload means VideoProxy must call the channel's
	// /download endpoint (mobile-cloud /aicc/ encrypted proxy) to obtain a
	// decrypted MP4 stream. ResultURL is intentionally empty.
	VideoFetchModeChannelDownload VideoFetchMode = "channel_download"
	// VideoFetchModeProviderContent means VideoProxy must call the provider's
	// /v1/videos/{upstream_task_id}/content endpoint (OpenAI / Sora).
	VideoFetchModeProviderContent VideoFetchMode = "provider_content"
	// VideoFetchModeProviderResolve means the video URL must be resolved from
	// the provider's task payload (Gemini / Vertex).
	VideoFetchModeProviderResolve VideoFetchMode = "provider_resolve"
)
