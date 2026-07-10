package media

import "context"

// HoldAwareBalance 从上游可用余额中扣除本包未结算的预扣（held）金额。
//
// 可用额度 = 上游余额 − Σ(held 状态预扣)
// P2 预扣检查应使用本包装器，而非直接读上游余额。
type HoldAwareBalance struct {
	upstream BalanceReader
	holds    HoldStore
}

// NewHoldAwareBalance 构造 HoldAware 余额读取器。
func NewHoldAwareBalance(upstream BalanceReader, holds HoldStore) BalanceReader {
	return &HoldAwareBalance{upstream: upstream, holds: holds}
}

func (h *HoldAwareBalance) AvailableBalance(ctx context.Context, userID int64) (float64, error) {
	if h == nil || h.upstream == nil {
		return 0, ErrUpstreamNotWired
	}
	raw, err := h.upstream.AvailableBalance(ctx, userID)
	if err != nil {
		return 0, err
	}
	if h.holds == nil {
		return raw, nil
	}
	held, err := h.holds.SumHeldByUser(ctx, userID)
	if err != nil {
		return 0, err
	}
	available := raw - held
	if available < 0 {
		return 0, nil
	}
	return available, nil
}
