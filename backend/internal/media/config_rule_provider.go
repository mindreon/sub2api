package media

import "context"

// ConfigBackedRuleProvider lets admin-configured media prices override built-in
// defaults while preserving the Seedance fallback table for unconfigured models.
type ConfigBackedRuleProvider struct {
	store    *ConfigStore
	fallback RuleProvider
}

func NewConfigBackedRuleProvider(store *ConfigStore, fallback RuleProvider) *ConfigBackedRuleProvider {
	return &ConfigBackedRuleProvider{store: store, fallback: fallback}
}

func (p *ConfigBackedRuleProvider) RulesFor(model string) []MediaPricingRule {
	if p != nil && p.store != nil {
		cfg, err := p.store.Load(context.Background())
		if err == nil {
			if rules := cfg.RulesFor(model); len(rules) > 0 {
				return rules
			}
		}
	}
	if p == nil || p.fallback == nil {
		return nil
	}
	return p.fallback.RulesFor(model)
}
