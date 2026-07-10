package providers

import (
	"github.com/Wei-Shaw/sub2api/internal/media"
)

// Registry 按模型名路由到火山方舟或 OpenRouter 适配器。
type Registry struct {
	volcengine *VolcengineProvider
	openrouter *OpenRouterProvider
}

// NewRegistry 根据配置构造厂商注册表（未配置 API Key 的后端不会被选中）。
func NewRegistry(cfg Config) *Registry {
	client := cfg.HTTPClient
	return &Registry{
		volcengine: NewVolcengineProvider(cfg.Volcengine, client),
		openrouter: NewOpenRouterProvider(cfg.OpenRouter, client),
	}
}

// ProviderFor 实现 media.ProviderRegistry。
func (r *Registry) ProviderFor(model string) (media.Provider, error) {
	if r == nil {
		return nil, media.ErrProviderNotFound
	}
	backend, ok := media.RouteModel(model)
	if !ok {
		return nil, media.ErrProviderNotFound
	}
	switch backend {
	case media.BackendVolcengine:
		if r.volcengine != nil && r.volcengine.cfg.enabled() {
			return r.volcengine, nil
		}
	case media.BackendOpenRouter:
		if r.openrouter != nil && r.openrouter.cfg.enabled() {
			return r.openrouter, nil
		}
	}
	return nil, media.ErrProviderNotFound
}
