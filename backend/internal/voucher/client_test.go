package voucher

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

// Matches KVoucher B2B docs Node.js example (Authentication section).
func TestSign_matchesOfficialExample(t *testing.T) {
	secret := "your-api-secret"
	sum := sha256.Sum256([]byte(secret))
	secretKey := hex.EncodeToString(sum[:])

	timestamp := "1710000000"
	method := "POST"
	path := "/api/kv/merchant/v1/orders"
	body := `{"items":[{"product_id":5,"quantity":100}]}`

	payload := timestamp + method + path + body
	mac := hmac.New(sha256.New, []byte(secretKey))
	_, _ = mac.Write([]byte(payload))
	got := hex.EncodeToString(mac.Sum(nil))

	client := NewClient(Config{APISecret: secret})
	sig := client.sign(method, path, body, timestamp)
	if sig != got {
		t.Fatalf("signature mismatch:\n got %s\nwant %s", sig, got)
	}
}

func TestSignPath_account(t *testing.T) {
	c := NewClient(Config{APIBase: DefaultAPIBase})
	if got := c.signPath("/account"); got != "/api/kv/merchant/v1/account" {
		t.Fatalf("unexpected sign path: %s", got)
	}
}

func TestRequestURL_defaultBase(t *testing.T) {
	c := NewClient(Config{APIBase: DefaultAPIBase})
	url := c.requestURL("/api/kv/merchant/v1/account")
	want := "https://kvoucher.com/api/kv/merchant/v1/account"
	if url != want {
		t.Fatalf("url=%s want=%s", url, want)
	}
}
