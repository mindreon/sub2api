package media

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"sync"
)

// SettingsKV 是读写 setting 表所需的最小端口（避免 media 包 import service）。
type SettingsKV interface {
	GetValue(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string) error
}

// ConfigStore 从 setting 表加载/保存 media_cny_to_usd_rate，并做进程内缓存。
type ConfigStore struct {
	kv       SettingsKV
	fallback BillingConfig
	mu       sync.RWMutex
	cached   BillingConfig
	loaded   bool
}

// NewConfigStore 构造配置存储。fallback 在 DB 无配置或字段缺失时兜底（如默认汇率）。
func NewConfigStore(kv SettingsKV, fallback BillingConfig) *ConfigStore {
	return &ConfigStore{kv: kv, fallback: fallback}
}

// Load 返回有效汇率配置（优先独立 key，其次旧 JSON 配置，最后 fallback）。
func (s *ConfigStore) Load(ctx context.Context) (BillingConfig, error) {
	if s == nil {
		return BillingConfig{}, nil
	}
	if s.kv == nil {
		return s.fallback, nil
	}
	s.mu.RLock()
	if s.loaded {
		cfg := s.cached
		s.mu.RUnlock()
		return cfg, nil
	}
	s.mu.RUnlock()

	cfg := s.fallback
	if rate, ok, err := s.loadRateFromKey(ctx, MediaCNYToUSDRateSettingKey); err != nil {
		return s.fallback, err
	} else if ok {
		cfg.CNYToUSDRate = rate
	} else if rate, overrides, ok, err := s.loadFromLegacyJSON(ctx); err != nil {
		return s.fallback, err
	} else if ok {
		cfg.CNYToUSDRate = rate
		if len(overrides) > 0 {
			cfg.PricingOverrides = overrides
		}
	}
	if overrides, ok, err := s.loadPricingOverrides(ctx); err != nil {
		return s.fallback, err
	} else if ok {
		cfg.PricingOverrides = overrides
	}

	s.mu.Lock()
	s.cached = cfg
	s.loaded = true
	s.mu.Unlock()
	return cfg, nil
}

func (s *ConfigStore) loadRateFromKey(ctx context.Context, key string) (float64, bool, error) {
	raw, err := s.kv.GetValue(ctx, key)
	if err != nil {
		return 0, false, err
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, false, nil
	}
	rate, err := strconv.ParseFloat(raw, 64)
	if err != nil || rate <= 0 {
		return 0, false, nil
	}
	return rate, true, nil
}

func (s *ConfigStore) loadFromLegacyJSON(ctx context.Context) (float64, []PricingOverride, bool, error) {
	raw, err := s.kv.GetValue(ctx, legacyBillingConfigSettingKey)
	if err != nil {
		return 0, nil, false, err
	}
	if strings.TrimSpace(raw) == "" {
		return 0, nil, false, nil
	}
	var stored struct {
		CNYToUSDRate     float64           `json:"cny_to_usd_rate"`
		PricingOverrides []PricingOverride `json:"pricing_overrides"`
	}
	if err := json.Unmarshal([]byte(raw), &stored); err != nil {
		return 0, nil, false, nil
	}
	if stored.CNYToUSDRate <= 0 {
		return 0, nil, false, nil
	}
	overrides := sanitizePricingOverrides(stored.PricingOverrides)
	return stored.CNYToUSDRate, overrides, true, nil
}

func (s *ConfigStore) loadPricingOverrides(ctx context.Context) ([]PricingOverride, bool, error) {
	raw, err := s.kv.GetValue(ctx, MediaPricingOverridesSettingKey)
	if err != nil {
		return nil, false, err
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, false, nil
	}
	var overrides []PricingOverride
	if err := json.Unmarshal([]byte(raw), &overrides); err != nil {
		return nil, false, nil
	}
	return sanitizePricingOverrides(overrides), true, nil
}

// Save 持久化汇率并刷新缓存（仅写入 media_cny_to_usd_rate）。
func (s *ConfigStore) Save(ctx context.Context, cfg BillingConfig) error {
	if s == nil || s.kv == nil {
		return ErrUpstreamNotWired
	}
	if err := cfg.Validate(); err != nil {
		return err
	}
	value := strconv.FormatFloat(cfg.CNYToUSDRate, 'f', -1, 64)
	if err := s.kv.Set(ctx, MediaCNYToUSDRateSettingKey, value); err != nil {
		return err
	}
	rawOverrides, err := json.Marshal(cfg.PricingOverrides)
	if err != nil {
		return err
	}
	if err := s.kv.Set(ctx, MediaPricingOverridesSettingKey, string(rawOverrides)); err != nil {
		return err
	}
	s.mu.Lock()
	s.cached = cfg
	s.loaded = true
	s.mu.Unlock()
	return nil
}

// Invalidate 清除进程内缓存（测试或热更新后可选调用）。
func (s *ConfigStore) Invalidate() {
	if s == nil {
		return
	}
	s.mu.Lock()
	s.loaded = false
	s.mu.Unlock()
}

func sanitizePricingOverrides(overrides []PricingOverride) []PricingOverride {
	if len(overrides) == 0 {
		return nil
	}
	out := make([]PricingOverride, 0, len(overrides))
	for _, override := range overrides {
		if err := override.Validate(); err == nil {
			out = append(out, override)
		}
	}
	return out
}
