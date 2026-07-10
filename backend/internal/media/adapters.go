package media

import (
	"context"
	"errors"
)

// 本文件把上游能力桥接为 media 的端口实现。
//
// 关键解耦手法：用「结构化接口」描述上游方法形状，而非 import 上游包。
// Go 的结构化类型让 *service.BillingCacheService 等具体类型在组合根（wire，P4）
// 被直接传入即可满足这些接口，因此 media 包始终零上游 import。

// ErrUpstreamNotWired 表示适配器未注入上游实现。
var ErrUpstreamNotWired = errors.New("media: upstream dependency not wired")

// UpstreamBalanceSource 是对上游"余额读取"的最小结构化描述。
// *service.BillingCacheService 的 GetUserBalance(ctx, userID) (float64, error) 天然满足。
type UpstreamBalanceSource interface {
	GetUserBalance(ctx context.Context, userID int64) (float64, error)
}

// NewBalanceReader 用上游余额源构造 BalanceReader 端口。
//
// 说明：此处返回的是"原始可用余额"。异步任务的"扣除未结算预扣（holds）"
// 由 P2 的台账在此之上再包一层（HoldAware），本适配器只负责读上游真实余额。
func NewBalanceReader(src UpstreamBalanceSource) BalanceReader {
	return &balanceReaderAdapter{src: src}
}

type balanceReaderAdapter struct {
	src UpstreamBalanceSource
}

func (a *balanceReaderAdapter) AvailableBalance(ctx context.Context, userID int64) (float64, error) {
	if a == nil || a.src == nil {
		return 0, ErrUpstreamNotWired
	}
	return a.src.GetUserBalance(ctx, userID)
}

// 关于 Charger 的真实实现：
//
// 结算扣费需要把 ChargeRequest 翻译成上游的 UsageBillingCommand 并调用
// UsageBillingRepository.Apply，这会引用上游具体类型。为保持 media 零上游 import，
// 该翻译放在组合根（P4，一个 fork 自有的 wiring 文件，同样不属于上游文件，
// 不会引发 merge 冲突）实现，并作为 Charger 注入本包。
// P1 阶段 Charger 通过接口/ChargerFunc 供 quoter/task_service 面向契约开发与测试。
