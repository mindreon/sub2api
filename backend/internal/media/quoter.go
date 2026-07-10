package media

import (
	"context"
	"errors"
)

// ErrModelNotPriceable 表示没有任何定价规则来源能为该模型定价。
var ErrModelNotPriceable = errors.New("media: model not priceable")

// RuleProvider 按模型名返回候选定价规则。
//
// P1 默认用内置 Seedance 表（SeedanceRuleProvider）；
// 未来可替换为读取渠道定价的实现（端口注入，无需改本文件）。
type RuleProvider interface {
	RulesFor(model string) []MediaPricingRule
}

// RuleProviderFunc 让普通函数满足 RuleProvider。
type RuleProviderFunc func(model string) []MediaPricingRule

func (f RuleProviderFunc) RulesFor(model string) []MediaPricingRule { return f(model) }

// SeedanceRuleProvider 是基于内置火山 Seedance 单价矩阵的默认规则来源。
// 对 OpenRouter 模型名（bytedance/seedance-*）走 USD 定价表，其余走火山 CNY 表。
func SeedanceRuleProvider() RuleProvider {
	return RuleProviderFunc(func(model string) []MediaPricingRule {
		if backend, ok := RouteModel(model); ok && backend == BackendOpenRouter {
			if rules := OpenRouterSeedancePricingRules(model); len(rules) > 0 {
				return rules
			}
		}
		return SeedancePricingRules(model)
	})
}

// Quote 是一次费用预估结果（含原始币种明细与换算后计费币种金额）。
type Quote struct {
	Cost              *MediaCost // 引擎原始结果（币种见 Cost.Currency）
	BillingCurrency   Currency   // 平台计费币种
	TotalBillingCost  float64    // 换算后计费币种费用（未乘倍率）
	ActualBillingCost float64    // 换算后 × 倍率的实际费用
}

// Quoter 是多模态计费的门面：串联「规则 → 引擎 → 换算 → 余额检查」。
// P2/P3 的任务编排面向本门面开发，不直接触碰引擎细节。
type Quoter struct {
	rules           RuleProvider
	converter       CurrencyConverter
	balance         BalanceReader
	billingCurrency Currency
}

// NewQuoter 构造门面。billingCurrency 为空时默认 USD。
// balance 允许为 nil（仅预估、不做余额检查的场景）。
func NewQuoter(rules RuleProvider, converter CurrencyConverter, balance BalanceReader, billingCurrency Currency) *Quoter {
	if billingCurrency == "" {
		billingCurrency = CurrencyUSD
	}
	return &Quoter{rules: rules, converter: converter, balance: balance, billingCurrency: billingCurrency}
}

// Estimate 预估一次多模态调用的费用，并换算到计费币种。
func (q *Quoter) Estimate(model string, usage BillingUsage, rateMultiplier float64) (*Quote, error) {
	rules := q.rules.RulesFor(model)
	if len(rules) == 0 {
		return nil, ErrModelNotPriceable
	}

	cost, err := CalculateMediaCost(MediaCostInput{
		Rules:          rules,
		Usage:          usage,
		RateMultiplier: rateMultiplier,
	})
	if err != nil {
		return nil, err
	}

	totalBilling, err := q.converter.Convert(cost.TotalCost, cost.Currency)
	if err != nil {
		return nil, err
	}
	actualBilling, err := q.converter.Convert(cost.ActualCost, cost.Currency)
	if err != nil {
		return nil, err
	}

	return &Quote{
		Cost:              cost,
		BillingCurrency:   q.billingCurrency,
		TotalBillingCost:  totalBilling,
		ActualBillingCost: actualBilling,
	}, nil
}

// EnsureAffordable 检查用户可用额度是否覆盖 requiredBillingCost（计费币种）。
// 不足返回 ErrInsufficientBalance；balance 端口未配置时跳过检查。
func (q *Quoter) EnsureAffordable(ctx context.Context, userID int64, requiredBillingCost float64) error {
	if q.balance == nil {
		return nil
	}
	available, err := q.balance.AvailableBalance(ctx, userID)
	if err != nil {
		return err
	}
	if available < requiredBillingCost {
		return ErrInsufficientBalance
	}
	return nil
}
