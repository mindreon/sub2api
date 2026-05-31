package voucher

import (
	"bytes"
	"context"
	"io"
	"strings"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/voucherauditlog"
	"github.com/Wei-Shaw/sub2api/ent/voucherb2border"
	"github.com/Wei-Shaw/sub2api/ent/voucherproduct"
	"entgo.io/ent/dialect/sql"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

// B2BOrderLine is a persisted line item snapshot.
type B2BOrderLine struct {
	ProductID    int     `json:"product_id"`
	Name         string  `json:"name,omitempty"`
	Denomination float64 `json:"denomination,omitempty"`
	Quantity     int     `json:"quantity"`
	UnitPrice    float64 `json:"unit_price,omitempty"`
	LineTotal    float64 `json:"line_total,omitempty"`
}

// B2BOrderView is returned to admin APIs.
type B2BOrderView struct {
	ID               int64          `json:"id"`
	KVOrderID        int64          `json:"kv_order_id"`
	OrderNo          string         `json:"order_no"`
	Status           string         `json:"status"`
	Subtotal         float64        `json:"subtotal"`
	FeeAmount        float64        `json:"fee_amount"`
	TotalAmount      float64        `json:"total_amount"`
	Currency         string         `json:"currency"`
	Items            []B2BOrderLine `json:"items"`
	PaymentRef       string         `json:"payment_ref,omitempty"`
	BankAccountID    *int           `json:"bank_account_id,omitempty"`
	MerchantNotes    string         `json:"merchant_notes,omitempty"`
	RejectReason     string         `json:"reject_reason,omitempty"`
	PaymentInfo      *PaymentInfo   `json:"payment_info,omitempty"`
	KVLastRequestID  string         `json:"kv_last_request_id,omitempty"`
	KVLastSyncedAt   *time.Time     `json:"kv_last_synced_at,omitempty"`
	CreatedBy        string         `json:"created_by"`
	VerifiedAt       *time.Time     `json:"verified_at,omitempty"`
	PinsLoadedAt     *time.Time     `json:"pins_loaded_at,omitempty"`
	CompletedAt      *time.Time     `json:"completed_at,omitempty"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
}

// AuditEntry is a voucher audit log row for admin display.
type AuditEntry struct {
	ID        int64          `json:"id"`
	Action    string         `json:"action"`
	Operator  string         `json:"operator"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
}

// AdminProductView is a catalog row for replenishment UI.
type AdminProductView struct {
	ID             int64   `json:"id"`
	KVProductID    *int64  `json:"kv_product_id,omitempty"`
	Name           string  `json:"name"`
	Denomination   float64 `json:"denomination"`
	WholesalePrice float64 `json:"wholesale_price"`
	RetailPrice    float64 `json:"retail_price"`
	Stock          int     `json:"stock"`
	Currency       string  `json:"currency"`
	IsActive       bool    `json:"is_active"`
}

type CreateB2BOrderInput struct {
	Items          []B2BOrderItem
	Currency       string
	MerchantNotes  string
	IdempotencyKey string
	Operator       string
}

type SubmitB2BProofInput struct {
	LocalOrderID int64
	PaymentRef   string
	BankID       *int
	ProofReader  io.Reader
	ProofName    string
	Operator     string
}

func (s *Service) AdminListProducts(ctx context.Context) ([]AdminProductView, error) {
	rows, err := s.ent.VoucherProduct.Query().
		Order(voucherproduct.BySortOrder(), voucherproduct.ByDenomination()).
		All(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]AdminProductView, 0, len(rows))
	for _, p := range rows {
		var kvID *int64
		if p.KvProductID != nil {
			kvID = p.KvProductID
		}
		out = append(out, AdminProductView{
			ID:             p.ID,
			KVProductID:    kvID,
			Name:           p.Name,
			Denomination:   p.Denomination,
			WholesalePrice: p.WholesalePrice,
			RetailPrice:    p.RetailPrice,
			Stock:          p.StockAvailable,
			Currency:       p.Currency,
			IsActive:       p.IsActive,
		})
	}
	return out, nil
}

func (s *Service) AdminCreateB2BOrder(ctx context.Context, in CreateB2BOrderInput) (*B2BOrderView, error) {
	cfg, _, err := s.store.Load(ctx)
	if err != nil {
		return nil, err
	}
	if !cfg.Configured() {
		return nil, infraerrors.BadRequest("KVOUCHER_NOT_CONFIGURED", "KVoucher API credentials not configured")
	}
	if len(in.Items) == 0 {
		return nil, infraerrors.BadRequest("INVALID_ITEMS", "at least one item is required")
	}
	for _, item := range in.Items {
		if item.ProductID <= 0 || item.Quantity <= 0 {
			return nil, infraerrors.BadRequest("INVALID_ITEMS", "invalid product_id or quantity")
		}
	}
	if key := strings.TrimSpace(in.IdempotencyKey); key != "" {
		if existing, qerr := s.ent.VoucherB2BOrder.Query().
			Where(voucherb2border.IdempotencyKeyEQ(key)).
			Only(ctx); qerr == nil && existing != nil {
			return s.b2bOrderView(existing)
		}
	}

	client := NewClient(cfg)
	kvReq := CreateB2BOrderRequest{
		Items:          in.Items,
		Currency:       strings.TrimSpace(in.Currency),
		MerchantNotes:  strings.TrimSpace(in.MerchantNotes),
		IdempotencyKey: strings.TrimSpace(in.IdempotencyKey),
	}
	kvOrder, reqID, err := client.CreateB2BOrder(ctx, kvReq)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	itemsJSON := kvOrderItemsToMaps(kvOrder.Items)
	paymentJSON := paymentInfoToMap(kvOrder.PaymentInfo)

	builder := s.ent.VoucherB2BOrder.Create().
		SetKvOrderID(kvOrder.ID).
		SetOrderNo(kvOrder.OrderNo).
		SetStatus(kvOrder.Status).
		SetSubtotal(kvOrder.Subtotal).
		SetFeeAmount(kvOrder.TotalFees).
		SetTotalAmount(kvOrder.TotalAmount).
		SetCurrency(defaultCurrency(kvOrder.Currency)).
		SetItemsJSON(itemsJSON).
		SetCreatedBy(strings.TrimSpace(in.Operator)).
		SetKvLastRequestID(reqID).
		SetKvLastSyncedAt(now)
	if paymentJSON != nil {
		builder.SetPaymentInfoJSON(paymentJSON)
	}
	if notes := strings.TrimSpace(in.MerchantNotes); notes != "" {
		builder.SetMerchantNotes(notes)
	}
	if key := strings.TrimSpace(in.IdempotencyKey); key != "" {
		builder.SetIdempotencyKey(key)
	}
	if kvOrder.RejectReason != "" {
		builder.SetRejectReason(kvOrder.RejectReason)
	}
	row, err := builder.Save(ctx)
	if err != nil {
		return nil, err
	}
	_ = s.writeB2BAudit(ctx, row.ID, "B2B_ORDER_CREATED", in.Operator, map[string]any{
		"kv_order_id": kvOrder.ID,
		"request_id":  reqID,
		"total":       kvOrder.TotalAmount,
	})
	return s.b2bOrderView(row)
}

func (s *Service) AdminListB2BOrders(ctx context.Context, status string, page, perPage int) ([]B2BOrderView, int, error) {
	if page <= 0 {
		page = 1
	}
	if perPage <= 0 || perPage > 100 {
		perPage = 25
	}
	q := s.ent.VoucherB2BOrder.Query()
	if status = strings.TrimSpace(status); status != "" {
		q = q.Where(voucherb2border.StatusEQ(status))
	}
	total, err := q.Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	rows, err := q.
		Order(voucherb2border.ByCreatedAt(sql.OrderDesc())).
		Offset((page - 1) * perPage).
		Limit(perPage).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}
	out := make([]B2BOrderView, 0, len(rows))
	for _, row := range rows {
		v, err := s.b2bOrderView(row)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, *v)
	}
	return out, total, nil
}

func (s *Service) AdminGetB2BOrder(ctx context.Context, id int64) (*B2BOrderView, error) {
	row, err := s.ent.VoucherB2BOrder.Get(ctx, id)
	if err != nil {
		return nil, infraerrors.NotFound("NOT_FOUND", "B2B order not found")
	}
	return s.b2bOrderView(row)
}

func (s *Service) AdminSyncB2BOrder(ctx context.Context, id int64, operator string) (*B2BOrderView, error) {
	cfg, _, err := s.store.Load(ctx)
	if err != nil {
		return nil, err
	}
	if !cfg.Configured() {
		return nil, infraerrors.BadRequest("KVOUCHER_NOT_CONFIGURED", "KVoucher API credentials not configured")
	}
	row, err := s.ent.VoucherB2BOrder.Get(ctx, id)
	if err != nil {
		return nil, infraerrors.NotFound("NOT_FOUND", "B2B order not found")
	}
	client := NewClient(cfg)
	kvOrder, reqID, err := client.GetB2BOrder(ctx, row.KvOrderID)
	if err != nil {
		return nil, err
	}
	prevStatus := row.Status
	updated, err := s.applyKVB2BOrder(ctx, row, kvOrder, reqID)
	if err != nil {
		return nil, err
	}
	_ = s.writeB2BAudit(ctx, updated.ID, "B2B_STATUS_SYNCED", operator, map[string]any{
		"request_id": reqID,
		"from":       prevStatus,
		"to":         updated.Status,
	})
	if updated.Status == "pins_loaded" || updated.Status == "completed" {
		_, _ = s.SyncStock(ctx)
	}
	return s.b2bOrderView(updated)
}

func (s *Service) AdminSubmitB2BProof(ctx context.Context, in SubmitB2BProofInput) (*B2BOrderView, error) {
	cfg, _, err := s.store.Load(ctx)
	if err != nil {
		return nil, err
	}
	if !cfg.Configured() {
		return nil, infraerrors.BadRequest("KVOUCHER_NOT_CONFIGURED", "KVoucher API credentials not configured")
	}
	row, err := s.ent.VoucherB2BOrder.Get(ctx, in.LocalOrderID)
	if err != nil {
		return nil, infraerrors.NotFound("NOT_FOUND", "B2B order not found")
	}
	if row.Status != "pending_payment" {
		return nil, infraerrors.BadRequest("INVALID_STATUS", "order cannot accept payment proof")
	}
	ref := strings.TrimSpace(in.PaymentRef)
	proofPath := ""
	var proofBytes []byte
	if in.ProofReader != nil && strings.TrimSpace(in.ProofName) != "" {
		var readErr error
		proofBytes, readErr = io.ReadAll(io.LimitReader(in.ProofReader, 5<<20))
		if readErr != nil {
			return nil, readErr
		}
		path, err := saveProofFile(in.ProofName, bytes.NewReader(proofBytes))
		if err != nil {
			return nil, err
		}
		proofPath = path
	}
	if ref == "" && proofPath == "" {
		return nil, infraerrors.BadRequest("PROOF_REQUIRED", "payment reference or proof file required")
	}

	client := NewClient(cfg)
	var uploadReader io.Reader
	if len(proofBytes) > 0 {
		uploadReader = bytes.NewReader(proofBytes)
	}
	kvOrder, reqID, err := client.SubmitB2BPaymentProof(ctx, row.KvOrderID, ref, uploadReader, in.ProofName, in.BankID)
	if err != nil {
		return nil, err
	}

	up := s.ent.VoucherB2BOrder.UpdateOneID(row.ID).
		SetStatus(kvOrder.Status).
		SetKvLastRequestID(reqID).
		SetKvLastSyncedAt(time.Now())
	if ref != "" {
		up.SetPaymentRef(ref)
	}
	if proofPath != "" {
		up.SetPaymentProofPath(proofPath)
	}
	if in.BankID != nil {
		up.SetBankAccountID(*in.BankID)
	}
	updated, err := up.Save(ctx)
	if err != nil {
		return nil, err
	}
	if kvOrder.PaymentInfo.Reference != "" || len(kvOrder.PaymentInfo.BankAccounts) > 0 {
		if m := paymentInfoToMap(kvOrder.PaymentInfo); m != nil {
			updated, _ = s.ent.VoucherB2BOrder.UpdateOneID(updated.ID).SetPaymentInfoJSON(m).Save(ctx)
		}
	}
	_ = s.writeB2BAudit(ctx, updated.ID, "B2B_PAYMENT_SUBMITTED", in.Operator, map[string]any{"request_id": reqID})
	return s.b2bOrderView(updated)
}

func (s *Service) AdminListB2BAudit(ctx context.Context, b2bOrderID int64) ([]AuditEntry, error) {
	rows, err := s.ent.VoucherAuditLog.Query().
		Where(voucherauditlog.B2bOrderIDEQ(b2bOrderID)).
		Order(voucherauditlog.ByCreatedAt(sql.OrderAsc())).
		All(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]AuditEntry, 0, len(rows))
	for _, row := range rows {
		out = append(out, AuditEntry{
			ID:        row.ID,
			Action:    row.Action,
			Operator:  row.Operator,
			Metadata:  row.Metadata,
			CreatedAt: row.CreatedAt,
		})
	}
	return out, nil
}

func (s *Service) applyKVB2BOrder(ctx context.Context, row *dbent.VoucherB2BOrder, kv *KVB2BOrder, reqID string) (*dbent.VoucherB2BOrder, error) {
	now := time.Now()
	up := s.ent.VoucherB2BOrder.UpdateOneID(row.ID).
		SetStatus(kv.Status).
		SetSubtotal(kv.Subtotal).
		SetFeeAmount(kv.TotalFees).
		SetTotalAmount(kv.TotalAmount).
		SetCurrency(defaultCurrency(kv.Currency)).
		SetItemsJSON(kvOrderItemsToMaps(kv.Items)).
		SetKvLastRequestID(reqID).
		SetKvLastSyncedAt(now)
	if m := paymentInfoToMap(kv.PaymentInfo); m != nil {
		up.SetPaymentInfoJSON(m)
	}
	if kv.RejectReason != "" {
		up.SetRejectReason(kv.RejectReason)
	}
	switch kv.Status {
	case "payment_verified":
		up.SetVerifiedAt(now)
	case "pins_loaded":
		up.SetPinsLoadedAt(now)
	case "completed":
		up.SetCompletedAt(now)
	}
	return up.Save(ctx)
}

func (s *Service) b2bOrderView(row *dbent.VoucherB2BOrder) (*B2BOrderView, error) {
	items := make([]B2BOrderLine, 0)
	for _, raw := range row.ItemsJSON {
		line := B2BOrderLine{}
		if v, ok := raw["product_id"]; ok {
			switch n := v.(type) {
			case float64:
				line.ProductID = int(n)
			case int:
				line.ProductID = n
			}
		}
		if v, ok := raw["name"].(string); ok {
			line.Name = v
		}
		if v, ok := raw["denomination"].(float64); ok {
			line.Denomination = v
		}
		if v, ok := raw["quantity"].(float64); ok {
			line.Quantity = int(v)
		} else if v, ok := raw["quantity"].(int); ok {
			line.Quantity = v
		}
		if v, ok := raw["unit_price"].(float64); ok {
			line.UnitPrice = v
		}
		if v, ok := raw["line_total"].(float64); ok {
			line.LineTotal = v
		}
		items = append(items, line)
	}
	var paymentInfo *PaymentInfo
	if row.PaymentInfoJSON != nil {
		paymentInfo = mapToPaymentInfo(row.PaymentInfoJSON)
	}
	view := &B2BOrderView{
		ID:              row.ID,
		KVOrderID:       row.KvOrderID,
		OrderNo:         row.OrderNo,
		Status:          row.Status,
		Subtotal:        row.Subtotal,
		FeeAmount:       row.FeeAmount,
		TotalAmount:     row.TotalAmount,
		Currency:        row.Currency,
		Items:           items,
		MerchantNotes:   derefStr(row.MerchantNotes),
		RejectReason:    derefStr(row.RejectReason),
		PaymentInfo:     paymentInfo,
		KVLastRequestID: derefStr(row.KvLastRequestID),
		CreatedBy:       row.CreatedBy,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}
	if row.PaymentRef != nil {
		view.PaymentRef = *row.PaymentRef
	}
	if row.BankAccountID != nil {
		view.BankAccountID = row.BankAccountID
	}
	if row.KvLastSyncedAt != nil {
		view.KVLastSyncedAt = row.KvLastSyncedAt
	}
	if row.VerifiedAt != nil {
		view.VerifiedAt = row.VerifiedAt
	}
	if row.PinsLoadedAt != nil {
		view.PinsLoadedAt = row.PinsLoadedAt
	}
	if row.CompletedAt != nil {
		view.CompletedAt = row.CompletedAt
	}
	return view, nil
}

func (s *Service) writeB2BAudit(ctx context.Context, b2bOrderID int64, action, operator string, metadata map[string]any) error {
	b := s.ent.VoucherAuditLog.Create().
		SetB2bOrderID(b2bOrderID).
		SetAction(action).
		SetOperator(operator)
	if metadata != nil {
		b.SetMetadata(metadata)
	}
	_, err := b.Save(ctx)
	return err
}

func kvOrderItemsToMaps(items []KVB2BOrderItem) []map[string]any {
	out := make([]map[string]any, 0, len(items))
	for _, it := range items {
		out = append(out, map[string]any{
			"product_id":   it.ProductID,
			"name":         it.Name,
			"denomination": it.Denomination,
			"quantity":     it.Quantity,
			"unit_price":   it.UnitPrice,
			"line_total":   it.LineTotal,
		})
	}
	return out
}

func paymentInfoToMap(p PaymentInfo) map[string]any {
	if p.Reference == "" && len(p.BankAccounts) == 0 && p.AmountDue == 0 {
		return nil
	}
	banks := make([]map[string]any, 0, len(p.BankAccounts))
	for _, b := range p.BankAccounts {
		banks = append(banks, map[string]any{
			"id":             b.ID,
			"bank_name":      b.BankName,
			"account_name":   b.AccountName,
			"account_number": b.AccountNumber,
			"type":           b.Type,
		})
	}
	return map[string]any{
		"reference":     p.Reference,
		"amount_due":    p.AmountDue,
		"currency":      p.Currency,
		"instructions":  p.Instructions,
		"bank_accounts": banks,
	}
}

func mapToPaymentInfo(m map[string]any) *PaymentInfo {
	p := &PaymentInfo{}
	if v, ok := m["reference"].(string); ok {
		p.Reference = v
	}
	if v, ok := m["amount_due"].(float64); ok {
		p.AmountDue = v
	}
	if v, ok := m["currency"].(string); ok {
		p.Currency = v
	}
	if v, ok := m["instructions"].(string); ok {
		p.Instructions = v
	}
	if banks, ok := m["bank_accounts"].([]any); ok {
		for _, raw := range banks {
			bm, ok := raw.(map[string]any)
			if !ok {
				continue
			}
			acc := BankAccount{}
			if v, ok := bm["id"].(float64); ok {
				acc.ID = int(v)
			}
			if v, ok := bm["bank_name"].(string); ok {
				acc.BankName = v
			}
			if v, ok := bm["account_name"].(string); ok {
				acc.AccountName = v
			}
			if v, ok := bm["account_number"].(string); ok {
				acc.AccountNumber = v
			}
			if v, ok := bm["type"].(string); ok {
				acc.Type = v
			}
			p.BankAccounts = append(p.BankAccounts, acc)
		}
	}
	return p
}

func defaultCurrency(c string) string {
	if strings.TrimSpace(c) == "" {
		return "MYR"
	}
	return strings.TrimSpace(c)
}

func derefStr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// IsVoucherEnabled reports whether KVoucher backend integration is on.
func (s *Service) IsVoucherEnabled(ctx context.Context) (bool, error) {
	cfg, _, err := s.store.Load(ctx)
	if err != nil {
		return false, err
	}
	return cfg.Enabled, nil
}
