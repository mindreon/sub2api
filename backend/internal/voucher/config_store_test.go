package voucher

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

type mockSettingsRepo struct {
	values map[string]string
}

func (m *mockSettingsRepo) GetValue(_ context.Context, key string) (string, error) {
	v, ok := m.values[key]
	if !ok {
		return "", service.ErrSettingNotFound
	}
	return v, nil
}

func (m *mockSettingsRepo) Set(_ context.Context, key, value string) error {
	if m.values == nil {
		m.values = map[string]string{}
	}
	m.values[key] = value
	return nil
}

func TestConfigStore_DefaultWhenMissing(t *testing.T) {
	store := NewConfigStore(&mockSettingsRepo{})
	cfg, runtime, err := store.Load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Enabled {
		t.Fatalf("expected disabled by default")
	}
	if !runtime.UIEnabled {
		t.Fatalf("expected ui enabled by default")
	}
	if runtime.OrderTimeoutHours != 24 {
		t.Fatalf("unexpected timeout: %d", runtime.OrderTimeoutHours)
	}
}

func TestConfigStore_UpdateAndLoad(t *testing.T) {
	repo := &mockSettingsRepo{}
	store := NewConfigStore(repo)
	secret := "top-secret"
	key := "kvm_test_abc123"
	enabled := true
	err := store.Update(context.Background(), UpdateSettingsInput{
		Enabled:   &enabled,
		APIKey:    &key,
		APISecret: &secret,
		BankAccounts: []BankAccount{{
			ID: 1, BankName: "Maybank", AccountName: "Demo", AccountNumber: "123",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	cfg, runtime, err := store.Load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Enabled || !cfg.Sandbox {
		t.Fatalf("expected enabled sandbox config")
	}
	if len(runtime.BankAccounts) != 1 {
		t.Fatalf("expected bank account")
	}
	view, err := store.AdminView(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if view.SecretConfigured != true || view.APIKeyMasked == "" {
		t.Fatalf("expected masked admin view")
	}
	raw := repo.values[configSettingKey]
	var p persistedConfig
	if err := json.Unmarshal([]byte(raw), &p); err != nil {
		t.Fatal(err)
	}
	if p.APISecret != secret {
		t.Fatalf("secret not persisted")
	}
}

func TestConfigStore_RejectEnableWithoutCredentials(t *testing.T) {
	store := NewConfigStore(&mockSettingsRepo{})
	enabled := true
	err := store.Update(context.Background(), UpdateSettingsInput{Enabled: &enabled})
	if err == nil {
		t.Fatal("expected error when enabling without credentials")
	}
}

func TestMaskAPIKey(t *testing.T) {
	masked := MaskAPIKey("kvm_live_abcdefghijklmnop")
	if masked == "" || masked == "kvm_live_abcdefghijklmnop" {
		t.Fatalf("unexpected mask: %q", masked)
	}
}
