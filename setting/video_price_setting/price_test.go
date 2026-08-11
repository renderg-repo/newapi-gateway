package video_price_setting

import "testing"

// withTable 临时替换全局价格表，测试结束后恢复。
func withTable(t *testing.T, table string) {
	t.Helper()
	orig := videoPriceSetting.Table
	videoPriceSetting.Table = table
	t.Cleanup(func() { videoPriceSetting.Table = orig })
}

func TestGetVideoInputRatio(t *testing.T) {
	// 含组合档与 snake_case 键的标准配置
	withTable(t, `{"m": [
		{"resolution": "480p/720p", "has_video": false, "price": 92},
		{"resolution": "480p/720p", "has_video": true,  "price": 56},
		{"resolution": "1080p",     "has_video": false, "price": 102},
		{"resolution": "1080p",     "has_video": true,  "price": 62}
	]}`)

	cases := []struct {
		name     string
		res      string
		hasVideo bool
		want     float64
	}{
		{"基准档 720p 无视频", "720p", false, 1.0},
		{"组合档命中 480p", "480p", false, 1.0},
		{"组合档含视频 720p", "720p", true, 56.0 / 92.0},
		{"组合档含视频 480p", "480p", true, 56.0 / 92.0},
		{"1080p 无视频", "1080p", false, 102.0 / 92.0},
		{"1080p 含视频", "1080p", true, 62.0 / 92.0},
		{"大小写不敏感", "1080P", false, 102.0 / 92.0},
		{"空白容忍", " 720p ", false, 1.0},
		{"未知分辨率回退基准", "4k", false, 1.0},
		{"空分辨率回退基准", "", false, 1.0},
		{"未知分辨率含视频回退基准", "4k", true, 1.0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := GetVideoInputRatio("m", tc.res, tc.hasVideo)
			if !ok {
				t.Fatalf("expected ok=true")
			}
			if diff := got - tc.want; diff > 1e-9 || diff < -1e-9 {
				t.Fatalf("res=%q hasVideo=%v: got %v, want %v", tc.res, tc.hasVideo, got, tc.want)
			}
		})
	}
}

func TestGetVideoInputRatioLegacyCamelCase(t *testing.T) {
	// 历史配置使用 camelCase "hasVideo"，解析必须兼容，否则含视频档位静默丢失
	withTable(t, `{"m": [
		{"resolution": "720p", "hasVideo": false, "price": 92},
		{"resolution": "720p", "hasVideo": true,  "price": 56}
	]}`)

	got, ok := GetVideoInputRatio("m", "720p", true)
	if !ok {
		t.Fatalf("expected ok=true")
	}
	if want := 56.0 / 92.0; got != want {
		t.Fatalf("camelCase hasVideo not honored: got %v, want %v", got, want)
	}
}

func TestGetVideoInputRatioUnknownModel(t *testing.T) {
	withTable(t, `{"m": [{"resolution": "720p", "has_video": false, "price": 92}]}`)
	if _, ok := GetVideoInputRatio("other", "720p", false); ok {
		t.Fatalf("expected ok=false for unknown model")
	}
	if HasVariablePricing("other") {
		t.Fatalf("expected HasVariablePricing=false for unknown model")
	}
}
