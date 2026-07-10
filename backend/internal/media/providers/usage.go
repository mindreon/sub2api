package providers

import (
	"encoding/json"
	"math"

	"github.com/Wei-Shaw/sub2api/internal/media"
)

// openRouterVideoTokenUnitUSD 与 media.openRouterVideoTokenUSDPerM 对齐（$/token）。
const openRouterVideoTokenUnitUSD = 7.0 / 1_000_000

// usageFromVolcengineResponse 从火山方舟任务查询响应提取计费用量。
func usageFromVolcengineResponse(raw map[string]any, task *media.Task) media.BillingUsage {
	usage := baseUsageFromTask(task)
	if raw == nil {
		return usage
	}
	if u, ok := raw["usage"].(map[string]any); ok {
		if tokens, ok := asInt64(u["completion_tokens"]); ok && tokens > 0 {
			usage.VideoTokens = tokens
		}
	}
	return usage
}

// usageFromOpenRouterResponse 从 OpenRouter 任务查询响应提取计费用量。
func usageFromOpenRouterResponse(raw map[string]any, task *media.Task) media.BillingUsage {
	usage := baseUsageFromTask(task)
	if raw == nil {
		return usage
	}
	if u, ok := raw["usage"].(map[string]any); ok {
		if tokens, ok := asInt64(u["video_tokens"]); ok && tokens > 0 {
			usage.VideoTokens = tokens
		} else if tokens, ok := asInt64(u["completion_tokens"]); ok && tokens > 0 {
			usage.VideoTokens = tokens
		} else if cost, ok := asFloat64(u["cost"]); ok && cost > 0 && openRouterVideoTokenUnitUSD > 0 {
			usage.VideoTokens = int64(math.Round(cost / openRouterVideoTokenUnitUSD))
		}
	}
	return usage
}

func baseUsageFromTask(task *media.Task) media.BillingUsage {
	if task == nil {
		return media.BillingUsage{}
	}
	return usageFromMap(task.RequestParams)
}

func usageFromMap(params map[string]any) media.BillingUsage {
	if len(params) == 0 {
		return media.BillingUsage{}
	}
	u := media.BillingUsage{}
	if v, ok := asString(params["resolution"]); ok {
		u.Resolution = v
	}
	if v, ok := asString(params["aspect_ratio"]); ok {
		u.AspectRatio = v
	}
	if v, ok := asString(params["ratio"]); ok && u.AspectRatio == "" {
		u.AspectRatio = v
	}
	if v, ok := asFloat64(params["video_output_seconds"]); ok {
		u.VideoOutputSeconds = v
	}
	if v, ok := asFloat64(params["duration"]); ok && u.VideoOutputSeconds == 0 {
		u.VideoOutputSeconds = v
	}
	if v, ok := asInt(params["video_width"]); ok {
		u.VideoWidth = v
	}
	if v, ok := asInt(params["video_height"]); ok {
		u.VideoHeight = v
	}
	if v, ok := asInt(params["video_fps"]); ok {
		u.VideoFPS = v
	}
	if v, ok := asBool(params["has_video_input"]); ok {
		u.HasVideoInput = v
	}
	if v, ok := asBool(params["has_audio"]); ok {
		u.HasAudio = v
	}
	if v, ok := asBool(params["generate_audio"]); ok {
		u.HasAudio = v
	}
	return u
}

// resultURLFromResponse 从任务查询响应中尽力提取生成视频地址。
//
// 兼容多种上游结构：content.video_url（火山方舟）、顶层 video_url/url、
// output.video_url、data[0].url/video_url。找不到返回空串（非致命）。
func resultURLFromResponse(raw map[string]any) string {
	if raw == nil {
		return ""
	}
	if content := rawMap(raw["content"]); content != nil {
		if u, ok := asString(content["video_url"]); ok {
			return u
		}
		if u, ok := asString(content["url"]); ok {
			return u
		}
	}
	if output := rawMap(raw["output"]); output != nil {
		if u, ok := asString(output["video_url"]); ok {
			return u
		}
	}
	if u, ok := asString(raw["video_url"]); ok {
		return u
	}
	if u, ok := asString(raw["url"]); ok {
		return u
	}
	if list, ok := raw["data"].([]any); ok && len(list) > 0 {
		if first := rawMap(list[0]); first != nil {
			if u, ok := asString(first["url"]); ok {
				return u
			}
			if u, ok := asString(first["video_url"]); ok {
				return u
			}
		}
	}
	return ""
}

func rawMap(v any) map[string]any {
	if m, ok := v.(map[string]any); ok {
		return m
	}
	return nil
}

func asString(v any) (string, bool) {
	switch t := v.(type) {
	case string:
		return t, t != ""
	case json.Number:
		return t.String(), true
	default:
		return "", false
	}
}

func asFloat64(v any) (float64, bool) {
	switch t := v.(type) {
	case float64:
		return t, true
	case float32:
		return float64(t), true
	case int:
		return float64(t), true
	case int64:
		return float64(t), true
	case json.Number:
		f, err := t.Float64()
		return f, err == nil
	default:
		return 0, false
	}
}

func asInt(v any) (int, bool) {
	switch t := v.(type) {
	case int:
		return t, true
	case int64:
		return int(t), true
	case float64:
		return int(t), true
	case json.Number:
		i, err := t.Int64()
		return int(i), err == nil
	default:
		return 0, false
	}
}

func asInt64(v any) (int64, bool) {
	switch t := v.(type) {
	case int64:
		return t, true
	case int:
		return int64(t), true
	case float64:
		return int64(t), true
	case json.Number:
		i, err := t.Int64()
		return i, err == nil
	default:
		return 0, false
	}
}

func asBool(v any) (bool, bool) {
	switch t := v.(type) {
	case bool:
		return t, true
	default:
		return false, false
	}
}
