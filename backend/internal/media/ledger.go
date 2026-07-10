package media

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

// ReserveInput 创建异步任务并预扣额度。
type ReserveInput struct {
	TaskID         string // 调用方指定或留空自动生成
	UserID         int64
	APIKeyID       int64
	AccountID      int64
	GroupID        *int64
	SubscriptionID *int64
	IsSubscription bool
	Model          string
	MediaType      string
	Usage          BillingUsage
	RateMultiplier float64
	RequestParams  map[string]any
	ExpiresAt      time.Time
}

// SettleInput 任务成功完成后结算真实费用。
type SettleInput struct {
	TaskID         string
	Usage          BillingUsage // 上游返回后的实际用量
	UpstreamUsage  map[string]any
	ResultURL      string // 上游生成结果链接；结算时尽力转存到自有存储
	SubscriptionID *int64
	IsSubscription bool
}

// ReleaseInput 任务失败/超时/取消时释放预扣。
type ReleaseInput struct {
	TaskID       string
	Status       TaskStatus // TaskFailed / TaskExpired / TaskCancelled
	ErrorMessage string
}

// Ledger 编排预扣 → 结算 → 释放三段式额度管理。
type Ledger struct {
	quoter  *Quoter
	charger Charger
	tasks   TaskStore
	holds   HoldStore
	balance BalanceReader // 应为 HoldAwareBalance

	// 可选依赖（组合根按需注入）：
	assets AssetStore // 结算成功后把上游结果转存自有存储；nil 时降级保留上游直链
	logger *slog.Logger

	// 严格预扣事务（生产路径）：设置后 Reserve 走原子建库 + 锁定重算余额。
	// reservation 为 nil 时退回非事务的 legacy 路径（内存测试用）。
	reservation Reservation
	rawBalance  BalanceReader // 未扣预扣的上游余额；reservation 路径用它校验
}

// NewLedger 构造台账编排器。balance 建议传入 NewHoldAwareBalance(upstream, holds)。
func NewLedger(quoter *Quoter, charger Charger, tasks TaskStore, holds HoldStore, balance BalanceReader) *Ledger {
	return &Ledger{
		quoter:  quoter,
		charger: charger,
		tasks:   tasks,
		holds:   holds,
		balance: balance,
		logger:  slog.Default(),
	}
}

// WithAssetStore 注入结果转存存储（可选）。返回自身便于链式装配。
func (l *Ledger) WithAssetStore(assets AssetStore) *Ledger {
	if l != nil {
		l.assets = assets
	}
	return l
}

// WithLogger 覆盖默认日志器（可选）。
func (l *Ledger) WithLogger(logger *slog.Logger) *Ledger {
	if l != nil && logger != nil {
		l.logger = logger
	}
	return l
}

// WithReservation 启用严格预扣事务路径。rawBalance 必须是未扣预扣的上游余额，
// held 总额由 reservation 在事务内重算。返回自身便于链式装配。
func (l *Ledger) WithReservation(reservation Reservation, rawBalance BalanceReader) *Ledger {
	if l != nil {
		l.reservation = reservation
		l.rawBalance = rawBalance
	}
	return l
}

// Reserve 预估费用、检查余额、创建任务与预扣记录。
func (l *Ledger) Reserve(ctx context.Context, in ReserveInput) (*Task, error) {
	if l == nil || l.quoter == nil || l.tasks == nil || l.holds == nil {
		return nil, ErrUpstreamNotWired
	}

	taskID := in.TaskID
	if taskID == "" {
		taskID = uuid.NewString()
	}
	if existing, err := l.tasks.GetByTaskID(ctx, taskID); err == nil && existing != nil {
		return existing, nil // 幂等：同一 task_id 重复提交返回已有任务
	} else if err != nil && !errors.Is(err, ErrTaskNotFound) {
		return nil, err
	}

	quote, err := l.quoter.Estimate(in.Model, in.Usage, in.RateMultiplier)
	if err != nil {
		return nil, err
	}

	reserveAmount := quote.ActualBillingCost

	expiresAt := in.ExpiresAt
	if expiresAt.IsZero() {
		expiresAt = time.Now().Add(30 * time.Minute)
	}

	var accountID *int64
	if in.AccountID > 0 {
		accountID = &in.AccountID
	}

	now := time.Now()
	task := &Task{
		TaskID:          taskID,
		UserID:          in.UserID,
		APIKeyID:        in.APIKeyID,
		AccountID:       accountID,
		GroupID:         in.GroupID,
		SubscriptionID:  snapshotSubscriptionID(in.SubscriptionID, in.IsSubscription),
		Model:           in.Model,
		MediaType:       in.MediaType,
		Status:          TaskPending,
		BillingMetric:   quote.Cost.Metric,
		ReservedCost:    reserveAmount,
		RateMultiplier:  in.RateMultiplier,
		BillingCurrency: quote.BillingCurrency,
		RequestParams:   in.RequestParams,
		ExpiresAt:       expiresAt,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	hold := &Hold{
		HoldID:    uuid.NewString(),
		TaskID:    taskID,
		UserID:    in.UserID,
		Amount:    reserveAmount,
		Currency:  quote.BillingCurrency,
		Status:    HoldHeld,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// 严格路径：单事务内原子建 task+hold，并按锁定重算的 held 校验上游余额。
	if l.reservation != nil {
		avail, err := l.reserveAvailable(ctx, in.UserID)
		if err != nil {
			return nil, err
		}
		if err := l.reservation.Reserve(ctx, task, hold, avail); err != nil {
			return nil, err
		}
		return task, nil
	}

	// legacy 路径（内存测试）：非事务，用 hold-aware 余额检查。
	if err := l.ensureAffordable(ctx, in.UserID, reserveAmount); err != nil {
		return nil, err
	}
	if err := l.tasks.Create(ctx, task); err != nil {
		return nil, err
	}
	if err := l.holds.Create(ctx, hold); err != nil {
		return nil, err
	}
	return task, nil
}

// reserveAvailable 返回严格路径用于校验的上游可用余额（未扣预扣）。
// 优先用 rawBalance；未注入时退回 balance（可能是 hold-aware，此时 held 会被双减，
// 校验偏保守，仍安全）。
func (l *Ledger) reserveAvailable(ctx context.Context, userID int64) (float64, error) {
	reader := l.rawBalance
	if reader == nil {
		reader = l.balance
	}
	if reader == nil {
		return 0, ErrUpstreamNotWired
	}
	return reader.AvailableBalance(ctx, userID)
}

// MarkInProgress 任务已提交上游，更新状态、上游任务 ID 与调度账号。
func (l *Ledger) MarkInProgress(ctx context.Context, taskID, upstreamTaskID string, accountID int64) (*Task, error) {
	task, err := l.tasks.GetByTaskID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if task.Status.IsTerminal() {
		return nil, ErrTaskAlreadyTerminal
	}
	if !task.Status.CanTransitionTo(TaskInProgress) {
		return nil, ErrInvalidTaskTransition
	}
	task.Status = TaskInProgress
	task.UpstreamTaskID = upstreamTaskID
	if accountID > 0 {
		task.AccountID = &accountID
	}
	task.UpdatedAt = time.Now()
	if err := l.tasks.Update(ctx, task); err != nil {
		return nil, err
	}
	return task, nil
}

// Settle 按实际用量结算：扣费 + 释放预扣差额 + 标记任务完成。
func (l *Ledger) Settle(ctx context.Context, in SettleInput) (*Task, error) {
	if l == nil || l.quoter == nil || l.tasks == nil || l.holds == nil {
		return nil, ErrUpstreamNotWired
	}

	task, err := l.tasks.GetByTaskID(ctx, in.TaskID)
	if err != nil {
		return nil, err
	}
	if task.Status == TaskCompleted {
		return task, nil // 幂等
	}
	if !task.Status.CanTransitionTo(TaskCompleted) {
		return nil, ErrInvalidTaskTransition
	}

	hold, err := l.holds.GetByTaskID(ctx, in.TaskID)
	if err != nil {
		return nil, err
	}

	quote, err := l.quoter.Estimate(task.Model, in.Usage, task.RateMultiplier)
	if err != nil {
		return nil, err
	}

	actualCost := quote.ActualBillingCost

	if l.charger != nil && actualCost > 0 {
		subscriptionID, isSubscription := settleSubscriptionSnapshot(task, in)
		_, err = l.charger.Charge(ctx, ChargeRequest{
			RequestID:           task.TaskID,
			UserID:              task.UserID,
			APIKeyID:            task.APIKeyID,
			AccountID:           derefInt64(task.AccountID),
			GroupID:             task.GroupID,
			SubscriptionID:      subscriptionID,
			Model:               task.Model,
			MediaType:           task.MediaType,
			Metric:              quote.Cost.Metric,
			Units:               quote.Cost.Units,
			CostBillingCurrency: quote.TotalBillingCost,
			ActualCost:          actualCost,
			RateMultiplier:      task.RateMultiplier,
			IsSubscription:      isSubscription,
		})
		if err != nil {
			return nil, err
		}
	}

	if hold.Status == HoldHeld {
		if err := l.holds.UpdateStatus(ctx, hold.HoldID, HoldHeld, HoldSettled); err != nil {
			return nil, err
		}
	}

	// 尽力把上游结果转存到自有存储。转存是 best-effort：算力已消耗、扣费已完成，
	// 转存失败只降级为保留上游直链（可能 24h 过期），绝不阻断结算。
	resultURL, storageKey := l.rehostResult(ctx, task, in.ResultURL)

	now := time.Now()
	task.Status = TaskCompleted
	task.ActualCost = &actualCost
	task.BillingMetric = quote.Cost.Metric
	task.UpstreamUsage = in.UpstreamUsage
	task.ResultURL = resultURL
	task.ResultStorageKey = storageKey
	task.SettledAt = &now
	task.UpdatedAt = now
	if err := l.tasks.Update(ctx, task); err != nil {
		return nil, err
	}
	return task, nil
}

// Release 失败/超时/取消：释放预扣，不写 usage_log。
func (l *Ledger) Release(ctx context.Context, in ReleaseInput) (*Task, error) {
	if l == nil || l.tasks == nil || l.holds == nil {
		return nil, ErrUpstreamNotWired
	}

	task, err := l.tasks.GetByTaskID(ctx, in.TaskID)
	if err != nil {
		return nil, err
	}
	if task.Status.IsTerminal() {
		return task, nil // 幂等
	}
	if !task.Status.CanTransitionTo(in.Status) {
		return nil, ErrInvalidTaskTransition
	}

	hold, err := l.holds.GetByTaskID(ctx, in.TaskID)
	if err != nil {
		return nil, err
	}
	if hold.Status == HoldHeld {
		if err := l.holds.UpdateStatus(ctx, hold.HoldID, HoldHeld, HoldReleased); err != nil {
			return nil, err
		}
	}

	task.Status = in.Status
	task.ErrorMessage = in.ErrorMessage
	task.UpdatedAt = time.Now()
	if err := l.tasks.Update(ctx, task); err != nil {
		return nil, err
	}
	return task, nil
}

// rehostResult 尽力把上游结果地址转存到自有存储，返回 (对客户端展示的链接, 存储对象键)。
//
//	未配置存储 / 上游无地址 / 转存失败 → 返回 (上游直链, "")，并记录降级日志。
//	转存成功 → 返回 (存储链接, 存储键)。
func (l *Ledger) rehostResult(ctx context.Context, task *Task, upstreamURL string) (resultURL, storageKey string) {
	if l.assets == nil || upstreamURL == "" {
		return upstreamURL, ""
	}
	key := mediaAssetKey(task.UserID, task.TaskID)
	url, err := l.assets.Rehost(ctx, key, upstreamURL, "video/mp4")
	if err != nil || url == "" {
		if l.logger != nil {
			l.logger.Warn("media ledger: rehost failed, falling back to upstream url",
				"task_id", task.TaskID, "error", err)
		}
		return upstreamURL, ""
	}
	return url, key
}

func (l *Ledger) ensureAffordable(ctx context.Context, userID int64, amount float64) error {
	if l.balance == nil {
		return nil
	}
	available, err := l.balance.AvailableBalance(ctx, userID)
	if err != nil {
		return err
	}
	if available < amount {
		return ErrInsufficientBalance
	}
	return nil
}

func derefInt64(v *int64) int64 {
	if v == nil {
		return 0
	}
	return *v
}

func snapshotSubscriptionID(id *int64, isSubscription bool) *int64 {
	if !isSubscription || id == nil || *id <= 0 {
		return nil
	}
	cp := *id
	return &cp
}

func settleSubscriptionSnapshot(task *Task, in SettleInput) (*int64, bool) {
	if task != nil && task.SubscriptionID != nil && *task.SubscriptionID > 0 {
		id := *task.SubscriptionID
		return &id, true
	}
	return snapshotSubscriptionID(in.SubscriptionID, in.IsSubscription), in.IsSubscription && in.SubscriptionID != nil && *in.SubscriptionID > 0
}
