package providers

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/media"
)

// OpenRouterProvider 对接 OpenRouter 异步视频 API（POST/GET /api/v1/videos）。
type OpenRouterProvider struct {
	cfg    OpenRouterConfig
	client httpDoer
}

// NewOpenRouterProvider 构造 OpenRouter 适配器。
func NewOpenRouterProvider(cfg OpenRouterConfig, client httpDoer) *OpenRouterProvider {
	return &OpenRouterProvider{cfg: cfg, client: client}
}

func (p *OpenRouterProvider) Submit(ctx context.Context, task *media.Task) (string, error) {
	if p == nil || !p.cfg.enabled() {
		return "", media.ErrUpstreamNotWired
	}
	body := buildOpenRouterSubmitBody(task)
	var resp map[string]any
	url := p.cfg.baseURL() + "/api/v1/videos"
	if err := doJSON(ctx, p.client, http.MethodPost, url, p.cfg.APIKey, body, &resp); err != nil {
		return "", err
	}
	id, _ := asString(resp["id"])
	if id == "" {
		return "", fmt.Errorf("openrouter: missing job id in response")
	}
	return id, nil
}

func (p *OpenRouterProvider) QueryStatus(ctx context.Context, task *media.Task) (*media.ProviderStatus, error) {
	if p == nil || !p.cfg.enabled() {
		return nil, media.ErrUpstreamNotWired
	}
	if task == nil || strings.TrimSpace(task.UpstreamTaskID) == "" {
		return nil, fmt.Errorf("openrouter: missing upstream task id")
	}
	var resp map[string]any
	url := p.cfg.baseURL() + "/api/v1/videos/" + task.UpstreamTaskID
	if err := doJSON(ctx, p.client, http.MethodGet, url, p.cfg.APIKey, nil, &resp); err != nil {
		return nil, err
	}
	return mapOpenRouterStatus(resp, task), nil
}

func buildOpenRouterSubmitBody(task *media.Task) map[string]any {
	body := cloneMap(task.RequestParams)
	if body == nil {
		body = make(map[string]any)
	}
	body["model"] = task.Model
	return body
}

func mapOpenRouterStatus(raw map[string]any, task *media.Task) *media.ProviderStatus {
	status, _ := asString(raw["status"])
	out := &media.ProviderStatus{RawUsage: raw}
	switch strings.ToLower(status) {
	case "completed", "succeeded", "success":
		out.State = media.ProviderSucceeded
		out.Usage = usageFromOpenRouterResponse(raw, task)
		out.ResultURL = resultURLFromResponse(raw)
	case "failed", "cancelled", "canceled", "expired":
		out.State = media.ProviderFailed
		if msg, ok := asString(raw["error"]); ok {
			out.ErrorMessage = msg
		}
		if out.ErrorMessage == "" {
			out.ErrorMessage = "openrouter task " + status
		}
	default:
		out.State = media.ProviderInProgress
	}
	return out
}
