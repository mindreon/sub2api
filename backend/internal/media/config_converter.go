package media

import "context"

// ConfigBackedConverter 从 ConfigStore 读取 CNY→USD 汇率，缺省时用 fallback。
type ConfigBackedConverter struct {
	store    *ConfigStore
	fallback float64
	billing  Currency
}

// NewConfigBackedConverter 构造动态汇率换算器。
func NewConfigBackedConverter(store *ConfigStore, billing Currency, fallbackCNYToUSD float64) *ConfigBackedConverter {
	if billing == "" {
		billing = CurrencyUSD
	}
	return &ConfigBackedConverter{store: store, billing: billing, fallback: fallbackCNYToUSD}
}

// Convert 实现 CurrencyConverter。
func (c *ConfigBackedConverter) Convert(amount float64, from Currency) (float64, error) {
	if from == "" || from == c.billing {
		return amount, nil
	}
	if from != CurrencyCNY {
		return 0, ErrCurrencyRateMissing
	}
	rate := c.fallback
	if c.store != nil {
		cfg, err := c.store.Load(context.Background())
		if err == nil && cfg.CNYToUSDRate > 0 {
			rate = cfg.CNYToUSDRate
		}
	}
	if rate <= 0 {
		return 0, ErrCurrencyRateMissing
	}
	return amount * rate, nil
}
