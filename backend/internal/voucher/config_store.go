package voucher

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

const configSettingKey = "kvoucher_config"

type settingsRepo interface {
	GetValue(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string) error
}

// ConfigStore persists KVoucher integration settings in the settings table.
type ConfigStore struct {
	repo settingsRepo
}

func NewConfigStore(repo settingsRepo) *ConfigStore {
	return &ConfigStore{repo: repo}
}

type persistedConfig struct {
	Enabled             bool          `json:"enabled"`
	UIEnabled           bool          `json:"ui_enabled"`
	APIKey              string        `json:"api_key"`
	APISecret           string        `json:"api_secret"`
	APIBase             string        `json:"api_base"`
	Sandbox             bool          `json:"sandbox"`
	BankAccounts        []BankAccount `json:"bank_accounts"`
	OrderTimeoutHours   int           `json:"order_timeout_hours"`
	MaxQuantityPerOrder int           `json:"max_quantity_per_order"`
	ReviewSLAHours      int           `json:"review_sla_hours"`
	FeeRate             float64       `json:"fee_rate"`
	HelpText            string        `json:"help_text"`
	RetailMarkupPercent float64       `json:"retail_markup_percent"`
}

// AdminSettingsView is returned to the admin UI (no API secret plaintext).
type AdminSettingsView struct {
	Enabled             bool          `json:"enabled"`
	UIEnabled           bool          `json:"ui_enabled"`
	Sandbox             bool          `json:"sandbox"`
	SandboxFromKey      bool          `json:"sandbox_from_key"`
	APIBase             string        `json:"api_base"`
	APIKeyMasked        string        `json:"api_key_masked"`
	SecretConfigured    bool          `json:"secret_configured"`
	BankAccounts        []BankAccount `json:"bank_accounts"`
	OrderTimeoutHours   int           `json:"order_timeout_hours"`
	MaxQuantityPerOrder int           `json:"max_quantity_per_order"`
	ReviewSLAHours      int           `json:"review_sla_hours"`
	FeeRate             float64       `json:"fee_rate"`
	HelpText            string        `json:"help_text"`
	RetailMarkupPercent float64       `json:"retail_markup_percent"`
}

// UpdateSettingsInput carries partial admin updates.
type UpdateSettingsInput struct {
	Enabled             *bool
	UIEnabled           *bool
	APIKey              *string
	APISecret           *string
	APIBase             *string
	Sandbox             *bool
	BankAccounts        []BankAccount
	OrderTimeoutHours   *int
	MaxQuantityPerOrder *int
	ReviewSLAHours      *int
	FeeRate             *float64
	HelpText            *string
	RetailMarkupPercent *float64
}

func (s *ConfigStore) Load(ctx context.Context) (Config, RuntimeSettings, error) {
	p, err := s.readPersisted(ctx)
	if err != nil {
		return Config{}, RuntimeSettings{}, err
	}
	return p.toConfig(), p.toRuntime(), nil
}

func (s *ConfigStore) AdminView(ctx context.Context) (AdminSettingsView, error) {
	p, err := s.readPersisted(ctx)
	if err != nil {
		return AdminSettingsView{}, err
	}
	key := strings.TrimSpace(p.APIKey)
	sandbox := p.effectiveSandbox()
	return AdminSettingsView{
		Enabled:             p.Enabled,
		UIEnabled:           p.UIEnabled,
		Sandbox:             sandbox,
		SandboxFromKey:      strings.HasPrefix(key, "kvm_test_"),
		APIBase:             p.apiBase(),
		APIKeyMasked:        MaskAPIKey(key),
		SecretConfigured:    strings.TrimSpace(p.APISecret) != "",
		BankAccounts:        append([]BankAccount(nil), p.BankAccounts...),
		OrderTimeoutHours:   p.OrderTimeoutHours,
		MaxQuantityPerOrder: p.MaxQuantityPerOrder,
		ReviewSLAHours:      p.ReviewSLAHours,
		FeeRate:             p.FeeRate,
		HelpText:            p.HelpText,
		RetailMarkupPercent: p.RetailMarkupPercent,
	}, nil
}

func (s *ConfigStore) Update(ctx context.Context, in UpdateSettingsInput) error {
	p, err := s.readPersisted(ctx)
	if err != nil {
		return err
	}

	if in.Enabled != nil {
		p.Enabled = *in.Enabled
	}
	if in.UIEnabled != nil {
		p.UIEnabled = *in.UIEnabled
	}
	if in.APIKey != nil {
		p.APIKey = strings.TrimSpace(*in.APIKey)
	}
	if in.APISecret != nil && strings.TrimSpace(*in.APISecret) != "" {
		p.APISecret = strings.TrimSpace(*in.APISecret)
	}
	if in.APIBase != nil {
		p.APIBase = strings.TrimRight(strings.TrimSpace(*in.APIBase), "/")
	}
	if in.Sandbox != nil {
		p.Sandbox = *in.Sandbox
	}
	if in.BankAccounts != nil {
		p.BankAccounts = in.BankAccounts
	}
	if in.OrderTimeoutHours != nil && *in.OrderTimeoutHours > 0 {
		p.OrderTimeoutHours = *in.OrderTimeoutHours
	}
	if in.MaxQuantityPerOrder != nil && *in.MaxQuantityPerOrder > 0 {
		p.MaxQuantityPerOrder = *in.MaxQuantityPerOrder
	}
	if in.ReviewSLAHours != nil && *in.ReviewSLAHours > 0 {
		p.ReviewSLAHours = *in.ReviewSLAHours
	}
	if in.FeeRate != nil && *in.FeeRate >= 0 {
		p.FeeRate = *in.FeeRate
	}
	if in.HelpText != nil {
		p.HelpText = strings.TrimSpace(*in.HelpText)
	}
	if in.RetailMarkupPercent != nil && *in.RetailMarkupPercent >= 0 {
		p.RetailMarkupPercent = *in.RetailMarkupPercent
	}

	encoded, err := json.Marshal(p)
	if err != nil {
		return err
	}
	return s.repo.Set(ctx, configSettingKey, string(encoded))
}

func (s *ConfigStore) readPersisted(ctx context.Context) (persistedConfig, error) {
	raw, err := s.repo.GetValue(ctx, configSettingKey)
	if err != nil {
		if errors.Is(err, service.ErrSettingNotFound) {
			return defaultPersistedConfig(), nil
		}
		return persistedConfig{}, err
	}
	return decodePersisted(raw), nil
}

func decodePersisted(raw string) persistedConfig {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return defaultPersistedConfig()
	}
	var p persistedConfig
	if err := json.Unmarshal([]byte(raw), &p); err != nil {
		return defaultPersistedConfig()
	}
	return p.withDefaults()
}

func defaultPersistedConfig() persistedConfig {
	return persistedConfig{
		UIEnabled:           true,
		APIBase:             DefaultAPIBase,
		OrderTimeoutHours:   24,
		MaxQuantityPerOrder: 10,
		ReviewSLAHours:      24,
		RetailMarkupPercent: 5,
		BankAccounts:        []BankAccount{},
	}
}

func (p persistedConfig) withDefaults() persistedConfig {
	d := defaultPersistedConfig()
	if p.APIBase == "" {
		p.APIBase = d.APIBase
	}
	if p.OrderTimeoutHours <= 0 {
		p.OrderTimeoutHours = d.OrderTimeoutHours
	}
	if p.MaxQuantityPerOrder <= 0 {
		p.MaxQuantityPerOrder = d.MaxQuantityPerOrder
	}
	if p.ReviewSLAHours <= 0 {
		p.ReviewSLAHours = d.ReviewSLAHours
	}
	if p.RetailMarkupPercent <= 0 {
		p.RetailMarkupPercent = d.RetailMarkupPercent
	}
	if p.BankAccounts == nil {
		p.BankAccounts = []BankAccount{}
	}
	return p
}

func (p persistedConfig) apiBase() string {
	base := strings.TrimSpace(p.APIBase)
	if base == "" {
		return DefaultAPIBase
	}
	return strings.TrimRight(base, "/")
}

func (p persistedConfig) effectiveSandbox() bool {
	if strings.HasPrefix(strings.TrimSpace(p.APIKey), "kvm_test_") {
		return true
	}
	return p.Sandbox
}

func (p persistedConfig) toConfig() Config {
	key := strings.TrimSpace(p.APIKey)
	secret := strings.TrimSpace(p.APISecret)
	enabled := p.Enabled
	if !enabled && key != "" && secret != "" {
		enabled = false
	}
	return Config{
		Enabled:   enabled,
		APIKey:    strings.TrimSpace(key),
		APISecret: strings.TrimSpace(secret),
		APIBase:   p.apiBase(),
		Sandbox:   p.effectiveSandbox(),
	}
}

func (p persistedConfig) toRuntime() RuntimeSettings {
	p = p.withDefaults()
	return RuntimeSettings{
		UIEnabled:           p.uiEnabled(),
		OrderTimeoutHours:   p.OrderTimeoutHours,
		MaxQuantityPerOrder: p.MaxQuantityPerOrder,
		ReviewSLAHours:      p.ReviewSLAHours,
		FeeRate:             p.FeeRate,
		HelpText:            p.HelpText,
		BankAccounts:        append([]BankAccount(nil), p.BankAccounts...),
		RetailMarkupPercent: p.RetailMarkupPercent,
	}
}

func (p persistedConfig) uiEnabled() bool {
	return p.UIEnabled
}
