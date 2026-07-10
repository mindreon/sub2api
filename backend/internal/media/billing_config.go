package media

import (
	"errors"
	"fmt"
	"strings"
)

// MediaCNYToUSDRateSettingKey 是 setting 表中 CNY→USD 汇率的独立 key（纯数字字符串）。
const MediaCNYToUSDRateSettingKey = "media_cny_to_usd_rate"

// MediaPricingOverridesSettingKey 是 setting 表中多模态价格覆盖规则的 JSON key。
const MediaPricingOverridesSettingKey = "media_pricing_overrides"

// legacyBillingConfigSettingKey 是旧版 JSON 配置 key，仅用于读取迁移汇率。
const legacyBillingConfigSettingKey = "media_billing_config"

var ErrInvalidPricingOverride = errors.New("media: invalid pricing override")

// PricingOverride 是管理端可配置的模型价格覆盖。
type PricingOverride struct {
	Model           string        `json:"model"`
	Metric          BillingMetric `json:"metric"`
	PricePerMillion float64       `json:"price_per_million"`
	Currency        Currency      `json:"currency"`
	Resolutions     []string      `json:"resolutions,omitempty"`
	HasVideoInput   *bool         `json:"has_video_input,omitempty"`
	HasAudio        *bool         `json:"has_audio,omitempty"`
}

// BillingConfig 存储可持久化的多模态计费站点配置。
type BillingConfig struct {
	CNYToUSDRate     float64           `json:"cny_to_usd_rate"`
	PricingOverrides []PricingOverride `json:"pricing_overrides"`
}

// BillingConfigPublic 是管理端 GET 响应用的安全视图。
type BillingConfigPublic struct {
	CNYToUSDRate     float64           `json:"cny_to_usd_rate"`
	PricingOverrides []PricingOverride `json:"pricing_overrides"`
}

// BillingConfigUpdate 是管理端 PUT 的局部更新载荷。
type BillingConfigUpdate struct {
	CNYToUSDRate     *float64           `json:"cny_to_usd_rate"`
	PricingOverrides *[]PricingOverride `json:"pricing_overrides"`
}

func (c BillingConfig) PublicView() BillingConfigPublic {
	return BillingConfigPublic{
		CNYToUSDRate:     c.CNYToUSDRate,
		PricingOverrides: append([]PricingOverride(nil), c.PricingOverrides...),
	}
}

func (c BillingConfig) ApplyUpdate(u BillingConfigUpdate) BillingConfig {
	out := c
	if u.CNYToUSDRate != nil {
		out.CNYToUSDRate = *u.CNYToUSDRate
	}
	if u.PricingOverrides != nil {
		out.PricingOverrides = append([]PricingOverride(nil), (*u.PricingOverrides)...)
	}
	return out
}

func (c BillingConfig) Validate() error {
	if c.CNYToUSDRate <= 0 {
		return ErrCurrencyRateMissing
	}
	for i := range c.PricingOverrides {
		if err := c.PricingOverrides[i].Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (o PricingOverride) Validate() error {
	if strings.TrimSpace(o.Model) == "" {
		return fmt.Errorf("%w: model is required", ErrInvalidPricingOverride)
	}
	if o.PricePerMillion <= 0 {
		return fmt.Errorf("%w: price_per_million must be positive", ErrInvalidPricingOverride)
	}
	if o.Currency != CurrencyCNY && o.Currency != CurrencyUSD {
		return fmt.Errorf("%w: currency must be CNY or USD", ErrInvalidPricingOverride)
	}
	metric := o.Metric
	if metric == "" {
		metric = MetricVideoToken
	}
	if metric != MetricVideoToken {
		return fmt.Errorf("%w: unsupported metric %q", ErrInvalidPricingOverride, metric)
	}
	return nil
}

func (o PricingOverride) matchesModel(model string) bool {
	return normalizeModelKey(o.Model) == normalizeModelKey(model)
}

func (o PricingOverride) toRule() MediaPricingRule {
	metric := o.Metric
	if metric == "" {
		metric = MetricVideoToken
	}
	unit, currency := PerMillion(o.PricePerMillion, o.Currency)
	return MediaPricingRule{
		Metric:            metric,
		UnitPrice:         unit,
		Currency:          currency,
		Resolutions:       append([]string(nil), o.Resolutions...),
		RequireVideoInput: o.HasVideoInput,
		RequireAudio:      o.HasAudio,
	}
}

func (c BillingConfig) RulesFor(model string) []MediaPricingRule {
	if len(c.PricingOverrides) == 0 {
		return nil
	}
	rules := make([]MediaPricingRule, 0, len(c.PricingOverrides))
	for _, override := range c.PricingOverrides {
		if override.matchesModel(model) {
			rules = append(rules, override.toRule())
		}
	}
	return rules
}
