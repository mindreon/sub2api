package media

import "testing"

func TestRouteModel(t *testing.T) {
	cases := []struct {
		model   string
		backend Backend
		ok      bool
	}{
		{"doubao-seedance-2.0", BackendVolcengine, true},
		{"seedance-2.0-fast", BackendVolcengine, true},
		{"bytedance/seedance-2.0", BackendOpenRouter, true},
		{"gpt-5.1", "", false},
	}
	for _, tc := range cases {
		got, ok := RouteModel(tc.model)
		if ok != tc.ok || got != tc.backend {
			t.Fatalf("RouteModel(%q) = (%q, %v), want (%q, %v)", tc.model, got, ok, tc.backend, tc.ok)
		}
	}
}

func TestOpenRouterSeedancePricingRules(t *testing.T) {
	rules := OpenRouterSeedancePricingRules("bytedance/seedance-2.0")
	if len(rules) != 1 || rules[0].Currency != CurrencyUSD {
		t.Fatalf("unexpected rules: %+v", rules)
	}
	cost, err := CalculateMediaCost(MediaCostInput{
		Rules: rules,
		Usage: BillingUsage{VideoTokens: 1_000_000},
	})
	if err != nil || cost.TotalCost != 7 {
		t.Fatalf("expected $7 for 1M tokens, got %f err=%v", cost.TotalCost, err)
	}
}

func TestSeedanceRuleProvider_OpenRouterModelUsesUSD(t *testing.T) {
	q := NewQuoter(SeedanceRuleProvider(), NewStaticCurrencyConverter(CurrencyUSD, nil), nil, CurrencyUSD)
	quote, err := q.Estimate("bytedance/seedance-2.0", BillingUsage{VideoTokens: 1_000_000}, 1)
	if err != nil {
		t.Fatalf("estimate: %v", err)
	}
	if quote.Cost.Currency != CurrencyUSD || quote.TotalBillingCost != 7 {
		t.Fatalf("expected $7 USD, got currency=%s total=%f", quote.Cost.Currency, quote.TotalBillingCost)
	}
}
