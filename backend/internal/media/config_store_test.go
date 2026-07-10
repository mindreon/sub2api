package media

import (
	"context"
	"testing"
)

type mockSettingsKV struct {
	values map[string]string
}

func (m *mockSettingsKV) GetValue(_ context.Context, key string) (string, error) {
	if m.values == nil {
		return "", nil
	}
	return m.values[key], nil
}

func (m *mockSettingsKV) Set(_ context.Context, key, value string) error {
	if m.values == nil {
		m.values = map[string]string{}
	}
	m.values[key] = value
	return nil
}

func TestConfigStore_LoadFromDedicatedKey(t *testing.T) {
	kv := &mockSettingsKV{values: map[string]string{
		MediaCNYToUSDRateSettingKey: "0.15",
	}}
	store := NewConfigStore(kv, BillingConfig{CNYToUSDRate: 0.14})

	cfg, err := store.Load(context.Background())
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.CNYToUSDRate != 0.15 {
		t.Fatalf("expected 0.15, got %v", cfg.CNYToUSDRate)
	}
}

func TestConfigStore_LoadFromLegacyJSON(t *testing.T) {
	kv := &mockSettingsKV{values: map[string]string{
		legacyBillingConfigSettingKey: `{"cny_to_usd_rate":0.16}`,
	}}
	store := NewConfigStore(kv, BillingConfig{CNYToUSDRate: 0.14})

	cfg, err := store.Load(context.Background())
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.CNYToUSDRate != 0.16 {
		t.Fatalf("expected legacy rate 0.16, got %v", cfg.CNYToUSDRate)
	}
}

func TestConfigStore_SaveWritesDedicatedKey(t *testing.T) {
	kv := &mockSettingsKV{values: map[string]string{}}
	store := NewConfigStore(kv, BillingConfig{CNYToUSDRate: 0.14})

	if err := store.Save(context.Background(), BillingConfig{CNYToUSDRate: 0.17}); err != nil {
		t.Fatalf("save: %v", err)
	}
	if kv.values[MediaCNYToUSDRateSettingKey] != "0.17" {
		t.Fatalf("expected dedicated key written, got %q", kv.values[MediaCNYToUSDRateSettingKey])
	}
}
