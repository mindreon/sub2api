// Package voucher holds KVoucher B2B integration config and HTTP client helpers.
package voucher

import (
	"strings"
)

const DefaultAPIBase = "https://kvoucher.com/api/kv/merchant/v1"

// Config holds KVoucher API credentials (server-side only).
type Config struct {
	Enabled   bool
	APIKey    string
	APISecret string
	APIBase   string
	Sandbox   bool
}

// MaskAPIKey returns a safe display form, e.g. kvm_test_****abcd.
func MaskAPIKey(key string) string {
	key = strings.TrimSpace(key)
	if key == "" {
		return ""
	}
	if len(key) <= 12 {
		return key[:4] + "****"
	}
	return key[:12] + "****" + key[len(key)-4:]
}

func (c Config) Configured() bool {
	return c.APIKey != "" && c.APISecret != ""
}
