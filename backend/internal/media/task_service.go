package media

import "context"

// TaskService 编排"提交异步任务"的完整生命周期：预扣额度 → 选择账号 → 提交上游厂商 → 标记进行中。
//
// 任意阶段失败都会释放预扣，绝不遗留悬空 hold：
//
//	Reserve 失败      -> 直接返回错误，未创建任务/预扣
//	找不到 Provider   -> Release(failed)
//	Provider.Submit 失败 -> Release(failed)
//	MarkInProgress 失败  -> 返回错误（任务与预扣仍存在，交给轮询兜底处理）
type TaskService struct {
	ledger        *Ledger
	accounts      AccountSelector
	creds         AccountCredentialsLoader
	factories     ProviderFactory
	subscriptions SubscriptionResolver
}

// NewTaskService 构造任务编排器。
func NewTaskService(
	ledger *Ledger,
	accounts AccountSelector,
	creds AccountCredentialsLoader,
	factories ProviderFactory,
	subscriptions SubscriptionResolver,
) *TaskService {
	return &TaskService{
		ledger:        ledger,
		accounts:      accounts,
		creds:         creds,
		factories:     factories,
		subscriptions: subscriptions,
	}
}

// Submit 预扣额度并把任务提交到上游厂商，返回状态为 in_progress 的任务。
//
// 幂等：相同 TaskID 重复调用时，若任务已存在且不是 pending（说明已经提交过），
// 直接返回现有任务，不会重复提交上游或重复预扣。
func (s *TaskService) Submit(ctx context.Context, in ReserveInput) (*Task, error) {
	if s == nil || s.ledger == nil || s.factories == nil {
		return nil, ErrUpstreamNotWired
	}

	reserve := in
	if err := s.applySubscriptionSnapshot(ctx, &reserve); err != nil {
		return nil, err
	}

	task, err := s.ledger.Reserve(ctx, reserve)
	if err != nil {
		return nil, err
	}

	if task.Status != TaskPending {
		return task, nil // 幂等命中：已提交过
	}

	sel, err := s.selectAccount(ctx, reserve, task.Model)
	if err != nil {
		return s.releaseAsFailed(ctx, task.TaskID, "account selection failed: "+err.Error(), err)
	}

	provider, err := s.factories.NewProvider(sel, task.Model)
	if err != nil {
		return s.releaseAsFailed(ctx, task.TaskID, "no provider registered: "+err.Error(), err)
	}

	upstreamTaskID, err := provider.Submit(ctx, task)
	if err != nil {
		return s.releaseAsFailed(ctx, task.TaskID, "submit to upstream failed: "+err.Error(), err)
	}

	return s.ledger.MarkInProgress(ctx, task.TaskID, upstreamTaskID, sel.AccountID)
}

func (s *TaskService) applySubscriptionSnapshot(ctx context.Context, in *ReserveInput) error {
	if s == nil || s.subscriptions == nil || in == nil {
		return nil
	}
	subscription, err := s.subscriptions.ResolveSubscription(ctx, in.UserID, in.GroupID)
	if err != nil {
		return err
	}
	if subscription == nil || !subscription.IsSubscription || subscription.SubscriptionID <= 0 {
		return nil
	}
	id := subscription.SubscriptionID
	in.SubscriptionID = &id
	in.IsSubscription = true
	return nil
}

func (s *TaskService) selectAccount(ctx context.Context, in ReserveInput, model string) (AccountSelection, error) {
	if in.AccountID > 0 && s.creds != nil {
		return s.creds.Load(ctx, in.AccountID)
	}
	if in.GroupID != nil && *in.GroupID > 0 && s.accounts != nil {
		return s.accounts.Select(ctx, AccountSelectInput{
			GroupID: *in.GroupID,
			Model:   model,
		})
	}
	return AccountSelection{}, nil
}

// releaseAsFailed 释放预扣并把原始错误透传给调用方，方便上层区分
// "计费/预扣问题" 和 "厂商提交问题"。
func (s *TaskService) releaseAsFailed(ctx context.Context, taskID, message string, cause error) (*Task, error) {
	if _, relErr := s.ledger.Release(ctx, ReleaseInput{
		TaskID:       taskID,
		Status:       TaskFailed,
		ErrorMessage: message,
	}); relErr != nil {
		return nil, relErr
	}
	return nil, cause
}
