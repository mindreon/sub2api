package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// DistributionLevelConfig represents a commission level definition for agents.
// It is stored as JSON in settings or channel organization config.
type DistributionLevelConfig struct {
	Code           string  `json:"code"`
	Name           string  `json:"name"`
	CommissionRate float64 `json:"commission_rate"`
	Active         bool    `json:"active"`
	SortOrder      int     `json:"sort_order"`
	Note           string  `json:"note"`
}

func normalizeDistributionLevelConfig(cfg DistributionLevelConfig) DistributionLevelConfig {
	cfg.Code = strings.ToUpper(strings.TrimSpace(cfg.Code))
	cfg.Name = strings.TrimSpace(cfg.Name)
	if cfg.Name == "" {
		cfg.Name = cfg.Code
	}
	if cfg.CommissionRate < 0 {
		cfg.CommissionRate = 0
	}
	if cfg.CommissionRate > 100 {
		cfg.CommissionRate = 100
	}
	cfg.Note = strings.TrimSpace(cfg.Note)
	return cfg
}

func normalizeDistributionLevelConfigs(configs []DistributionLevelConfig) []DistributionLevelConfig {
	out := make([]DistributionLevelConfig, 0, len(configs))
	for _, cfg := range configs {
		cfg = normalizeDistributionLevelConfig(cfg)
		if cfg.Code == "" || cfg.Name == "" {
			continue
		}
		out = append(out, cfg)
	}
	return out
}

func parseDistributionLevelConfigs(raw any) ([]DistributionLevelConfig, error) {
	switch v := raw.(type) {
	case nil:
		return nil, nil
	case []DistributionLevelConfig:
		return normalizeDistributionLevelConfigs(v), nil
	case string:
		if strings.TrimSpace(v) == "" {
			return nil, nil
		}
		var out []DistributionLevelConfig
		if err := json.Unmarshal([]byte(v), &out); err != nil {
			return nil, err
		}
		return normalizeDistributionLevelConfigs(out), nil
	case []any, map[string]any, []map[string]any:
		data, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		var out []DistributionLevelConfig
		if err := json.Unmarshal(data, &out); err != nil {
			return nil, err
		}
		return normalizeDistributionLevelConfigs(out), nil
	default:
		return nil, fmt.Errorf("unsupported distribution levels type %T", raw)
	}
}

func mustParseDistributionLevelConfigs(raw any) []DistributionLevelConfig {
	configs, err := parseDistributionLevelConfigs(raw)
	if err != nil {
		return nil
	}
	return configs
}

func findDistributionLevelConfig(configs []DistributionLevelConfig, code string) (DistributionLevelConfig, bool) {
	code = strings.ToUpper(strings.TrimSpace(code))
	if code == "" {
		return DistributionLevelConfig{}, false
	}
	for _, cfg := range configs {
		if strings.EqualFold(strings.TrimSpace(cfg.Code), code) && cfg.Active {
			return normalizeDistributionLevelConfig(cfg), true
		}
	}
	return DistributionLevelConfig{}, false
}

func distributionLevelRateToPercent(rate float64) float64 {
	if rate <= 0 {
		return 0
	}
	if rate > 100 {
		return 100
	}
	return rate / 100
}

func distributionLevelRateToRaw(rate float64) float64 {
	if rate < 0 {
		return 0
	}
	if rate <= 1 {
		return rate * 100
	}
	if rate > 100 {
		return 100
	}
	return rate
}

var errDistributionLevelNotFound = errors.New("distribution level not found")
