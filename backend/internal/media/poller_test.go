package media

import (
	"context"
	"testing"
	"time"
)

func setupInProgressTask(t *testing.T, ledger *Ledger, taskID string, expiresAt time.Time) *Task {
	t.Helper()
	task, err := ledger.Reserve(context.Background(), ReserveInput{
		TaskID: taskID, UserID: 1, APIKeyID: 2,
		Model: "doubao-seedance-2.0", MediaType: "video",
		Usage: seedanceUsage(), RateMultiplier: 1, ExpiresAt: expiresAt,
	})
	if err != nil {
		t.Fatalf("reserve: %v", err)
	}
	task, err = ledger.MarkInProgress(context.Background(), taskID, "up-"+taskID, 1)
	if err != nil {
		t.Fatalf("mark in progress: %v", err)
	}
	return task
}

func TestPoller_SettlesOnSuccess(t *testing.T) {
	ledger, mem, charger := newTestLedger(t, 10)
	setupInProgressTask(t, ledger, "poll-success", time.Now().Add(time.Hour))

	usage := seedanceUsage()
	provider := &fakeProvider{
		statusFunc: func(task *Task) (*ProviderStatus, error) {
			return &ProviderStatus{State: ProviderSucceeded, Usage: usage}, nil
		},
	}
	loader := testAccountLoader{sel: AccountSelection{AccountID: 1, APIKey: "test-key"}}
	factory := testProviderFactory{p: provider}
	poller := NewPoller(mem, ledger, loader, factory, PollerConfig{})

	processed, errs := poller.RunOnce(context.Background())
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if processed != 1 {
		t.Fatalf("expected 1 processed, got %d", processed)
	}

	task, err := mem.GetByTaskID(context.Background(), "poll-success")
	if err != nil {
		t.Fatalf("get task: %v", err)
	}
	if task.Status != TaskCompleted {
		t.Fatalf("expected completed, got %s", task.Status)
	}
	if len(charger.calls) != 1 {
		t.Fatalf("expected 1 charge, got %d", len(charger.calls))
	}
}

type testSubscriptionResolver struct {
	sub   *SubscriptionBilling
	calls int
}

func (r *testSubscriptionResolver) ResolveSubscription(ctx context.Context, userID int64, groupID *int64) (*SubscriptionBilling, error) {
	_ = ctx
	_ = userID
	_ = groupID
	r.calls++
	return r.sub, nil
}

func TestPoller_SettlesSubscriptionGroupUsage(t *testing.T) {
	ledger, mem, charger := newTestLedger(t, 10)
	groupID := int64(9)
	subID := int64(44)
	task, err := ledger.Reserve(context.Background(), ReserveInput{
		TaskID:         "poll-subscription",
		UserID:         1,
		APIKeyID:       2,
		GroupID:        &groupID,
		Model:          "doubao-seedance-2.0",
		MediaType:      "video",
		Usage:          seedanceUsage(),
		RateMultiplier: 1,
		ExpiresAt:      time.Now().Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("reserve: %v", err)
	}
	if _, err := ledger.MarkInProgress(context.Background(), task.TaskID, "up-sub", 1); err != nil {
		t.Fatalf("mark in progress: %v", err)
	}

	resolver := &testSubscriptionResolver{sub: &SubscriptionBilling{SubscriptionID: subID, IsSubscription: true}}
	provider := &fakeProvider{
		statusFunc: func(task *Task) (*ProviderStatus, error) {
			return &ProviderStatus{State: ProviderSucceeded, Usage: seedanceUsage()}, nil
		},
	}
	loader := testAccountLoader{sel: AccountSelection{AccountID: 1, APIKey: "test-key"}}
	factory := testProviderFactory{p: provider}
	poller := NewPoller(mem, ledger, loader, factory, PollerConfig{SubscriptionResolver: resolver})

	_, errs := poller.RunOnce(context.Background())
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if resolver.calls != 1 {
		t.Fatalf("expected subscription resolver to be called once, got %d", resolver.calls)
	}
	if len(charger.calls) != 1 {
		t.Fatalf("expected 1 charge, got %d", len(charger.calls))
	}
	if !charger.calls[0].IsSubscription || charger.calls[0].SubscriptionID == nil || *charger.calls[0].SubscriptionID != subID {
		t.Fatalf("charge did not carry subscription billing: %#v", charger.calls[0])
	}
}

func TestPoller_UsesTaskSubscriptionSnapshotWithoutResolver(t *testing.T) {
	ledger, mem, charger := newTestLedger(t, 10)
	groupID := int64(9)
	subID := int64(44)
	task, err := ledger.Reserve(context.Background(), ReserveInput{
		TaskID:         "poll-subscription-snapshot",
		UserID:         1,
		APIKeyID:       2,
		GroupID:        &groupID,
		Model:          "doubao-seedance-2.0",
		MediaType:      "video",
		Usage:          seedanceUsage(),
		RateMultiplier: 1,
		SubscriptionID: &subID,
		IsSubscription: true,
		ExpiresAt:      time.Now().Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("reserve: %v", err)
	}
	if _, err := ledger.MarkInProgress(context.Background(), task.TaskID, "up-sub-snapshot", 1); err != nil {
		t.Fatalf("mark in progress: %v", err)
	}

	resolver := &testSubscriptionResolver{sub: nil}
	provider := &fakeProvider{
		statusFunc: func(task *Task) (*ProviderStatus, error) {
			return &ProviderStatus{State: ProviderSucceeded, Usage: seedanceUsage()}, nil
		},
	}
	loader := testAccountLoader{sel: AccountSelection{AccountID: 1, APIKey: "test-key"}}
	factory := testProviderFactory{p: provider}
	poller := NewPoller(mem, ledger, loader, factory, PollerConfig{SubscriptionResolver: resolver})

	_, errs := poller.RunOnce(context.Background())
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if resolver.calls != 0 {
		t.Fatalf("subscription snapshot should avoid live resolver, got %d calls", resolver.calls)
	}
	if len(charger.calls) != 1 {
		t.Fatalf("expected 1 charge, got %d", len(charger.calls))
	}
	if !charger.calls[0].IsSubscription || charger.calls[0].SubscriptionID == nil || *charger.calls[0].SubscriptionID != subID {
		t.Fatalf("charge did not carry snapshot subscription billing: %#v", charger.calls[0])
	}
}

func TestPoller_ReleasesOnFailure(t *testing.T) {
	ledger, mem, charger := newTestLedger(t, 10)
	setupInProgressTask(t, ledger, "poll-fail", time.Now().Add(time.Hour))

	provider := &fakeProvider{
		statusFunc: func(task *Task) (*ProviderStatus, error) {
			return &ProviderStatus{State: ProviderFailed, ErrorMessage: "upstream generation failed"}, nil
		},
	}
	loader := testAccountLoader{sel: AccountSelection{AccountID: 1, APIKey: "test-key"}}
	factory := testProviderFactory{p: provider}
	poller := NewPoller(mem, ledger, loader, factory, PollerConfig{})

	_, errs := poller.RunOnce(context.Background())
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	task, err := mem.GetByTaskID(context.Background(), "poll-fail")
	if err != nil {
		t.Fatalf("get task: %v", err)
	}
	if task.Status != TaskFailed {
		t.Fatalf("expected failed, got %s", task.Status)
	}
	if len(charger.calls) != 0 {
		t.Fatalf("failure should not charge, got %d calls", len(charger.calls))
	}
}

func TestPoller_StillInProgressBumpsAttemptsOnly(t *testing.T) {
	ledger, mem, _ := newTestLedger(t, 10)
	setupInProgressTask(t, ledger, "poll-pending", time.Now().Add(time.Hour))

	provider := &fakeProvider{
		statusFunc: func(task *Task) (*ProviderStatus, error) {
			return &ProviderStatus{State: ProviderInProgress}, nil
		},
	}
	loader := testAccountLoader{sel: AccountSelection{AccountID: 1, APIKey: "test-key"}}
	factory := testProviderFactory{p: provider}
	poller := NewPoller(mem, ledger, loader, factory, PollerConfig{})

	_, errs := poller.RunOnce(context.Background())
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	task, err := mem.GetByTaskID(context.Background(), "poll-pending")
	if err != nil {
		t.Fatalf("get task: %v", err)
	}
	if task.Status != TaskInProgress {
		t.Fatalf("expected still in_progress, got %s", task.Status)
	}
	if task.PollAttempts != 1 {
		t.Fatalf("expected 1 poll attempt recorded, got %d", task.PollAttempts)
	}
}

func TestPoller_ExpiredTaskReleasesEvenWithoutProviderCall(t *testing.T) {
	ledger, mem, charger := newTestLedger(t, 10)
	setupInProgressTask(t, ledger, "poll-expired", time.Now().Add(-time.Minute)) // 已过期

	provider := &fakeProvider{}
	loader := testAccountLoader{sel: AccountSelection{AccountID: 1, APIKey: "test-key"}}
	factory := testProviderFactory{p: provider}
	poller := NewPoller(mem, ledger, loader, factory, PollerConfig{})

	_, errs := poller.RunOnce(context.Background())
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	task, err := mem.GetByTaskID(context.Background(), "poll-expired")
	if err != nil {
		t.Fatalf("get task: %v", err)
	}
	if task.Status != TaskExpired {
		t.Fatalf("expected expired, got %s", task.Status)
	}
	if provider.queryCalls != 0 {
		t.Fatalf("expired task should short-circuit before querying provider, got %d calls", provider.queryCalls)
	}
	if len(charger.calls) != 0 {
		t.Fatalf("expired task should not charge, got %d calls", len(charger.calls))
	}
}

func TestPoller_OrphanPendingTaskWaitsForExpiry(t *testing.T) {
	ledger, mem, _ := newTestLedger(t, 10)
	// Reserve 但从未 MarkInProgress（模拟提交阶段崩溃）。
	_, err := ledger.Reserve(context.Background(), ReserveInput{
		TaskID: "poll-orphan", UserID: 1, APIKeyID: 2,
		Model: "doubao-seedance-2.0", MediaType: "video",
		Usage: seedanceUsage(), RateMultiplier: 1, ExpiresAt: time.Now().Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("reserve: %v", err)
	}

	loader := testAccountLoader{sel: AccountSelection{AccountID: 1, APIKey: "test-key"}}
	factory := testProviderFactory{p: nil}
	poller := NewPoller(mem, ledger, loader, factory, PollerConfig{})

	_, errs := poller.RunOnce(context.Background())
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	task, err := mem.GetByTaskID(context.Background(), "poll-orphan")
	if err != nil {
		t.Fatalf("get task: %v", err)
	}
	if task.Status != TaskPending {
		t.Fatalf("expected still pending (not yet expired), got %s", task.Status)
	}
}

func TestPoller_RespectsBatchSize(t *testing.T) {
	ledger, mem, _ := newTestLedger(t, 100)
	for i := 0; i < 5; i++ {
		setupInProgressTask(t, ledger, "poll-batch-"+string(rune('a'+i)), time.Now().Add(time.Hour))
	}

	provider := &fakeProvider{
		statusFunc: func(task *Task) (*ProviderStatus, error) {
			return &ProviderStatus{State: ProviderInProgress}, nil
		},
	}
	loader := testAccountLoader{sel: AccountSelection{AccountID: 1, APIKey: "test-key"}}
	factory := testProviderFactory{p: provider}
	poller := NewPoller(mem, ledger, loader, factory, PollerConfig{BatchSize: 2})

	processed, errs := poller.RunOnce(context.Background())
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if processed != 2 {
		t.Fatalf("expected batch size of 2, got %d", processed)
	}
}
