package media

// StaticCurrencyConverter 用一组固定汇率把各货币换算到计费币种。
//
// 汇率语义：Rates[c] = 「1 单位货币 c 等于多少单位计费币种」。
// 例：计费币种 USD，1 CNY ≈ 0.14 USD，则 Rates[CurrencyCNY] = 0.14。
//
// 汇率应由配置注入（P4），本类型不内置任何生产汇率，避免把汇率写死进代码。
type StaticCurrencyConverter struct {
	BillingCurrency Currency
	Rates           map[Currency]float64
}

// NewStaticCurrencyConverter 构造换算器。billing 为空时默认 USD。
func NewStaticCurrencyConverter(billing Currency, rates map[Currency]float64) *StaticCurrencyConverter {
	if billing == "" {
		billing = CurrencyUSD
	}
	cp := make(map[Currency]float64, len(rates))
	for k, v := range rates {
		cp[k] = v
	}
	return &StaticCurrencyConverter{BillingCurrency: billing, Rates: cp}
}

// Convert 把 from 货币的 amount 换算为计费币种金额。
// from 已是计费币种时原样返回；缺少汇率时返回 ErrCurrencyRateMissing。
func (c *StaticCurrencyConverter) Convert(amount float64, from Currency) (float64, error) {
	if from == "" || from == c.BillingCurrency {
		return amount, nil
	}
	rate, ok := c.Rates[from]
	if !ok || rate <= 0 {
		return 0, ErrCurrencyRateMissing
	}
	return amount * rate, nil
}
