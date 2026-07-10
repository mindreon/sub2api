package media

import (
	"context"
	"errors"
	"testing"
	"time"
)

func newTestLedger(t *testing.T, balance float64) (*Ledger, *MemoryStore, *fakeCharger) {
	t.Helper()
	mem := NewMemoryStore()
	holds := mem.MemoryHoldStore()
	conv := NewStaticCurrencyConverter(CurrencyUSD, map[Currency]float64{CurrencyCNY: 0.14})
	quoter := NewQuoter(SeedanceRuleProvider(), conv, nil, CurrencyUSD)
	upstream := BalanceReaderFunc(func(ctx context.Context, uid int64) (float64, error) {
		return balance, nil
	})
	holdAware := NewHoldAwareBalance(upstream, holds)
	charger := &fakeCharger{}
	ledger := NewLedger(quoter, charger, mem, holds, holdAware)
	return ledger, mem, charger
}

type fakeCharger struct {
	calls []ChargeRequest
}

func (f *fakeCharger) Charge(ctx context.Context, req ChargeRequest) (*ChargeResult, error) {
	_ = ctx
	f.calls = append(f.calls, req)
	applied := true
	return &ChargeResult{Applied: applied}, nil
}

func seedanceUsage() BillingUsage {
	return BillingUsage{
		Resolution:         "720p",
		AspectRatio:        "16:9",
		VideoOutputSeconds: 5,
		VideoWidth:         1280,
		VideoHeight:        720,
		VideoFPS:           24,
		HasVideoInput:      false,
	}
}

func TestLedger_ReserveAndSettle(t *testing.T) {
	ledger, _, charger := newTestLedger(t, 10)

	task, err := ledger.Reserve(context.Background(), ReserveInput{
		TaskID:         "task-1",
		UserID:         1,
		APIKeyID:       2,
		AccountID:      3,
		Model:          "doubao-seedance-2.0",
		MediaType:      "video",
		Usage:          seedanceUsage(),
		RateMultiplier: 1,
		ExpiresAt:      time.Now().Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("reserve: %v", err)
	}
	if task.Status != TaskPending {
		t.Fatalf("expected pending, got %s", task.Status)
	}
	if !approxEqual(task.ReservedCost, 0.69552) {
		t.Fatalf("reserved cost mismatch: %f", task.ReservedCost)
	}

	_, err = ledger.MarkInProgress(context.Background(), "task-1", "upstream-abc", 0)
	if err != nil {
		t.Fatalf("mark in progress: %v", err)
	}

	settled, err := ledger.Settle(context.Background(), SettleInput{
		TaskID: "task-1",
		Usage:  seedanceUsage(),
	})
	if err != nil {
		t.Fatalf("settle: %v", err)
	}
	if settled.Status != TaskCompleted {
		t.Fatalf("expected completed, got %s", settled.Status)
	}
	if len(charger.calls) != 1 {
		t.Fatalf("expected 1 charge, got %d", len(charger.calls))
	}
	if !approxEqual(charger.calls[0].ActualCost, 0.69552) {
		t.Fatalf("charge amount mismatch: %f", charger.calls[0].ActualCost)
	}
}

func TestLedger_SettleUsesTaskSubscriptionSnapshot(t *testing.T) {
	ledger, _, charger := newTestLedger(t, 10)
	subscriptionID := int64(44)

	if _, err := ledger.Reserve(context.Background(), ReserveInput{
		TaskID: "task-sub-snapshot", UserID: 1, APIKeyID: 2, AccountID: 3,
		Model: "doubao-seedance-2.0", MediaType: "video",
		Usage: seedanceUsage(), RateMultiplier: 1,
		SubscriptionID: &subscriptionID, IsSubscription: true,
	}); err != nil {
		t.Fatalf("reserve: %v", err)
	}
	if _, err := ledger.MarkInProgress(context.Background(), "task-sub-snapshot", "up-sub-snapshot", 0); err != nil {
		t.Fatalf("mark in progress: %v", err)
	}

	if _, err := ledger.Settle(context.Background(), SettleInput{
		TaskID: "task-sub-snapshot",
		Usage:  seedanceUsage(),
	}); err != nil {
		t.Fatalf("settle: %v", err)
	}

	if len(charger.calls) != 1 {
		t.Fatalf("expected 1 charge, got %d", len(charger.calls))
	}
	if !charger.calls[0].IsSubscription || charger.calls[0].SubscriptionID == nil || *charger.calls[0].SubscriptionID != subscriptionID {
		t.Fatalf("charge should use task subscription snapshot, got %#v", charger.calls[0])
	}
}

// fakeAssetStore 是 AssetStore 的测试替身：可配置 Rehost 成功链接或失败。
type fakeAssetStore struct {
	rehostURL string
	rehostErr error
	lastKey   string
	lastSrc   string
}

func (f *fakeAssetStore) Rehost(ctx context.Context, key, srcURL, contentType string) (string, error) {
	f.lastKey = key
	f.lastSrc = srcURL
	if f.rehostErr != nil {
		return "", f.rehostErr
	}
	return f.rehostURL, nil
}

func (f *fakeAssetStore) PresignedURL(ctx context.Context, key string) (string, error) {
	return "presigned://" + key, nil
}

func settleWithRehost(t *testing.T, assets AssetStore, upstreamURL string) *Task {
	t.Helper()
	ledger, _, _ := newTestLedger(t, 10)
	ledger.WithAssetStore(assets)

	in := ReserveInput{
		TaskID: "task-rehost", UserID: 7, APIKeyID: 2,
		Model: "doubao-seedance-2.0", MediaType: "video", Usage: seedanceUsage(),
		RateMultiplier: 1,
	}
	if _, err := ledger.Reserve(context.Background(), in); err != nil {
		t.Fatalf("reserve: %v", err)
	}
	if _, err := ledger.MarkInProgress(context.Background(), "task-rehost", "up", 0); err != nil {
		t.Fatalf("mark: %v", err)
	}
	task, err := ledger.Settle(context.Background(), SettleInput{
		TaskID:    "task-rehost",
		Usage:     seedanceUsage(),
		ResultURL: upstreamURL,
	})
	if err != nil {
		t.Fatalf("settle: %v", err)
	}
	return task
}

// 转存成功：ResultURL 为存储链接，ResultStorageKey 被写入。
func TestLedger_Settle_RehostsResult(t *testing.T) {
	assets := &fakeAssetStore{rehostURL: "https://cdn.example.com/media/7/task-rehost.mp4"}
	task := settleWithRehost(t, assets, "https://volc.example.com/tmp.mp4")

	if task.ResultURL != assets.rehostURL {
		t.Fatalf("expected rehosted url, got %q", task.ResultURL)
	}
	if task.ResultStorageKey != "media/7/task-rehost.mp4" {
		t.Fatalf("expected storage key, got %q", task.ResultStorageKey)
	}
	if assets.lastSrc != "https://volc.example.com/tmp.mp4" {
		t.Fatalf("rehost should receive upstream url, got %q", assets.lastSrc)
	}
}

// 转存失败：降级为上游直链，仍完成结算（不报错）。
func TestLedger_Settle_RehostFailureFallsBack(t *testing.T) {
	assets := &fakeAssetStore{rehostErr: errors.New("s3 down")}
	task := settleWithRehost(t, assets, "https://volc.example.com/tmp.mp4")

	if task.Status != TaskCompleted {
		t.Fatalf("settle must complete despite rehost failure, got %s", task.Status)
	}
	if task.ResultURL != "https://volc.example.com/tmp.mp4" {
		t.Fatalf("expected fallback to upstream url, got %q", task.ResultURL)
	}
	if task.ResultStorageKey != "" {
		t.Fatalf("storage key must be empty on rehost failure, got %q", task.ResultStorageKey)
	}
}

// 未配置存储：保留上游直链。
func TestLedger_Settle_NoAssetStoreKeepsUpstream(t *testing.T) {
	task := settleWithRehost(t, nil, "https://volc.example.com/tmp.mp4")
	if task.ResultURL != "https://volc.example.com/tmp.mp4" {
		t.Fatalf("expected upstream url when no asset store, got %q", task.ResultURL)
	}
}

func TestLedger_ReserveInsufficientBalance(t *testing.T) {
	ledger, _, _ := newTestLedger(t, 0.01)
	_, err := ledger.Reserve(context.Background(), ReserveInput{
		TaskID:         "task-poor",
		UserID:         1,
		APIKeyID:       2,
		Model:          "doubao-seedance-2.0",
		MediaType:      "video",
		Usage:          seedanceUsage(),
		RateMultiplier: 1,
	})
	if !errors.Is(err, ErrInsufficientBalance) {
		t.Fatalf("expected ErrInsufficientBalance, got %v", err)
	}
}

func TestLedger_ReleaseDoesNotCharge(t *testing.T) {
	ledger, _, charger := newTestLedger(t, 10)

	_, err := ledger.Reserve(context.Background(), ReserveInput{
		TaskID:         "task-fail",
		UserID:         1,
		APIKeyID:       2,
		Model:          "doubao-seedance-2.0",
		MediaType:      "video",
		Usage:          seedanceUsage(),
		RateMultiplier: 1,
	})
	if err != nil {
		t.Fatalf("reserve: %v", err)
	}

	_, err = ledger.MarkInProgress(context.Background(), "task-fail", "up-1", 0)
	if err != nil {
		t.Fatalf("mark in progress: %v", err)
	}

	released, err := ledger.Release(context.Background(), ReleaseInput{
		TaskID:       "task-fail",
		Status:       TaskFailed,
		ErrorMessage: "audit rejected",
	})
	if err != nil {
		t.Fatalf("release: %v", err)
	}
	if released.Status != TaskFailed {
		t.Fatalf("expected failed, got %s", released.Status)
	}
	if len(charger.calls) != 0 {
		t.Fatalf("release should not charge, got %d calls", len(charger.calls))
	}
}

// 严格预扣事务路径：用 rawBalance（未扣预扣）+ 内存 Reservation 原子建库。
func newReservationLedger(t *testing.T, rawBalance float64) (*Ledger, *MemoryStore) {
	t.Helper()
	mem := NewMemoryStore()
	holds := mem.MemoryHoldStore()
	conv := NewStaticCurrencyConverter(CurrencyUSD, map[Currency]float64{CurrencyCNY: 0.14})
	quoter := NewQuoter(SeedanceRuleProvider(), conv, nil, CurrencyUSD)
	raw := BalanceReaderFunc(func(ctx context.Context, uid int64) (float64, error) { return rawBalance, nil })
	ledger := NewLedger(quoter, &fakeCharger{}, mem, holds, NewHoldAwareBalance(raw, holds)).
		WithReservation(mem.MemoryReservation(), raw)
	return ledger, mem
}

// 严格路径：余额只够一次预扣时，第二次原子预扣因 held 重算被拒，且不残留 task/hold。
func TestLedger_StrictReservation_BlocksSecond(t *testing.T) {
	ledger, mem := newReservationLedger(t, 1) // 单次约 0.69552，两次需 ~1.39

	if _, err := ledger.Reserve(context.Background(), ReserveInput{
		TaskID: "s-a", UserID: 1, APIKeyID: 2,
		Model: "doubao-seedance-2.0", MediaType: "video", Usage: seedanceUsage(), RateMultiplier: 1,
	}); err != nil {
		t.Fatalf("first reserve: %v", err)
	}
	_, err := ledger.Reserve(context.Background(), ReserveInput{
		TaskID: "s-b", UserID: 1, APIKeyID: 2,
		Model: "doubao-seedance-2.0", MediaType: "video", Usage: seedanceUsage(), RateMultiplier: 1,
	})
	if !errors.Is(err, ErrInsufficientBalance) {
		t.Fatalf("second reserve should be blocked, got %v", err)
	}
	// 被拒的第二次不应残留任务或预扣。
	if _, err := mem.GetByTaskID(context.Background(), "s-b"); !errors.Is(err, ErrTaskNotFound) {
		t.Fatalf("rejected reserve must not persist task, got %v", err)
	}
	if _, err := mem.MemoryHoldStore().GetByTaskID(context.Background(), "s-b"); !errors.Is(err, ErrHoldNotFound) {
		t.Fatalf("rejected reserve must not persist hold, got %v", err)
	}
}

// 严格路径：余额充足时预扣成功，可继续 Settle。
func TestLedger_StrictReservation_Succeeds(t *testing.T) {
	ledger, _ := newReservationLedger(t, 10)
	task, err := ledger.Reserve(context.Background(), ReserveInput{
		TaskID: "s-ok", UserID: 1, APIKeyID: 2,
		Model: "doubao-seedance-2.0", MediaType: "video", Usage: seedanceUsage(), RateMultiplier: 1,
	})
	if err != nil {
		t.Fatalf("reserve: %v", err)
	}
	if task.Status != TaskPending || !approxEqual(task.ReservedCost, 0.69552) {
		t.Fatalf("unexpected reserved task: %+v", task)
	}
}

func TestHoldAwareBalance_SubtractsHeld(t *testing.T) {
	mem := NewMemoryStore()
	holds := mem.MemoryHoldStore()
	upstream := BalanceReaderFunc(func(ctx context.Context, uid int64) (float64, error) { return 10, nil })
	aware := NewHoldAwareBalance(upstream, holds)

	_ = holds.Create(context.Background(), &Hold{
		HoldID: "h1", TaskID: "t1", UserID: 1, Amount: 3, Currency: CurrencyUSD, Status: HoldHeld,
	})

	got, err := aware.AvailableBalance(context.Background(), 1)
	if err != nil || !approxEqual(got, 7) {
		t.Fatalf("expected 7 available, got %f err %v", got, err)
	}
}

func TestLedger_ReserveIdempotent(t *testing.T) {
	ledger, _, _ := newTestLedger(t, 10)
	in := ReserveInput{
		TaskID: "task-dup", UserID: 1, APIKeyID: 2,
		Model: "doubao-seedance-2.0", MediaType: "video", Usage: seedanceUsage(),
		RateMultiplier: 1,
	}
	first, err := ledger.Reserve(context.Background(), in)
	if err != nil {
		t.Fatalf("first reserve: %v", err)
	}
	second, err := ledger.Reserve(context.Background(), in)
	if err != nil {
		t.Fatalf("second reserve: %v", err)
	}
	if first.TaskID != second.TaskID || first.ReservedCost != second.ReservedCost {
		t.Fatalf("idempotent reserve should return same task")
	}
}

func TestLedger_SettleIdempotent(t *testing.T) {
	ledger, _, charger := newTestLedger(t, 10)
	in := ReserveInput{
		TaskID: "task-settle-dup", UserID: 1, APIKeyID: 2,
		Model: "doubao-seedance-2.0", MediaType: "video", Usage: seedanceUsage(),
		RateMultiplier: 1,
	}
	_, _ = ledger.Reserve(context.Background(), in)
	_, _ = ledger.MarkInProgress(context.Background(), "task-settle-dup", "up", 0)

	usage := seedanceUsage()
	_, err := ledger.Settle(context.Background(), SettleInput{TaskID: "task-settle-dup", Usage: usage})
	if err != nil {
		t.Fatalf("first settle: %v", err)
	}
	_, err = ledger.Settle(context.Background(), SettleInput{TaskID: "task-settle-dup", Usage: usage})
	if err != nil {
		t.Fatalf("second settle: %v", err)
	}
	if len(charger.calls) != 1 {
		t.Fatalf("idempotent settle should charge once, got %d", len(charger.calls))
	}
}

func TestLedger_SecondReserveBlockedByHold(t *testing.T) {
	ledger, _, _ := newTestLedger(t, 1) // 只够一次预扣（~0.69552 需要更多）

	_, err := ledger.Reserve(context.Background(), ReserveInput{
		TaskID: "t-a", UserID: 1, APIKeyID: 2,
		Model: "doubao-seedance-2.0", MediaType: "video", Usage: seedanceUsage(),
		RateMultiplier: 1,
	})
	if err != nil {
		t.Fatalf("first reserve: %v", err)
	}
	_, err = ledger.Reserve(context.Background(), ReserveInput{
		TaskID: "t-b", UserID: 1, APIKeyID: 2,
		Model: "doubao-seedance-2.0", MediaType: "video", Usage: seedanceUsage(),
		RateMultiplier: 1,
	})
	if !errors.Is(err, ErrInsufficientBalance) {
		t.Fatalf("second reserve should fail due to hold, got %v", err)
	}
}
