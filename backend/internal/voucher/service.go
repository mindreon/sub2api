package voucher

import (
	"context"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/user"
	"github.com/Wei-Shaw/sub2api/ent/voucherorder"
	"github.com/Wei-Shaw/sub2api/ent/voucherpindelivery"
	"github.com/Wei-Shaw/sub2api/ent/voucherproduct"
	"github.com/Wei-Shaw/sub2api/internal/setup"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

// Service implements voucher order lifecycle.
type Service struct {
	ent   *dbent.Client
	store *ConfigStore
}

func NewService(entClient *dbent.Client, store *ConfigStore) *Service {
	return &Service{ent: entClient, store: store}
}

type CheckoutProduct struct {
	ID           int64   `json:"id"`
	Name         string  `json:"name"`
	Denomination float64 `json:"denomination"`
	RetailPrice  float64 `json:"retail_price"`
	Stock        int     `json:"stock"`
	Currency     string  `json:"currency"`
}

type CheckoutInfo struct {
	Enabled           bool              `json:"enabled"`
	CheckoutReady     bool              `json:"checkout_ready"`
	Currency          string            `json:"currency"`
	Products        []CheckoutProduct `json:"products"`
	BankAccounts    []BankAccount `json:"bank_accounts"`
	OrderTimeoutHours int         `json:"order_timeout_hours"`
	MaxQuantity     int           `json:"max_quantity"`
	ReviewSLAHours  int           `json:"review_sla_hours"`
	FeeRate         float64       `json:"fee_rate"`
	HelpText        string        `json:"help_text"`
	Sandbox         bool          `json:"sandbox"`
}

type CreateOrderInput struct {
	UserID         int64
	ProductID      int64
	Quantity       int
	IdempotencyKey string
	ClientIP       string
}

type PaymentInfo struct {
	Reference     string        `json:"reference"`
	AmountDue     float64       `json:"amount_due"`
	Currency      string        `json:"currency"`
	BankAccounts  []BankAccount `json:"bank_accounts"`
	Instructions  string        `json:"instructions"`
}

type OrderView struct {
	ID            int64          `json:"id"`
	OrderNo       string         `json:"order_no"`
	Status        string         `json:"status"`
	ProductName   string         `json:"product_name"`
	Denomination  float64        `json:"denomination"`
	Quantity      int            `json:"quantity"`
	Subtotal      float64        `json:"subtotal"`
	FeeAmount     float64        `json:"fee_amount"`
	TotalAmount   float64        `json:"total_amount"`
	Currency      string         `json:"currency"`
	PaymentRef    string         `json:"payment_ref,omitempty"`
	RejectReason  string         `json:"reject_reason,omitempty"`
	FulfillError  string         `json:"fulfill_error,omitempty"`
	ExpiresAt     time.Time      `json:"expires_at"`
	CreatedAt     time.Time      `json:"created_at"`
	CompletedAt   *time.Time     `json:"completed_at,omitempty"`
	PaymentInfo   *PaymentInfo   `json:"payment_info,omitempty"`
	Pins          []PinView      `json:"pins,omitempty"`
}

type PinView struct {
	PinCode      string  `json:"pin_code"`
	Serial       string  `json:"serial"`
	Denomination float64 `json:"denomination"`
	ExpiresAt    string  `json:"expires_at,omitempty"`
	Masked       bool    `json:"masked,omitempty"`
}

func (s *Service) CheckoutInfo(ctx context.Context) (*CheckoutInfo, error) {
	cfg, runtime, err := s.store.Load(ctx)
	if err != nil {
		return nil, err
	}
	if !cfg.Enabled {
		return &CheckoutInfo{Enabled: false, CheckoutReady: false}, nil
	}
	products, err := s.ent.VoucherProduct.Query().
		Where(voucherproduct.IsActive(true)).
		Order(dbent.Asc(voucherproduct.FieldSortOrder), dbent.Asc(voucherproduct.FieldDenomination)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]CheckoutProduct, 0, len(products))
	currency := "MYR"
	for _, p := range products {
		if p.Currency != "" {
			currency = p.Currency
		}
		out = append(out, CheckoutProduct{
			ID:           p.ID,
			Name:         p.Name,
			Denomination: p.Denomination,
			RetailPrice:  p.RetailPrice,
			Stock:        p.StockAvailable,
			Currency:     p.Currency,
		})
	}
	return &CheckoutInfo{
		Enabled:           cfg.Enabled && runtime.UIEnabled,
		CheckoutReady:     len(out) > 0 && len(runtime.BankAccounts) > 0,
		Currency:          currency,
		Products:          out,
		BankAccounts:      runtime.BankAccounts,
		OrderTimeoutHours: runtime.OrderTimeoutHours,
		MaxQuantity:       runtime.MaxQuantityPerOrder,
		ReviewSLAHours:    runtime.ReviewSLAHours,
		FeeRate:           runtime.FeeRate,
		HelpText:          runtime.HelpText,
		Sandbox:           cfg.Sandbox,
	}, nil
}

func (s *Service) SyncCatalog(ctx context.Context) (int, error) {
	cfg, runtime, err := s.store.Load(ctx)
	if err != nil {
		return 0, err
	}
	if !cfg.Configured() {
		return 0, infraerrors.BadRequest("KVOUCHER_NOT_CONFIGURED", "KVoucher API credentials not configured")
	}
	client := NewClient(cfg)
	products, _, err := client.ListProducts(ctx)
	if err != nil {
		return 0, err
	}
	count := 0
	for i, p := range products {
		retail := p.WholesalePrice * (1 + runtime.RetailMarkupPercent/100)
		retail = math.Round(retail*100) / 100
		existing, qerr := s.ent.VoucherProduct.Query().
			Where(voucherproduct.KvProductIDEQ(int64(p.ID))).
			Only(ctx)
		if qerr == nil && existing != nil && existing.RetailPrice > 0 {
			retail = existing.RetailPrice
		}
		kvID := int64(p.ID)
		if existing != nil {
			_, err = s.ent.VoucherProduct.UpdateOneID(existing.ID).
				SetName(p.Name).
				SetDenomination(p.Denomination).
				SetWholesalePrice(p.WholesalePrice).
				SetRetailPrice(retail).
				SetStockAvailable(p.YourStock).
				SetSortOrder(i).
				SetIsActive(true).
				Save(ctx)
		} else {
			_, err = s.ent.VoucherProduct.Create().
				SetKvProductID(kvID).
				SetName(p.Name).
				SetDenomination(p.Denomination).
				SetWholesalePrice(p.WholesalePrice).
				SetRetailPrice(retail).
				SetStockAvailable(p.YourStock).
				SetSortOrder(i).
				SetIsActive(true).
				Save(ctx)
		}
		if err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

func (s *Service) SyncStock(ctx context.Context) (int, error) {
	cfg, _, err := s.store.Load(ctx)
	if err != nil {
		return 0, err
	}
	if !cfg.Configured() {
		return 0, infraerrors.BadRequest("KVOUCHER_NOT_CONFIGURED", "KVoucher API credentials not configured")
	}
	client := NewClient(cfg)
	stock, _, _, err := client.ListStock(ctx)
	if err != nil {
		return 0, err
	}
	updated := 0
	for _, entry := range stock {
		n, err := s.ent.VoucherProduct.Update().
			Where(voucherproduct.DenominationEQ(entry.Denomination)).
			SetStockAvailable(entry.Available).
			Save(ctx)
		if err != nil {
			return updated, err
		}
		updated += n
	}
	return updated, nil
}

func (s *Service) CreateOrder(ctx context.Context, in CreateOrderInput) (*OrderView, error) {
	cfg, runtime, err := s.store.Load(ctx)
	if err != nil {
		return nil, err
	}
	if !cfg.Enabled {
		return nil, infraerrors.Forbidden("VOUCHER_DISABLED", "voucher purchase is disabled")
	}
	if len(runtime.BankAccounts) == 0 {
		return nil, infraerrors.BadRequest("NO_BANK_ACCOUNT", "collection bank account not configured")
	}
	if in.Quantity <= 0 || in.Quantity > runtime.MaxQuantityPerOrder {
		return nil, infraerrors.BadRequest("INVALID_QUANTITY", "invalid quantity")
	}
	if key := strings.TrimSpace(in.IdempotencyKey); key != "" {
		if existing, err := s.ent.VoucherOrder.Query().
			Where(voucherorder.IdempotencyKeyEQ(key)).
			Only(ctx); err == nil && existing != nil {
			return s.orderView(ctx, existing, true)
		}
	}
	product, err := s.ent.VoucherProduct.Query().
		Where(voucherproduct.IDEQ(in.ProductID), voucherproduct.IsActive(true)).
		Only(ctx)
	if err != nil {
		return nil, infraerrors.NotFound("PRODUCT_NOT_FOUND", "product not found")
	}
	if product.StockAvailable < in.Quantity {
		return nil, infraerrors.BadRequest("INSUFFICIENT_STOCK", "insufficient stock")
	}
	u, err := s.ent.User.Query().Where(user.IDEQ(in.UserID)).Only(ctx)
	if err != nil {
		return nil, err
	}
	subtotal := roundMoney(product.RetailPrice * float64(in.Quantity))
	fee := 0.0
	if runtime.FeeRate > 0 {
		fee = roundMoney(subtotal * runtime.FeeRate / 100)
	}
	total := roundMoney(subtotal + fee)
	orderNo := allocateOrderNo()
	expires := time.Now().Add(time.Duration(runtime.OrderTimeoutHours) * time.Hour)
	builder := s.ent.VoucherOrder.Create().
		SetOrderNo(orderNo).
		SetUserID(u.ID).
		SetUserEmail(u.Email).
		SetUserName(u.Username).
		SetStatus(OrderStatusPendingPayment).
		SetProductID(product.ID).
		SetNillableKvProductID(product.KvProductID).
		SetProductName(product.Name).
		SetDenomination(product.Denomination).
		SetQuantity(in.Quantity).
		SetUnitPrice(product.RetailPrice).
		SetSubtotal(subtotal).
		SetFeeAmount(fee).
		SetTotalAmount(total).
		SetCurrency(product.Currency).
		SetKvRetrieveReference(orderNo).
		SetExpiresAt(expires).
		SetClientIP(in.ClientIP)
	if key := strings.TrimSpace(in.IdempotencyKey); key != "" {
		builder.SetIdempotencyKey(key)
	}
	o, err := builder.Save(ctx)
	if err != nil {
		return nil, err
	}
	_ = s.writeAudit(ctx, o.ID, "ORDER_CREATED", fmt.Sprintf("user:%d", u.ID), map[string]any{
		"total": total, "quantity": in.Quantity,
	})
	return s.orderView(ctx, o, true)
}

type SubmitProofInput struct {
	UserID      int64
	OrderID     int64
	PaymentRef  string
	BankID      *int
	ProofReader io.Reader
	ProofName   string
}

func (s *Service) SubmitPaymentProof(ctx context.Context, in SubmitProofInput) (*OrderView, error) {
	o, err := s.getUserOrder(ctx, in.UserID, in.OrderID)
	if err != nil {
		return nil, err
	}
	if o.Status != OrderStatusPendingPayment {
		return nil, infraerrors.BadRequest("INVALID_STATUS", "order cannot accept payment proof")
	}
	ref := strings.TrimSpace(in.PaymentRef)
	proofPath := ""
	if in.ProofReader != nil && strings.TrimSpace(in.ProofName) != "" {
		path, err := saveProofFile(in.ProofName, in.ProofReader)
		if err != nil {
			return nil, err
		}
		proofPath = path
	}
	if ref == "" && proofPath == "" {
		return nil, infraerrors.BadRequest("PROOF_REQUIRED", "payment reference or proof file required")
	}
	up := s.ent.VoucherOrder.UpdateOneID(o.ID).SetStatus(OrderStatusPaymentSubmitted)
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
	_ = s.writeAudit(ctx, o.ID, "PAYMENT_SUBMITTED", fmt.Sprintf("user:%d", in.UserID), nil)
	return s.orderView(ctx, updated, false)
}

func (s *Service) ListMyOrders(ctx context.Context, userID int64, page, perPage int) ([]OrderView, int, error) {
	if page <= 0 {
		page = 1
	}
	if perPage <= 0 || perPage > 100 {
		perPage = 25
	}
	total, err := s.ent.VoucherOrder.Query().Where(voucherorder.UserIDEQ(userID)).Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	rows, err := s.ent.VoucherOrder.Query().
		Where(voucherorder.UserIDEQ(userID)).
		Order(dbent.Desc(voucherorder.FieldCreatedAt)).
		Offset((page - 1) * perPage).
		Limit(perPage).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}
	out := make([]OrderView, 0, len(rows))
	for _, row := range rows {
		v, err := s.orderView(ctx, row, false)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, *v)
	}
	return out, total, nil
}

func (s *Service) GetOrder(ctx context.Context, userID, orderID int64, includePins bool) (*OrderView, error) {
	o, err := s.getUserOrder(ctx, userID, orderID)
	if err != nil {
		return nil, err
	}
	return s.orderView(ctx, o, includePins || o.Status == OrderStatusCompleted)
}

func (s *Service) CancelOrder(ctx context.Context, userID, orderID int64) (*OrderView, error) {
	o, err := s.getUserOrder(ctx, userID, orderID)
	if err != nil {
		return nil, err
	}
	if o.Status != OrderStatusPendingPayment {
		return nil, infraerrors.BadRequest("INVALID_STATUS", "only pending orders can be cancelled")
	}
	updated, err := s.ent.VoucherOrder.UpdateOneID(o.ID).
		SetStatus(OrderStatusExpired).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	_ = s.writeAudit(ctx, o.ID, "ORDER_CANCELLED", fmt.Sprintf("user:%d", userID), nil)
	return s.orderView(ctx, updated, false)
}

func (s *Service) AdminListOrders(ctx context.Context, status string, page, perPage int) ([]OrderView, int, error) {
	if page <= 0 {
		page = 1
	}
	if perPage <= 0 || perPage > 100 {
		perPage = 25
	}
	q := s.ent.VoucherOrder.Query()
	if status = strings.TrimSpace(status); status != "" {
		q = q.Where(voucherorder.StatusEQ(status))
	}
	total, err := q.Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	rows, err := q.Order(dbent.Desc(voucherorder.FieldCreatedAt)).
		Offset((page - 1) * perPage).
		Limit(perPage).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}
	out := make([]OrderView, 0, len(rows))
	for _, row := range rows {
		v, err := s.orderView(ctx, row, row.Status == OrderStatusCompleted)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, *v)
	}
	return out, total, nil
}

func (s *Service) AdminGetOrder(ctx context.Context, orderID int64) (*OrderView, error) {
	o, err := s.ent.VoucherOrder.Get(ctx, orderID)
	if err != nil {
		return nil, infraerrors.NotFound("NOT_FOUND", "order not found")
	}
	return s.orderView(ctx, o, true)
}

func (s *Service) AdminVerifyOrder(ctx context.Context, orderID int64, operator string) (*OrderView, error) {
	o, err := s.ent.VoucherOrder.Get(ctx, orderID)
	if err != nil {
		return nil, infraerrors.NotFound("NOT_FOUND", "order not found")
	}
	if o.Status != OrderStatusPaymentSubmitted {
		return nil, infraerrors.BadRequest("INVALID_STATUS", "order is not awaiting verification")
	}
	now := time.Now()
	o, err = s.ent.VoucherOrder.UpdateOneID(o.ID).
		SetStatus(OrderStatusPaymentVerified).
		SetVerifiedAt(now).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	_ = s.writeAudit(ctx, o.ID, "PAYMENT_VERIFIED", operator, nil)
	return s.fulfillOrder(ctx, o, operator)
}

func (s *Service) AdminRejectOrder(ctx context.Context, orderID int64, reason, operator string) (*OrderView, error) {
	o, err := s.ent.VoucherOrder.Get(ctx, orderID)
	if err != nil {
		return nil, infraerrors.NotFound("NOT_FOUND", "order not found")
	}
	if o.Status != OrderStatusPaymentSubmitted {
		return nil, infraerrors.BadRequest("INVALID_STATUS", "order cannot be rejected")
	}
	updated, err := s.ent.VoucherOrder.UpdateOneID(o.ID).
		SetStatus(OrderStatusRejected).
		SetRejectReason(strings.TrimSpace(reason)).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	_ = s.writeAudit(ctx, o.ID, "PAYMENT_REJECTED", operator, map[string]any{"reason": reason})
	return s.orderView(ctx, updated, false)
}

func (s *Service) AdminRetryFulfill(ctx context.Context, orderID int64, operator string) (*OrderView, error) {
	o, err := s.ent.VoucherOrder.Get(ctx, orderID)
	if err != nil {
		return nil, infraerrors.NotFound("NOT_FOUND", "order not found")
	}
	if o.Status != OrderStatusPaymentVerified && o.Status != OrderStatusFulfilling {
		return nil, infraerrors.BadRequest("INVALID_STATUS", "order cannot be fulfilled")
	}
	return s.fulfillOrder(ctx, o, operator)
}

func (s *Service) AdminSettings(ctx context.Context) (AdminSettingsView, error) {
	return s.store.AdminView(ctx)
}

func (s *Service) UpdateAdminSettings(ctx context.Context, in UpdateSettingsInput) (AdminSettingsView, error) {
	if err := s.store.Update(ctx, in); err != nil {
		return AdminSettingsView{}, err
	}
	return s.store.AdminView(ctx)
}

func (s *Service) LoadAPIConfig(ctx context.Context) (Config, error) {
	cfg, _, err := s.store.Load(ctx)
	return cfg, err
}

func (s *Service) fulfillOrder(ctx context.Context, o *dbent.VoucherOrder, operator string) (*OrderView, error) {
	cfg, _, err := s.store.Load(ctx)
	if err != nil {
		return nil, err
	}
	if !cfg.Configured() {
		return nil, infraerrors.BadRequest("KVOUCHER_NOT_CONFIGURED", "KVoucher API credentials not configured")
	}
	o, err = s.ent.VoucherOrder.UpdateOneID(o.ID).
		SetStatus(OrderStatusFulfilling).
		SetNillableFulfillError(nil).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	client := NewClient(cfg)
	pins, reqID, err := client.RetrieveStock(ctx, o.Denomination, o.Quantity, o.KvRetrieveReference)
	if err != nil {
		msg := err.Error()
		_, _ = s.ent.VoucherOrder.UpdateOneID(o.ID).
			SetStatus(OrderStatusPaymentVerified).
			SetFulfillError(msg).
			Save(ctx)
		_ = s.writeAudit(ctx, o.ID, "FULFILL_FAILED", operator, map[string]any{"error": msg, "request_id": reqID})
		return nil, infraerrors.BadRequest("FULFILL_FAILED", msg)
	}
	now := time.Now()
	for _, pin := range pins {
		enc, encErr := EncryptPIN(pin.PinCode)
		if encErr != nil {
			return nil, encErr
		}
		b := s.ent.VoucherPinDelivery.Create().
			SetOrderID(o.ID).
			SetPinCodeEnc(enc).
			SetSerial(pin.Serial).
			SetDenomination(pin.Denomination)
		if pin.ExpiresAt != "" {
			if t, parseErr := time.Parse("2006-01-02", pin.ExpiresAt); parseErr == nil {
				b.SetExpiresAt(t)
			}
		}
		if _, err = b.Save(ctx); err != nil {
			return nil, err
		}
	}
	updated, err := s.ent.VoucherOrder.UpdateOneID(o.ID).
		SetStatus(OrderStatusCompleted).
		SetFulfilledAt(now).
		SetCompletedAt(now).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	_ = s.syncProductStockAfterFulfill(ctx, o.Denomination, o.Quantity)
	_ = s.writeAudit(ctx, o.ID, "FULFILL_COMPLETED", operator, map[string]any{"request_id": reqID, "pin_count": len(pins)})
	return s.orderView(ctx, updated, true)
}

func (s *Service) syncProductStockAfterFulfill(ctx context.Context, denomination float64, qty int) error {
	p, err := s.ent.VoucherProduct.Query().
		Where(voucherproduct.DenominationEQ(denomination)).
		Only(ctx)
	if err != nil {
		return err
	}
	newStock := p.StockAvailable - qty
	if newStock < 0 {
		newStock = 0
	}
	_, err = s.ent.VoucherProduct.UpdateOneID(p.ID).SetStockAvailable(newStock).Save(ctx)
	return err
}

func (s *Service) getUserOrder(ctx context.Context, userID, orderID int64) (*dbent.VoucherOrder, error) {
	o, err := s.ent.VoucherOrder.Query().
		Where(voucherorder.IDEQ(orderID), voucherorder.UserIDEQ(userID)).
		Only(ctx)
	if err != nil {
		return nil, infraerrors.NotFound("NOT_FOUND", "order not found")
	}
	return o, nil
}

func (s *Service) orderView(ctx context.Context, o *dbent.VoucherOrder, includePins bool) (*OrderView, error) {
	_, runtime, err := s.store.Load(ctx)
	if err != nil {
		return nil, err
	}
	view := &OrderView{
		ID:           o.ID,
		OrderNo:      o.OrderNo,
		Status:       o.Status,
		ProductName:  o.ProductName,
		Denomination: o.Denomination,
		Quantity:     o.Quantity,
		Subtotal:     o.Subtotal,
		FeeAmount:    o.FeeAmount,
		TotalAmount:  o.TotalAmount,
		Currency:     o.Currency,
		ExpiresAt:    o.ExpiresAt,
		CreatedAt:    o.CreatedAt,
	}
	if o.PaymentRef != nil {
		view.PaymentRef = *o.PaymentRef
	}
	if o.RejectReason != nil {
		view.RejectReason = *o.RejectReason
	}
	if o.FulfillError != nil {
		view.FulfillError = *o.FulfillError
	}
	if o.CompletedAt != nil {
		view.CompletedAt = o.CompletedAt
	}
	if o.Status == OrderStatusPendingPayment {
		view.PaymentInfo = &PaymentInfo{
			Reference:    o.OrderNo,
			AmountDue:    o.TotalAmount,
			Currency:     o.Currency,
			BankAccounts: runtime.BankAccounts,
			Instructions: "Transfer the exact amount using the order number as reference",
		}
	}
	if includePins && o.Status == OrderStatusCompleted {
		deliveries, err := s.ent.VoucherPinDelivery.Query().
			Where(voucherpindelivery.OrderIDEQ(o.ID)).
			All(ctx)
		if err != nil {
			return nil, err
		}
		for _, d := range deliveries {
			pin, err := DecryptPIN(d.PinCodeEnc)
			if err != nil {
				return nil, err
			}
			pv := PinView{
				PinCode:      pin,
				Serial:       d.Serial,
				Denomination: d.Denomination,
			}
			if d.ExpiresAt != nil {
				pv.ExpiresAt = d.ExpiresAt.Format("2006-01-02")
			}
			view.Pins = append(view.Pins, pv)
		}
	}
	return view, nil
}

func (s *Service) writeAudit(ctx context.Context, orderID int64, action, operator string, metadata map[string]any) error {
	b := s.ent.VoucherAuditLog.Create().
		SetOrderID(orderID).
		SetAction(action).
		SetOperator(operator)
	if metadata != nil {
		b.SetMetadata(metadata)
	}
	_, err := b.Save(ctx)
	return err
}

func allocateOrderNo() string {
	now := time.Now()
	return fmt.Sprintf("VC-%s-%05d", now.Format("20060102"), now.Unix()%100000)
}

func roundMoney(v float64) float64 {
	return math.Round(v*100) / 100
}

func saveProofFile(name string, r io.Reader) (string, error) {
	dir := filepath.Join(setup.GetDataDir(), "voucher_proofs")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	safeName := filepath.Base(strings.ReplaceAll(name, "..", ""))
	filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), safeName)
	path := filepath.Join(dir, filename)
	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()
	if _, err := io.Copy(f, io.LimitReader(r, 5<<20)); err != nil {
		return "", err
	}
	return path, nil
}
