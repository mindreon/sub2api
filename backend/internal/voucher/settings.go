package voucher

// BankAccount is a platform collection account shown to users for bank transfer.
type BankAccount struct {
	ID            int    `json:"id"`
	BankName      string `json:"bank_name"`
	AccountName   string `json:"account_name"`
	AccountNumber string `json:"account_number"`
	Type          string `json:"type,omitempty"`
}

// RuntimeSettings are operational limits for checkout.
type RuntimeSettings struct {
	UIEnabled           bool
	OrderTimeoutHours   int
	MaxQuantityPerOrder int
	ReviewSLAHours      int
	FeeRate             float64
	HelpText            string
	BankAccounts        []BankAccount
	RetailMarkupPercent float64
}

// Order status constants (aligned with KVoucher B2B retail mapping).
const (
	OrderStatusPendingPayment   = "pending_payment"
	OrderStatusPaymentSubmitted = "payment_submitted"
	OrderStatusPaymentVerified  = "payment_verified"
	OrderStatusFulfilling       = "fulfilling"
	OrderStatusCompleted        = "completed"
	OrderStatusRejected         = "rejected"
	OrderStatusExpired          = "expired"
)
