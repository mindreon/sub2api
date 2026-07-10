package media

import (
	"context"
	"testing"
)

func TestConfigStore_LoadPricingOverrides(t *testing.T) {
	kv := &mockSettingsKV{values: map[string]string{
		MediaPricingOverridesSettingKey: `[{"model":"custom-seedance","metric":"video_token","price_per_million":99,"currency":"CNY","resolutions":["720p"],"has_video_input":false}]`,
	}}
	store := NewConfigStore(kv, BillingConfig{CNYToUSDRate: 0.14})

	cfg, err := store.Load(context.Background())
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(cfg.PricingOverrides) != 1 {
		t.Fatalf("expected 1 pricing override, got %d", len(cfg.PricingOverrides))
	}
	override := cfg.PricingOverrides[0]
	if override.Model != "custom-seedance" || override.PricePerMillion != 99 || override.Currency != CurrencyCNY {
		t.Fatalf("unexpected override: %#v", override)
	}
}

func TestConfigBackedRuleProviderUsesOverrideBeforeFallback(t *testing.T) {
	kv := &mockSettingsKV{values: map[string]string{
		MediaPricingOverridesSettingKey: `[{"model":"doubao-seedance-2.0","metric":"video_token","price_per_million":99,"currency":"CNY","resolutions":["720p"],"has_video_input":false}]`,
	}}
	store := NewConfigStore(kv, BillingConfig{CNYToUSDRate: 0.14})
	provider := NewConfigBackedRuleProvider(store, SeedanceRuleProvider())

	rules := provider.RulesFor("doubao-seedance-2.0")
	if len(rules) != 1 {
		t.Fatalf("expected configured override only, got %d rules", len(rules))
	}
	if rules[0].UnitPrice != 99.0/1_000_000 || rules[0].Currency != CurrencyCNY {
		t.Fatalf("unexpected rule: %#v", rules[0])
	}
	if rules := provider.RulesFor("doubao-seedance-2.0-fast"); len(rules) == 0 {
		t.Fatal("fallback rules should still be available")
	}
}

func TestConfigStore_SavePricingOverrides(t *testing.T) {
	kv := &mockSettingsKV{values: map[string]string{}}
	store := NewConfigStore(kv, BillingConfig{CNYToUSDRate: 0.14})

	err := store.Save(context.Background(), BillingConfig{
		CNYToUSDRate: 0.17,
		PricingOverrides: []PricingOverride{{
			Model:           "custom-seedance",
			Metric:          MetricVideoToken,
			PricePerMillion: 7,
			Currency:        CurrencyUSD,
		}},
	})
	if err != nil {
		t.Fatalf("save: %v", err)
	}
	if kv.values[MediaPricingOverridesSettingKey] == "" {
		t.Fatal("expected pricing override key to be written")
	}
}
