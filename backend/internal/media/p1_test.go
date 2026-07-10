package media

import (
	"context"
	"errors"
	"testing"
)

// --- 货币换算 ---

func TestStaticCurrencyConverter(t *testing.T) {
	conv := NewStaticCurrencyConverter(CurrencyUSD, map[Currency]float64{CurrencyCNY: 0.14})

	if got, err := conv.Convert(10, CurrencyUSD); err != nil || !approxEqual(got, 10) {
		t.Fatalf("same currency should be identity: got %f err %v", got, err)
	}
	if got, err := conv.Convert(100, CurrencyCNY); err != nil || !approxEqual(got, 14) {
		t.Fatalf("CNY->USD mismatch: got %f err %v", got, err)
	}
	if got, err := conv.Convert(5, ""); err != nil || !approxEqual(got, 5) {
		t.Fatalf("empty currency should be identity: got %f err %v", got, err)
	}
	if _, err := conv.Convert(1, Currency("JPY")); !errors.Is(err, ErrCurrencyRateMissing) {
		t.Fatalf("expected ErrCurrencyRateMissing, got %v", err)
	}
}

// --- Quoter 预估 + 换算 ---

func TestQuoter_Estimate_SeedanceConvertedToUSD(t *testing.T) {
	conv := NewStaticCurrencyConverter(CurrencyUSD, map[Currency]float64{CurrencyCNY: 0.14})
	q := NewQuoter(SeedanceRuleProvider(), conv, nil, CurrencyUSD)

	quote, err := q.Estimate("doubao-seedance-2.0", BillingUsage{
		Resolution:         "720p",
		AspectRatio:        "16:9",
		VideoOutputSeconds: 5,
		VideoWidth:         1280,
		VideoHeight:        720,
		VideoFPS:           24,
		HasVideoInput:      false,
	}, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 原始 4.968 CNY → 0.14 → 0.69552 USD
	if !approxEqual(quote.Cost.TotalCost, 4.968) {
		t.Fatalf("expected CNY total 4.968, got %f", quote.Cost.TotalCost)
	}
	if !approxEqual(quote.TotalBillingCost, 0.69552) {
		t.Fatalf("expected USD total 0.69552, got %f", quote.TotalBillingCost)
	}
	if quote.BillingCurrency != CurrencyUSD {
		t.Fatalf("expected USD billing currency, got %s", quote.BillingCurrency)
	}
}

func TestQuoter_Estimate_NotPriceable(t *testing.T) {
	conv := NewStaticCurrencyConverter(CurrencyUSD, nil)
	q := NewQuoter(SeedanceRuleProvider(), conv, nil, CurrencyUSD)
	if _, err := q.Estimate("gpt-5.1", BillingUsage{}, 1); !errors.Is(err, ErrModelNotPriceable) {
		t.Fatalf("expected ErrModelNotPriceable, got %v", err)
	}
}

func TestQuoter_Estimate_MissingRatePropagates(t *testing.T) {
	conv := NewStaticCurrencyConverter(CurrencyUSD, nil) // 无 CNY 汇率
	q := NewQuoter(SeedanceRuleProvider(), conv, nil, CurrencyUSD)
	_, err := q.Estimate("doubao-seedance-2.0", BillingUsage{
		Resolution: "720p", VideoTokens: 100000, HasVideoInput: false,
	}, 1)
	if !errors.Is(err, ErrCurrencyRateMissing) {
		t.Fatalf("expected ErrCurrencyRateMissing, got %v", err)
	}
}

// --- 余额检查 ---

func TestQuoter_EnsureAffordable(t *testing.T) {
	conv := NewStaticCurrencyConverter(CurrencyUSD, nil)

	rich := BalanceReaderFunc(func(ctx context.Context, uid int64) (float64, error) { return 100, nil })
	poor := BalanceReaderFunc(func(ctx context.Context, uid int64) (float64, error) { return 0.5, nil })

	if err := NewQuoter(nil, conv, rich, CurrencyUSD).EnsureAffordable(context.Background(), 1, 10); err != nil {
		t.Fatalf("rich user should afford: %v", err)
	}
	if err := NewQuoter(nil, conv, poor, CurrencyUSD).EnsureAffordable(context.Background(), 1, 10); !errors.Is(err, ErrInsufficientBalance) {
		t.Fatalf("expected ErrInsufficientBalance, got %v", err)
	}
	// balance 端口未配置时跳过检查
	if err := NewQuoter(nil, conv, nil, CurrencyUSD).EnsureAffordable(context.Background(), 1, 10); err != nil {
		t.Fatalf("nil balance should skip check: %v", err)
	}
}

// --- 结构化适配器（模拟上游 *BillingCacheService）---

type fakeUpstreamBalance struct {
	balance float64
	err     error
}

func (f fakeUpstreamBalance) GetUserBalance(ctx context.Context, userID int64) (float64, error) {
	return f.balance, f.err
}

func TestBalanceReaderAdapter(t *testing.T) {
	reader := NewBalanceReader(fakeUpstreamBalance{balance: 42})
	got, err := reader.AvailableBalance(context.Background(), 1)
	if err != nil || !approxEqual(got, 42) {
		t.Fatalf("expected 42, got %f err %v", got, err)
	}

	nilReader := NewBalanceReader(nil)
	if _, err := nilReader.AvailableBalance(context.Background(), 1); !errors.Is(err, ErrUpstreamNotWired) {
		t.Fatalf("expected ErrUpstreamNotWired, got %v", err)
	}
}

// ChargerFunc 满足 Charger 契约。
func TestChargerFunc_SatisfiesInterface(t *testing.T) {
	var called bool
	var charger Charger = ChargerFunc(func(ctx context.Context, req ChargeRequest) (*ChargeResult, error) {
		called = true
		applied := true
		return &ChargeResult{Applied: applied}, nil
	})
	res, err := charger.Charge(context.Background(), ChargeRequest{RequestID: "t1"})
	if err != nil || res == nil || !res.Applied || !called {
		t.Fatalf("charger func not invoked correctly: res=%v err=%v called=%v", res, err, called)
	}
}
