package media

import (
	"context"
	"errors"
	"testing"
)

type fakeProvider struct {
	submitErr   error
	upstreamID  string
	statusFunc  func(task *Task) (*ProviderStatus, error)
	submitCalls int
	queryCalls  int
}

func (f *fakeProvider) Submit(ctx context.Context, task *Task) (string, error) {
	f.submitCalls++
	if f.submitErr != nil {
		return "", f.submitErr
	}
	return f.upstreamID, nil
}

func (f *fakeProvider) QueryStatus(ctx context.Context, task *Task) (*ProviderStatus, error) {
	f.queryCalls++
	if f.statusFunc != nil {
		return f.statusFunc(task)
	}
	return &ProviderStatus{State: ProviderInProgress}, nil
}

func newTestTaskService(t *testing.T, balance float64, provider Provider) (*TaskService, *MemoryStore, *fakeCharger) {
	t.Helper()
	ledger, mem, charger := newTestLedger(t, balance)
	loader := testAccountLoader{sel: AccountSelection{AccountID: 1, APIKey: "test-key", Platform: "volcengine"}}
	factory := testProviderFactory{p: provider}
	svc := NewTaskService(ledger, nil, loader, factory, nil)
	return svc, mem, charger
}

type testAccountLoader struct {
	sel AccountSelection
}

func (l testAccountLoader) Load(ctx context.Context, accountID int64) (AccountSelection, error) {
	return l.sel, nil
}

type testProviderFactory struct {
	p Provider
}

func (f testProviderFactory) NewProvider(sel AccountSelection, model string) (Provider, error) {
	if f.p == nil {
		return nil, ErrProviderNotFound
	}
	return f.p, nil
}

func TestTaskService_Submit_Success(t *testing.T) {
	provider := &fakeProvider{upstreamID: "up-123"}
	svc, _, _ := newTestTaskService(t, 10, provider)

	task, err := svc.Submit(context.Background(), ReserveInput{
		TaskID: "svc-task-1", UserID: 1, APIKeyID: 2, AccountID: 1,
		Model: "doubao-seedance-2.0", MediaType: "video",
		Usage: seedanceUsage(), RateMultiplier: 1,
	})
	if err != nil {
		t.Fatalf("submit: %v", err)
	}
	if task.Status != TaskInProgress {
		t.Fatalf("expected in_progress, got %s", task.Status)
	}
	if task.UpstreamTaskID != "up-123" {
		t.Fatalf("expected upstream task id set, got %q", task.UpstreamTaskID)
	}
	if provider.submitCalls != 1 {
		t.Fatalf("expected 1 submit call, got %d", provider.submitCalls)
	}
}

func TestTaskService_SubmitSnapshotsSubscription(t *testing.T) {
	ledger, mem, _ := newTestLedger(t, 10)
	provider := &fakeProvider{upstreamID: "up-sub-snapshot"}
	loader := testAccountLoader{sel: AccountSelection{AccountID: 1, APIKey: "test-key"}}
	factory := testProviderFactory{p: provider}
	groupID := int64(9)
	subscriptionID := int64(44)
	resolver := SubscriptionResolverFunc(func(ctx context.Context, userID int64, groupID *int64) (*SubscriptionBilling, error) {
		return &SubscriptionBilling{SubscriptionID: subscriptionID, IsSubscription: true}, nil
	})
	svc := NewTaskService(ledger, nil, loader, factory, resolver)

	task, err := svc.Submit(context.Background(), ReserveInput{
		TaskID: "svc-task-subscription", UserID: 1, APIKeyID: 2, AccountID: 1, GroupID: &groupID,
		Model: "doubao-seedance-2.0", MediaType: "video",
		Usage: seedanceUsage(), RateMultiplier: 1,
	})
	if err != nil {
		t.Fatalf("submit: %v", err)
	}
	if task.SubscriptionID == nil || *task.SubscriptionID != subscriptionID {
		t.Fatalf("returned task should snapshot subscription id, got %+v", task.SubscriptionID)
	}
	stored, err := mem.GetByTaskID(context.Background(), "svc-task-subscription")
	if err != nil {
		t.Fatalf("get task: %v", err)
	}
	if stored.SubscriptionID == nil || *stored.SubscriptionID != subscriptionID {
		t.Fatalf("stored task should snapshot subscription id, got %+v", stored.SubscriptionID)
	}
}

func TestTaskService_Submit_NoProviderReleasesHold(t *testing.T) {
	ledger, mem, _ := newTestLedger(t, 10)
	loader := testAccountLoader{sel: AccountSelection{AccountID: 1, APIKey: "test-key"}}
	factory := testProviderFactory{p: nil}
	svc := NewTaskService(ledger, nil, loader, factory, nil)

	_, err := svc.Submit(context.Background(), ReserveInput{
		TaskID: "svc-task-no-provider", UserID: 1, APIKeyID: 2, AccountID: 1,
		Model: "doubao-seedance-2.0", MediaType: "video",
		Usage: seedanceUsage(), RateMultiplier: 1,
	})
	if !errors.Is(err, ErrProviderNotFound) {
		t.Fatalf("expected ErrProviderNotFound, got %v", err)
	}

	task, getErr := mem.GetByTaskID(context.Background(), "svc-task-no-provider")
	if getErr != nil {
		t.Fatalf("get task: %v", getErr)
	}
	if task.Status != TaskFailed {
		t.Fatalf("expected task released as failed, got %s", task.Status)
	}

	holds := mem.MemoryHoldStore()
	hold, holdErr := holds.GetByTaskID(context.Background(), "svc-task-no-provider")
	if holdErr != nil {
		t.Fatalf("get hold: %v", holdErr)
	}
	if hold.Status != HoldReleased {
		t.Fatalf("expected hold released, got %s", hold.Status)
	}
}

func TestTaskService_Submit_ProviderSubmitFailsReleasesHold(t *testing.T) {
	provider := &fakeProvider{submitErr: errors.New("upstream 500")}
	svc, mem, _ := newTestTaskService(t, 10, provider)

	_, err := svc.Submit(context.Background(), ReserveInput{
		TaskID: "svc-task-submit-fail", UserID: 1, APIKeyID: 2, AccountID: 1,
		Model: "doubao-seedance-2.0", MediaType: "video",
		Usage: seedanceUsage(), RateMultiplier: 1,
	})
	if err == nil {
		t.Fatal("expected error from failed submit")
	}

	task, getErr := mem.GetByTaskID(context.Background(), "svc-task-submit-fail")
	if getErr != nil {
		t.Fatalf("get task: %v", getErr)
	}
	if task.Status != TaskFailed {
		t.Fatalf("expected failed, got %s", task.Status)
	}
}

func TestTaskService_Submit_Idempotent(t *testing.T) {
	provider := &fakeProvider{upstreamID: "up-once"}
	svc, _, _ := newTestTaskService(t, 10, provider)

	in := ReserveInput{
		TaskID: "svc-task-dup", UserID: 1, APIKeyID: 2, AccountID: 1,
		Model: "doubao-seedance-2.0", MediaType: "video",
		Usage: seedanceUsage(), RateMultiplier: 1,
	}
	first, err := svc.Submit(context.Background(), in)
	if err != nil {
		t.Fatalf("first submit: %v", err)
	}
	second, err := svc.Submit(context.Background(), in)
	if err != nil {
		t.Fatalf("second submit: %v", err)
	}
	if first.TaskID != second.TaskID || second.UpstreamTaskID != "up-once" {
		t.Fatalf("expected idempotent result, got %+v vs %+v", first, second)
	}
	if provider.submitCalls != 1 {
		t.Fatalf("expected upstream submit called exactly once, got %d", provider.submitCalls)
	}
}
