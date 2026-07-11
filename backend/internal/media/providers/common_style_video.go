package providers

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/media"
)

// CommonStyleVideoProvider 对接 /v1/video/generations 风格的异步视频上游。
type CommonStyleVideoProvider struct {
	cfg    CommonStyleVideoConfig
	client httpDoer
}

func NewCommonStyleVideoProvider(cfg CommonStyleVideoConfig, client httpDoer) *CommonStyleVideoProvider {
	return &CommonStyleVideoProvider{cfg: cfg, client: client}
}

func (p *CommonStyleVideoProvider) Submit(ctx context.Context, task *media.Task) (string, error) {
	if p == nil || !p.cfg.enabled() {
		return "", media.ErrUpstreamNotWired
	}
	body := buildCommonStyleVideoSubmitBody(task)
	var resp map[string]any
	if err := doJSON(ctx, p.client, http.MethodPost, p.cfg.baseURL()+"/v1/video/generations", p.cfg.APIKey, body, &resp); err != nil {
		return "", err
	}
	id, _ := asString(resp["task_id"])
	if id == "" {
		id, _ = asString(resp["id"])
	}
	if id == "" {
		return "", fmt.Errorf("common style video: missing task id in response")
	}
	return id, nil
}

func (p *CommonStyleVideoProvider) QueryStatus(ctx context.Context, task *media.Task) (*media.ProviderStatus, error) {
	if p == nil || !p.cfg.enabled() {
		return nil, media.ErrUpstreamNotWired
	}
	if task == nil || strings.TrimSpace(task.UpstreamTaskID) == "" {
		return nil, fmt.Errorf("common style video: missing upstream task id")
	}
	var resp map[string]any
	url := p.cfg.baseURL() + "/v1/video/generations/" + task.UpstreamTaskID
	if err := doJSON(ctx, p.client, http.MethodGet, url, p.cfg.APIKey, nil, &resp); err != nil {
		return nil, err
	}
	return mapCommonStyleVideoStatus(resp, task), nil
}

func buildCommonStyleVideoSubmitBody(task *media.Task) map[string]any {
	metadata := cloneMap(task.RequestParams)
	prompt, _ := asString(metadata["prompt"])
	delete(metadata, "prompt")
	delete(metadata, "has_video_input")
	delete(metadata, "has_audio")
	return map[string]any{
		"model":    task.Model,
		"prompt":   prompt,
		"metadata": metadata,
	}
}

func mapCommonStyleVideoStatus(raw map[string]any, task *media.Task) *media.ProviderStatus {
	payload := rawMap(raw["data"])
	if payload == nil {
		payload = raw
	}
	status, _ := asString(payload["status"])
	out := &media.ProviderStatus{
		Usage:    usageFromCommonStyleVideoResponse(payload, task),
		RawUsage: raw,
	}
	switch strings.ToLower(status) {
	case "success", "succeeded", "completed":
		out.State = media.ProviderSucceeded
		out.ResultURL = resultURLFromCommonStyleVideoResponse(payload)
	case "failed", "cancelled", "canceled", "expired":
		out.State = media.ProviderFailed
		out.ErrorMessage, _ = asString(payload["fail_reason"])
		if out.ErrorMessage == "" {
			out.ErrorMessage, _ = asString(payload["error_message"])
		}
		if out.ErrorMessage == "" {
			out.ErrorMessage = "common style video task " + status
		}
	default:
		out.State = media.ProviderInProgress
	}
	return out
}

func usageFromCommonStyleVideoResponse(payload map[string]any, task *media.Task) media.BillingUsage {
	usage := baseUsageFromTask(task)
	usageMap := rawMap(payload["usage"])
	if nested := rawMap(payload["data"]); usageMap == nil && nested != nil {
		usageMap = rawMap(nested["usage"])
	}
	if tokens, ok := asInt64(usageMap["completion_tokens"]); ok && tokens > 0 {
		usage.VideoTokens = tokens
	}
	return usage
}

func resultURLFromCommonStyleVideoResponse(payload map[string]any) string {
	if result, ok := asString(payload["result_url"]); ok {
		return result
	}
	if result := resultURLFromResponse(payload); result != "" {
		return result
	}
	return resultURLFromResponse(rawMap(payload["data"]))
}
