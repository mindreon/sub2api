package media

import (
	"context"
	"log/slog"
	"time"
)

// Poller 周期性检查 pending/in_progress 任务的上游状态，驱动 Settle/Release。
//
// 这是"回收"路径的兜底实现（详见设计文档 §5）：即使 webhook 缺失或丢失，
// 轮询也能最终把每个任务收敛到终态，避免 hold 永久悬空占用用户额度。
type Poller struct {
	tasks         TaskStore
	ledger        *Ledger
	creds         AccountCredentialsLoader
	factories     ProviderFactory
	subscriptions SubscriptionResolver
	batchSize     int
	now           func() time.Time
	logger        *slog.Logger
}

// PollerConfig 控制轮询批次大小与时间源（测试可注入固定时钟）。
type PollerConfig struct {
	BatchSize            int
	Now                  func() time.Time
	Logger               *slog.Logger
	SubscriptionResolver SubscriptionResolver
}

// NewPoller 构造轮询 Worker。
func NewPoller(
	tasks TaskStore,
	ledger *Ledger,
	creds AccountCredentialsLoader,
	factories ProviderFactory,
	cfg PollerConfig,
) *Poller {
	batchSize := cfg.BatchSize
	if batchSize <= 0 {
		batchSize = 50
	}
	now := cfg.Now
	if now == nil {
		now = time.Now
	}
	logger := cfg.Logger
	if logger == nil {
		logger = slog.Default()
	}
	return &Poller{
		tasks:         tasks,
		ledger:        ledger,
		creds:         creds,
		factories:     factories,
		subscriptions: cfg.SubscriptionResolver,
		batchSize:     batchSize,
		now:           now,
		logger:        logger,
	}
}

// RunOnce 处理一批未终结任务，返回处理数量与遇到的（非致命）错误列表。
// 单个任务处理失败不影响其余任务，错误会被收集返回供调用方记录/告警。
func (p *Poller) RunOnce(ctx context.Context) (int, []error) {
	if p == nil || p.tasks == nil || p.ledger == nil || p.factories == nil {
		return 0, []error{ErrUpstreamNotWired}
	}

	tasks, err := p.tasks.ListByStatus(ctx, []TaskStatus{TaskPending, TaskInProgress}, p.batchSize)
	if err != nil {
		return 0, []error{err}
	}

	var errs []error
	for _, task := range tasks {
		if err := p.processOne(ctx, task); err != nil {
			errs = append(errs, err)
		}
	}
	return len(tasks), errs
}

// Run 按固定间隔循环调用 RunOnce，直到 ctx 被取消。适合在后台 goroutine 中启动。
func (p *Poller) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if _, errs := p.RunOnce(ctx); len(errs) > 0 {
				for _, e := range errs {
					p.logger.Error("media poller: task processing error", "error", e)
				}
			}
		}
	}
}

func (p *Poller) processOne(ctx context.Context, task *Task) error {
	now := p.now()

	// 过期兜底：无论 pending 还是 in_progress，只要过了 expires_at 就释放预扣，
	// 避免因厂商长期无响应或提交阶段崩溃导致 hold 永久占用额度。
	if !task.ExpiresAt.IsZero() && now.After(task.ExpiresAt) {
		_, err := p.ledger.Release(ctx, ReleaseInput{
			TaskID:       task.TaskID,
			Status:       TaskExpired,
			ErrorMessage: "poller: task expired before reaching terminal state",
		})
		return err
	}

	if task.Status == TaskPending {
		// 提交阶段中断（例如进程崩溃于 Reserve 和 Submit 之间）留下的孤儿任务：
		// 轮询不负责重新提交（那是 TaskService 的职责），只等待过期后释放。
		return nil
	}

	if task.UpstreamTaskID == "" {
		return nil // in_progress 但缺少上游任务 ID，异常数据，交给过期兜底处理
	}

	sel, err := p.loadAccountSelection(ctx, task)
	if err != nil {
		return err
	}

	provider, err := p.factories.NewProvider(sel, task.Model)
	if err != nil {
		return err
	}

	status, err := provider.QueryStatus(ctx, task)
	if err != nil {
		p.bumpAttempts(ctx, task, now)
		return err
	}

	switch status.State {
	case ProviderSucceeded:
		subscription, err := p.resolveSubscription(ctx, task)
		if err != nil {
			return err
		}
		var subscriptionID *int64
		isSubscription := false
		if subscription != nil && subscription.IsSubscription && subscription.SubscriptionID > 0 {
			id := subscription.SubscriptionID
			subscriptionID = &id
			isSubscription = true
		}
		_, err = p.ledger.Settle(ctx, SettleInput{
			TaskID:         task.TaskID,
			Usage:          status.Usage,
			UpstreamUsage:  status.RawUsage,
			ResultURL:      status.ResultURL,
			SubscriptionID: subscriptionID,
			IsSubscription: isSubscription,
		})
		return err
	case ProviderFailed:
		_, err := p.ledger.Release(ctx, ReleaseInput{
			TaskID:       task.TaskID,
			Status:       TaskFailed,
			ErrorMessage: status.ErrorMessage,
		})
		return err
	default: // ProviderInProgress：继续等待，仅记录轮询次数
		p.bumpAttempts(ctx, task, now)
		return nil
	}
}

func (p *Poller) resolveSubscription(ctx context.Context, task *Task) (*SubscriptionBilling, error) {
	if task != nil && task.SubscriptionID != nil && *task.SubscriptionID > 0 {
		return &SubscriptionBilling{SubscriptionID: *task.SubscriptionID, IsSubscription: true}, nil
	}
	if p == nil || p.subscriptions == nil || task == nil {
		return nil, nil
	}
	return p.subscriptions.ResolveSubscription(ctx, task.UserID, task.GroupID)
}

func (p *Poller) loadAccountSelection(ctx context.Context, task *Task) (AccountSelection, error) {
	if task.AccountID != nil && *task.AccountID > 0 && p.creds != nil {
		return p.creds.Load(ctx, *task.AccountID)
	}
	return AccountSelection{}, nil
}

func (p *Poller) bumpAttempts(ctx context.Context, task *Task, now time.Time) {
	task.PollAttempts++
	task.UpdatedAt = now
	_ = p.tasks.Update(ctx, task)
}
