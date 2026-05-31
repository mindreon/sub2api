package voucher

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const defaultHTTPTimeout = 15 * time.Second

// AccountSummary is a subset of GET /account data used for connection tests.
type AccountSummary struct {
	MerchantID   int    `json:"merchant_id"`
	CompanyName  string `json:"company_name"`
	Currency     string `json:"currency"`
	APIKeyScope  string `json:"api_key_scope"`
	IsTestMode   bool   `json:"is_test_mode"`
}

type envelope struct {
	OK        bool            `json:"ok"`
	RequestID string          `json:"request_id"`
	Data      json.RawMessage `json:"data"`
	Error     *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// Client calls KVoucher Merchant API v1 with HMAC auth.
type Client struct {
	cfg        Config
	httpClient *http.Client
}

func NewClient(cfg Config) *Client {
	return &Client{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: defaultHTTPTimeout,
		},
	}
}

func (c *Client) sign(method, pathWithQuery, body string, timestamp string) string {
	sum := sha256.Sum256([]byte(c.cfg.APISecret))
	// KVoucher docs: secret_hash = SHA256(api_secret) as hex string; HMAC key is that hex string.
	secretKey := hex.EncodeToString(sum[:])
	payload := timestamp + method + pathWithQuery + body
	mac := hmac.New(sha256.New, []byte(secretKey))
	_, _ = mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}

const merchantV1Prefix = "/api/kv/merchant/v1"

func normalizeEndpoint(endpoint string) string {
	if !strings.HasPrefix(endpoint, "/") {
		endpoint = "/" + endpoint
	}
	return endpoint
}

// signPath returns the path_with_query value used in HMAC payload (full path from site root).
func (c *Client) signPath(endpoint string) string {
	ep := normalizeEndpoint(endpoint)
	if strings.HasPrefix(ep, merchantV1Prefix) {
		return ep
	}
	return merchantV1Prefix + ep
}

// requestURL builds the HTTP URL from APIBase and the signed path.
func (c *Client) requestURL(signPath string) string {
	base := strings.TrimRight(c.cfg.APIBase, "/")
	if strings.HasSuffix(base, "/merchant/v1") {
		suffix := strings.TrimPrefix(signPath, merchantV1Prefix)
		return base + suffix
	}
	return base + signPath
}

func (c *Client) apiPath(endpoint string) string {
	return c.signPath(endpoint)
}

func (c *Client) do(ctx context.Context, method, endpoint, body string) ([]byte, int, error) {
	pathWithQuery := c.signPath(endpoint)
	url := c.requestURL(pathWithQuery)
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	sig := c.sign(method, pathWithQuery, body, ts)

	req, err := http.NewRequestWithContext(ctx, method, url, strings.NewReader(body))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-API-Key", c.cfg.APIKey)
	req.Header.Set("X-Signature", sig)
	req.Header.Set("X-Timestamp", ts)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return raw, resp.StatusCode, nil
}

// GetAccount verifies credentials via GET /account.
func (c *Client) GetAccount(ctx context.Context) (*AccountSummary, string, error) {
	raw, status, err := c.do(ctx, http.MethodGet, "/account", "")
	if err != nil {
		return nil, "", err
	}
	var env envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return nil, "", fmt.Errorf("parse response: %w", err)
	}
	if !env.OK {
		msg := "request failed"
		if env.Error != nil && env.Error.Message != "" {
			msg = env.Error.Message
		}
		return nil, env.RequestID, fmt.Errorf("HTTP %d: %s", status, msg)
	}
	var data struct {
		MerchantID  int    `json:"merchant_id"`
		CompanyName string `json:"company_name"`
		Currency    string `json:"currency"`
		APIKeyScope string `json:"api_key_scope"`
		IsTestMode  bool   `json:"is_test_mode"`
	}
	if err := json.Unmarshal(env.Data, &data); err != nil {
		return nil, env.RequestID, fmt.Errorf("parse account data: %w", err)
	}
	return &AccountSummary{
		MerchantID:  data.MerchantID,
		CompanyName: data.CompanyName,
		Currency:    data.Currency,
		APIKeyScope: data.APIKeyScope,
		IsTestMode:  data.IsTestMode,
	}, env.RequestID, nil
}

// TestConnectionResult is returned to admin after a connectivity check.
type TestConnectionResult struct {
	OK         bool            `json:"ok"`
	RequestID  string          `json:"request_id,omitempty"`
	Message    string          `json:"message"`
	Account    *AccountSummary `json:"account,omitempty"`
	Configured bool            `json:"configured"`
}

// TestConnection validates env credentials against KVoucher.
func TestConnection(ctx context.Context, cfg Config) TestConnectionResult {
	if !cfg.Configured() {
		return TestConnectionResult{
			OK:         false,
			Configured: false,
			Message:    "KVoucher API key and secret are required",
		}
	}
	client := NewClient(cfg)
	account, reqID, err := client.GetAccount(ctx)
	if err != nil {
		return TestConnectionResult{
			OK:         false,
			Configured: true,
			RequestID:  reqID,
			Message:    err.Error(),
		}
	}
	msg := fmt.Sprintf("Connected as merchant #%d (%s)", account.MerchantID, account.CompanyName)
	if account.IsTestMode || cfg.Sandbox {
		msg += " [sandbox]"
	}
	return TestConnectionResult{
		OK:         true,
		Configured: true,
		RequestID:  reqID,
		Message:    msg,
		Account:    account,
	}
}

// Product is a KVoucher catalog item.
type Product struct {
	ID             int     `json:"id"`
	Name           string  `json:"name"`
	Denomination   float64 `json:"denomination"`
	WholesalePrice float64 `json:"wholesale_price"`
	YourStock      int     `json:"your_stock"`
}

// StockEntry is available stock for one denomination.
type StockEntry struct {
	Denomination float64 `json:"denomination"`
	Available    int     `json:"available"`
}

// RetrievedPIN is one PIN from POST /stock/retrieve.
type RetrievedPIN struct {
	PinCode      string  `json:"pin_code"`
	Serial       string  `json:"serial"`
	ExpiresAt    string  `json:"expires_at"`
	Denomination float64 `json:"denomination"`
}

// ListProducts calls GET /products.
func (c *Client) ListProducts(ctx context.Context) ([]Product, string, error) {
	raw, status, err := c.do(ctx, http.MethodGet, "/products", "")
	if err != nil {
		return nil, "", err
	}
	var env envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return nil, "", err
	}
	if !env.OK {
		return nil, env.RequestID, fmt.Errorf("HTTP %d: %s", status, envelopeMessage(&env))
	}
	var data struct {
		Products []Product `json:"products"`
	}
	if err := json.Unmarshal(env.Data, &data); err != nil {
		return nil, env.RequestID, err
	}
	return data.Products, env.RequestID, nil
}

// ListStock calls GET /stock.
func (c *Client) ListStock(ctx context.Context) ([]StockEntry, int, string, error) {
	raw, status, err := c.do(ctx, http.MethodGet, "/stock", "")
	if err != nil {
		return nil, 0, "", err
	}
	var env envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return nil, 0, "", err
	}
	if !env.OK {
		return nil, 0, env.RequestID, fmt.Errorf("HTTP %d: %s", status, envelopeMessage(&env))
	}
	var data struct {
		Stock          []StockEntry `json:"stock"`
		TotalAvailable int          `json:"total_available"`
	}
	if err := json.Unmarshal(env.Data, &data); err != nil {
		return nil, 0, env.RequestID, err
	}
	return data.Stock, data.TotalAvailable, env.RequestID, nil
}

type retrieveRequest struct {
	Denomination float64 `json:"denomination"`
	Quantity     int     `json:"quantity"`
	Reference    string  `json:"reference"`
}

// RetrieveStock calls POST /stock/retrieve.
func (c *Client) RetrieveStock(ctx context.Context, denomination float64, quantity int, reference string) ([]RetrievedPIN, string, error) {
	bodyBytes, err := json.Marshal(retrieveRequest{
		Denomination: denomination,
		Quantity:     quantity,
		Reference:    reference,
	})
	if err != nil {
		return nil, "", err
	}
	body := string(bodyBytes)
	raw, status, err := c.do(ctx, http.MethodPost, "/stock/retrieve", body)
	if err != nil {
		return nil, "", err
	}
	var env envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return nil, "", err
	}
	if !env.OK {
		return nil, env.RequestID, fmt.Errorf("HTTP %d: %s", status, envelopeMessage(&env))
	}
	var data struct {
		Retrieved []RetrievedPIN `json:"retrieved"`
	}
	if err := json.Unmarshal(env.Data, &data); err != nil {
		return nil, env.RequestID, err
	}
	return data.Retrieved, env.RequestID, nil
}

func envelopeMessage(env *envelope) string {
	if env.Error != nil && env.Error.Message != "" {
		return env.Error.Message
	}
	return "request failed"
}
