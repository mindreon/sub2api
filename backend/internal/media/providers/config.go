package providers

import (
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	defaultVolcengineBaseURL = "https://ark.cn-beijing.volces.com"
	defaultOpenRouterBaseURL = "https://openrouter.ai"

	envVolcengineAPIKey  = "MEDIA_VOLCENGINE_API_KEY"
	envVolcengineBaseURL = "MEDIA_VOLCENGINE_BASE_URL"
	envOpenRouterAPIKey  = "MEDIA_OPENROUTER_API_KEY"
	envOpenRouterBaseURL = "MEDIA_OPENROUTER_BASE_URL"
)

// Config 是厂商适配器运行所需的最小配置（P5 从环境变量注入，后续可改为 settingService）。
type Config struct {
	Volcengine VolcengineConfig
	OpenRouter OpenRouterConfig
	HTTPClient *http.Client
}

// VolcengineConfig 火山方舟 Seedance 接入参数。
type VolcengineConfig struct {
	APIKey  string
	BaseURL string
}

// OpenRouterConfig OpenRouter 异步视频接入参数。
type OpenRouterConfig struct {
	APIKey  string
	BaseURL string
}

// ConfigFromEnv 从环境变量读取厂商配置（未设置 API Key 的厂商不会被注册）。
func ConfigFromEnv() Config {
	return Config{
		Volcengine: VolcengineConfig{
			APIKey:  strings.TrimSpace(os.Getenv(envVolcengineAPIKey)),
			BaseURL: strings.TrimSpace(os.Getenv(envVolcengineBaseURL)),
		},
		OpenRouter: OpenRouterConfig{
			APIKey:  strings.TrimSpace(os.Getenv(envOpenRouterAPIKey)),
			BaseURL: strings.TrimSpace(os.Getenv(envOpenRouterBaseURL)),
		},
		HTTPClient: &http.Client{Timeout: 120 * time.Second},
	}
}

func (c VolcengineConfig) enabled() bool  { return strings.TrimSpace(c.APIKey) != "" }
func (c OpenRouterConfig) enabled() bool { return strings.TrimSpace(c.APIKey) != "" }

func (c VolcengineConfig) baseURL() string {
	if u := strings.TrimRight(strings.TrimSpace(c.BaseURL), "/"); u != "" {
		return u
	}
	return defaultVolcengineBaseURL
}

func (c OpenRouterConfig) baseURL() string {
	if u := strings.TrimRight(strings.TrimSpace(c.BaseURL), "/"); u != "" {
		return u
	}
	return defaultOpenRouterBaseURL
}
