package handler

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/media"
	"github.com/Wei-Shaw/sub2api/internal/media/providers"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

// accountMediaProviderFactory 根据账号凭证构造火山/OpenRouter 适配器。
type accountMediaProviderFactory struct{}

// NewAccountMediaProviderFactory 构造 media.ProviderFactory。
func NewAccountMediaProviderFactory() media.ProviderFactory {
	return &accountMediaProviderFactory{}
}

func (f *accountMediaProviderFactory) NewProvider(sel media.AccountSelection, model string) (media.Provider, error) {
	if strings.TrimSpace(sel.APIKey) == "" {
		return nil, fmt.Errorf("media: missing api_key on account %d", sel.AccountID)
	}
	client := &http.Client{Timeout: 120 * time.Second}
	switch sel.Platform {
	case service.PlatformVolcengine:
		if useCommonStyleVideoAPI(sel) {
			return providers.NewCommonStyleVideoProvider(providers.CommonStyleVideoConfig{
				APIKey:  sel.APIKey,
				BaseURL: sel.BaseURL,
			}, client), nil
		}
		return providers.NewVolcengineProvider(providers.VolcengineConfig{
			APIKey:  sel.APIKey,
			BaseURL: sel.BaseURL,
		}, client), nil
	case service.PlatformOpenRouter:
		return providers.NewOpenRouterProvider(providers.OpenRouterConfig{
			APIKey:  sel.APIKey,
			BaseURL: sel.BaseURL,
		}, client), nil
	default:
		backend, ok := media.RouteModel(model)
		if !ok {
			return nil, media.ErrProviderNotFound
		}
		return nil, fmt.Errorf("media: account platform %q does not match model backend %q", sel.Platform, backend)
	}
}

func useCommonStyleVideoAPI(sel media.AccountSelection) bool {
	switch strings.ToLower(strings.TrimSpace(sel.APIStyle)) {
	case "common", "common_style":
		return true
	case "native", "volcengine":
		return false
	}
	rawBaseURL := strings.TrimSpace(sel.BaseURL)
	if rawBaseURL == "" {
		return false
	}
	parsed, err := url.Parse(rawBaseURL)
	if err != nil {
		return true
	}
	host := strings.ToLower(parsed.Hostname())
	return host == "" || !isNativeVolcengineHost(host)
}

func isNativeVolcengineHost(host string) bool {
	return host == "volces.com" || strings.HasSuffix(host, ".volces.com") ||
		host == "bytepluses.com" || strings.HasSuffix(host, ".bytepluses.com") ||
		host == "volcengineapi.com" || strings.HasSuffix(host, ".volcengineapi.com")
}

// envFallbackMediaProviderFactory 在账号凭证缺失时回退环境变量（过渡期兼容）。
type envFallbackMediaProviderFactory struct {
	inner media.ProviderFactory
	env   providers.Config
}

// NewEnvFallbackMediaProviderFactory 包装主工厂，env 未配置时不回退。
func NewEnvFallbackMediaProviderFactory(inner media.ProviderFactory) media.ProviderFactory {
	return &envFallbackMediaProviderFactory{
		inner: inner,
		env:   providers.ConfigFromEnv(),
	}
}

func (f *envFallbackMediaProviderFactory) NewProvider(sel media.AccountSelection, model string) (media.Provider, error) {
	if f == nil || f.inner == nil {
		return nil, media.ErrProviderNotFound
	}
	if strings.TrimSpace(sel.APIKey) != "" {
		return f.inner.NewProvider(sel, model)
	}
	backend, ok := media.RouteModel(model)
	if !ok {
		return nil, media.ErrProviderNotFound
	}
	client := &http.Client{Timeout: 120 * time.Second}
	switch backend {
	case media.BackendVolcengine:
		if strings.TrimSpace(f.env.Volcengine.APIKey) == "" {
			return nil, media.ErrProviderNotFound
		}
		return providers.NewVolcengineProvider(f.env.Volcengine, client), nil
	case media.BackendOpenRouter:
		if strings.TrimSpace(f.env.OpenRouter.APIKey) == "" {
			return nil, media.ErrProviderNotFound
		}
		return providers.NewOpenRouterProvider(f.env.OpenRouter, client), nil
	default:
		return nil, media.ErrProviderNotFound
	}
}
