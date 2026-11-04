package relay

import (
	"testing"

	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/relay/channel/task/taskcommon"
)

func TestPublicTaskResultURL(t *testing.T) {
	const upstream = "https://cdn.example.com/video.mp4"
	const doubaoPlatform = constant.TaskPlatform("54")
	const klingPlatform = constant.TaskPlatform("50")

	tests := []struct {
		name string
		task *model.Task
		want string
	}{
		{
			name: "doubao direct keeps gateway content endpoint",
			task: &model.Task{
				TaskID:   "task_direct",
				Platform: doubaoPlatform,
				Status:   model.TaskStatusSuccess,
				PrivateData: model.TaskPrivateData{
					ResultURL:      upstream,
					VideoFetchMode: model.VideoFetchModeDirectURL,
				},
			},
			want: taskcommon.BuildProxyURL("task_direct"),
		},
		{
			name: "doubao mobile cloud keeps gateway content endpoint",
			task: &model.Task{
				TaskID:   "task_aicc",
				Platform: doubaoPlatform,
				Status:   model.TaskStatusSuccess,
				PrivateData: model.TaskPrivateData{
					VideoFetchMode: model.VideoFetchModeChannelDownload,
				},
			},
			want: taskcommon.BuildProxyURL("task_aicc"),
		},
		{
			name: "provider modes keep gateway content endpoint",
			task: &model.Task{
				TaskID:   "task_provider",
				Platform: constant.TaskPlatform("55"),
				Status:   model.TaskStatusSuccess,
				PrivateData: model.TaskPrivateData{
					ResultURL:      upstream,
					VideoFetchMode: model.VideoFetchModeProviderContent,
				},
			},
			want: taskcommon.BuildProxyURL("task_provider"),
		},
		{
			name: "other direct channel keeps upstream url",
			task: &model.Task{
				TaskID:   "task_kling",
				Platform: klingPlatform,
				Status:   model.TaskStatusSuccess,
				PrivateData: model.TaskPrivateData{
					ResultURL:      upstream,
					VideoFetchMode: model.VideoFetchModeDirectURL,
				},
			},
			want: upstream,
		},
		{
			name: "legacy doubao task still exposes gateway endpoint",
			task: &model.Task{
				TaskID:   "task_legacy_doubao",
				Platform: doubaoPlatform,
				Status:   model.TaskStatusSuccess,
				PrivateData: model.TaskPrivateData{
					ResultURL: upstream,
				},
			},
			want: taskcommon.BuildProxyURL("task_legacy_doubao"),
		},
		{
			name: "suno task is not shaped as video content",
			task: &model.Task{
				TaskID:   "task_suno",
				Platform: constant.TaskPlatformSuno,
				Status:   model.TaskStatusSuccess,
				PrivateData: model.TaskPrivateData{
					ResultURL: upstream,
				},
			},
			want: upstream,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := publicTaskResultURL(tt.task); got != tt.want {
				t.Fatalf("publicTaskResultURL() = %q, want %q", got, tt.want)
			}
		})
	}
}
