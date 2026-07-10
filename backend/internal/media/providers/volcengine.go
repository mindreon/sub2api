package providers

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/media"
)

// VolcengineProvider 对接火山方舟 Seedance 原生异步视频 API。
//
// 创建：POST {base}/api/v3/contents/generations/tasks
// 查询：GET  {base}/api/v3/contents/generations/tasks/{id}
type VolcengineProvider struct {
	cfg    VolcengineConfig
	client httpDoer
}

// NewVolcengineProvider 构造火山方舟适配器。
func NewVolcengineProvider(cfg VolcengineConfig, client httpDoer) *VolcengineProvider {
	return &VolcengineProvider{cfg: cfg, client: client}
}

func (p *VolcengineProvider) Submit(ctx context.Context, task *media.Task) (string, error) {
	if p == nil || !p.cfg.enabled() {
		return "", media.ErrUpstreamNotWired
	}
	body := buildVolcengineSubmitBody(task)
	var resp map[string]any
	url := p.cfg.baseURL() + "/api/v3/contents/generations/tasks"
	if err := doJSON(ctx, p.client, http.MethodPost, url, p.cfg.APIKey, body, &resp); err != nil {
		return "", err
	}
	id, _ := asString(resp["id"])
	if id == "" {
		return "", fmt.Errorf("volcengine: missing task id in response")
	}
	return id, nil
}

func (p *VolcengineProvider) QueryStatus(ctx context.Context, task *media.Task) (*media.ProviderStatus, error) {
	if p == nil || !p.cfg.enabled() {
		return nil, media.ErrUpstreamNotWired
	}
	if task == nil || strings.TrimSpace(task.UpstreamTaskID) == "" {
		return nil, fmt.Errorf("volcengine: missing upstream task id")
	}
	var resp map[string]any
	url := p.cfg.baseURL() + "/api/v3/contents/generations/tasks/" + task.UpstreamTaskID
	if err := doJSON(ctx, p.client, http.MethodGet, url, p.cfg.APIKey, nil, &resp); err != nil {
		return nil, err
	}
	return mapVolcengineStatus(resp, task), nil
}

func buildVolcengineSubmitBody(task *media.Task) map[string]any {
	body := cloneMap(task.RequestParams)
	if body == nil {
		body = make(map[string]any)
	}
	body["model"] = task.Model

	// 兼容 OpenAI 风格 prompt 字段：转成火山 content 数组。
	if prompt, ok := asString(body["prompt"]); ok {
		delete(body, "prompt")
		text := map[string]any{"type": "text", "text": prompt}
		if content, ok := contentItems(body["content"]); ok {
			body["content"] = append([]any{text}, content...)
		} else {
			body["content"] = []any{text}
		}
	}
	return body
}

func contentItems(v any) ([]any, bool) {
	switch items := v.(type) {
	case []any:
		return items, true
	case []map[string]any:
		out := make([]any, 0, len(items))
		for _, item := range items {
			out = append(out, item)
		}
		return out, true
	default:
		return nil, false
	}
}

func mapVolcengineStatus(raw map[string]any, task *media.Task) *media.ProviderStatus {
	status, _ := asString(raw["status"])
	out := &media.ProviderStatus{RawUsage: raw}
	switch strings.ToLower(status) {
	case "succeeded", "success", "completed":
		out.State = media.ProviderSucceeded
		out.Usage = usageFromVolcengineResponse(raw, task)
		out.ResultURL = resultURLFromResponse(raw)
	case "failed", "cancelled", "canceled", "expired":
		out.State = media.ProviderFailed
		if msg, ok := asString(raw["error"]); ok {
			out.ErrorMessage = msg
		} else if errObj := rawMap(raw["error"]); errObj != nil {
			if msg, ok := asString(errObj["message"]); ok {
				out.ErrorMessage = msg
			}
		}
		if out.ErrorMessage == "" {
			out.ErrorMessage = "volcengine task " + status
		}
	default:
		out.State = media.ProviderInProgress
	}
	return out
}

func cloneMap(in map[string]any) map[string]any {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
