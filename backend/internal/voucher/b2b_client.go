package voucher

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// B2BOrderItem is a line item for KVoucher POST /orders.
type B2BOrderItem struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

// CreateB2BOrderRequest is the KVoucher wholesale order payload.
type CreateB2BOrderRequest struct {
	Items          []B2BOrderItem `json:"items"`
	Currency       string         `json:"currency,omitempty"`
	MerchantNotes  string         `json:"merchant_notes,omitempty"`
	IdempotencyKey string         `json:"idempotency_key,omitempty"`
}

// KVB2BOrderItem is a line on a KVoucher B2B order response.
type KVB2BOrderItem struct {
	ProductID    int     `json:"product_id"`
	Name         string  `json:"name"`
	Denomination float64 `json:"denomination"`
	Quantity     int     `json:"quantity"`
	UnitPrice    float64 `json:"unit_price"`
	LineTotal    float64 `json:"line_total"`
}

// KVB2BOrder is KVoucher's wholesale order object.
type KVB2BOrder struct {
	ID          int64            `json:"id"`
	OrderNo     string           `json:"order_no"`
	Status      string           `json:"status"`
	Subtotal    float64          `json:"subtotal"`
	TotalFees   float64          `json:"total_fees"`
	TotalAmount float64          `json:"total_amount"`
	Currency    string           `json:"currency"`
	Items       []KVB2BOrderItem `json:"items"`
	PaymentInfo PaymentInfo      `json:"payment_info"`
	RejectReason string          `json:"reject_reason,omitempty"`
}

// CreateB2BOrder calls POST /orders.
func (c *Client) CreateB2BOrder(ctx context.Context, req CreateB2BOrderRequest) (*KVB2BOrder, string, error) {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, "", err
	}
	body := string(bodyBytes)
	raw, status, err := c.do(ctx, http.MethodPost, "/orders", body)
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
		Order KVB2BOrder `json:"order"`
	}
	if err := json.Unmarshal(env.Data, &data); err != nil {
		return nil, env.RequestID, err
	}
	return &data.Order, env.RequestID, nil
}

// GetB2BOrder calls GET /orders/{id}.
func (c *Client) GetB2BOrder(ctx context.Context, kvOrderID int64) (*KVB2BOrder, string, error) {
	endpoint := "/orders/" + strconv.FormatInt(kvOrderID, 10)
	raw, status, err := c.do(ctx, http.MethodGet, endpoint, "")
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
		Order KVB2BOrder `json:"order"`
	}
	if err := json.Unmarshal(env.Data, &data); err != nil {
		return nil, env.RequestID, err
	}
	return &data.Order, env.RequestID, nil
}

// ListKVB2BOrders calls GET /orders with optional status filter.
func (c *Client) ListKVB2BOrders(ctx context.Context, status string, page, perPage int) ([]KVB2BOrder, int, string, error) {
	endpoint := "/orders?page=" + strconv.Itoa(page) + "&per_page=" + strconv.Itoa(perPage)
	if status = strings.TrimSpace(status); status != "" {
		endpoint += "&status=" + status
	}
	raw, httpStatus, err := c.do(ctx, http.MethodGet, endpoint, "")
	if err != nil {
		return nil, 0, "", err
	}
	var env envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return nil, 0, "", err
	}
	if !env.OK {
		return nil, 0, env.RequestID, fmt.Errorf("HTTP %d: %s", httpStatus, envelopeMessage(&env))
	}
	var data struct {
		Orders []KVB2BOrder `json:"orders"`
	}
	if err := json.Unmarshal(env.Data, &data); err != nil {
		return nil, 0, env.RequestID, err
	}
	total := 0
	if env.Pagination != nil {
		total = env.Pagination.Total
	}
	return data.Orders, total, env.RequestID, nil
}

// SubmitB2BPaymentProof calls POST /orders/{id}/payment-proof (multipart, empty HMAC body).
func (c *Client) SubmitB2BPaymentProof(ctx context.Context, kvOrderID int64, paymentRef string, proofReader io.Reader, proofFilename string, bankID *int) (*KVB2BOrder, string, error) {
	endpoint := "/orders/" + strconv.FormatInt(kvOrderID, 10) + "/payment-proof"
	pathWithQuery := c.signPath(endpoint)
	url := c.requestURL(pathWithQuery)

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	if paymentRef != "" {
		if err := mw.WriteField("payment_ref", paymentRef); err != nil {
			return nil, "", err
		}
	}
	if proofReader != nil && proofFilename != "" {
		fw, err := mw.CreateFormFile("payment_proof", proofFilename)
		if err != nil {
			return nil, "", err
		}
		if _, err := io.Copy(fw, io.LimitReader(proofReader, 5<<20)); err != nil {
			return nil, "", err
		}
	}
	if bankID != nil {
		if err := mw.WriteField("bank_id", strconv.Itoa(*bankID)); err != nil {
			return nil, "", err
		}
	}
	if err := mw.Close(); err != nil {
		return nil, "", err
	}

	ts := strconv.FormatInt(time.Now().Unix(), 10)
	sig := c.sign(http.MethodPost, pathWithQuery, "", ts)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &buf)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-API-Key", c.cfg.APIKey)
	req.Header.Set("X-Signature", sig)
	req.Header.Set("X-Timestamp", ts)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, "", err
	}
	var env envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return nil, "", err
	}
	if !env.OK {
		return nil, env.RequestID, fmt.Errorf("HTTP %d: %s", resp.StatusCode, envelopeMessage(&env))
	}
	var data struct {
		Order KVB2BOrder `json:"order"`
	}
	if err := json.Unmarshal(env.Data, &data); err != nil {
		return nil, env.RequestID, err
	}
	return &data.Order, env.RequestID, nil
}
