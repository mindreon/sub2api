package media

import (
	"context"
	"errors"
)

// 本文件定义 media 包对"上游能力"的最小需求（端口 / Ports）。
//
// 解耦要点：这些接口用 media 自己的领域类型表达，不引用任何上游具体类型。
// 上游的具体实现（如 *service.BillingCacheService、UsageBillingRepository）
// 通过结构化接口在组合根（wire，P4）注入，因此本包无需 import 上游包。

// ErrInsufficientBalance 表示用户可用额度不足以覆盖预估费用。
var ErrInsufficientBalance = errors.New("media: insufficient balance")

// ErrCurrencyRateMissing 表示缺少某货币到计费币种的换算汇率。
var ErrCurrencyRateMissing = errors.New("media: currency rate missing")

// BalanceReader 读取用户可用额度（计费币种，通常为 USD）。
type BalanceReader interface {
	AvailableBalance(ctx context.Context, userID int64) (float64, error)
}

// CurrencyConverter 把某货币金额换算为平台计费币种金额。
// Seedance 单价以人民币表达，落到用户余额前需换算。
type CurrencyConverter interface {
	Convert(amount float64, from Currency) (float64, error)
}

// Charger 执行真实扣费（幂等）。仅在异步任务"成功结算"时调用一次。
//
// 具体实现在 P3/P4 用上游 UsageBillingRepository.Apply 完成；
// P1 只固定契约，便于 quoter/task_service 面向接口开发与测试。
type Charger interface {
	Charge(ctx context.Context, req ChargeRequest) (*ChargeResult, error)
}

// ChargeRequest 是一次结算扣费的输入。
//
// RequestID 用作幂等键：同一 RequestID 重复结算必须只扣一次
// （对齐上游 UsageBillingCommand 的去重语义）。
type ChargeRequest struct {
	RequestID      string // 幂等键（通常复用任务 ID / 请求 ID）
	UserID         int64
	APIKeyID       int64
	AccountID      int64
	GroupID        *int64
	SubscriptionID *int64

	Model     string        // 计费模型名
	MediaType string        // video/audio/image
	Metric    BillingMetric // 命中的计费度量
	Units     int64         // 实际计费数量

	CostBillingCurrency float64 // 已换算到计费币种的费用（未乘倍率）
	ActualCost          float64 // 已乘倍率的实际扣费
	RateMultiplier      float64
	IsSubscription      bool
}

// ChargeResult 是扣费结果。
type ChargeResult struct {
	Applied    bool     // 是否发生扣费（false 表示幂等命中，已扣过）
	NewBalance *float64 // 扣费后余额（nil 表示非余额计费）
}

// SubscriptionBilling 是结算时需要携带的订阅计费上下文。
type SubscriptionBilling struct {
	SubscriptionID int64
	IsSubscription bool
}

// SubscriptionResolver 在异步结算前解析任务所属 user/group 是否应走订阅计费。
type SubscriptionResolver interface {
	ResolveSubscription(ctx context.Context, userID int64, groupID *int64) (*SubscriptionBilling, error)
}

type SubscriptionResolverFunc func(ctx context.Context, userID int64, groupID *int64) (*SubscriptionBilling, error)

func (f SubscriptionResolverFunc) ResolveSubscription(ctx context.Context, userID int64, groupID *int64) (*SubscriptionBilling, error) {
	return f(ctx, userID, groupID)
}

// --- 函数式适配器：便于用闭包/测试快速实现端口 ---

// BalanceReaderFunc 让普通函数满足 BalanceReader。
type BalanceReaderFunc func(ctx context.Context, userID int64) (float64, error)

func (f BalanceReaderFunc) AvailableBalance(ctx context.Context, userID int64) (float64, error) {
	return f(ctx, userID)
}

// ChargerFunc 让普通函数满足 Charger。
type ChargerFunc func(ctx context.Context, req ChargeRequest) (*ChargeResult, error)

func (f ChargerFunc) Charge(ctx context.Context, req ChargeRequest) (*ChargeResult, error) {
	return f(ctx, req)
}
