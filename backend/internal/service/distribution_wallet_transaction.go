package service

import (
	"context"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

var (
	ErrInvalidDistributionWalletTransaction = infraerrors.BadRequest(
		"INVALID_DISTRIBUTION_WALLET_TRANSACTION",
		"invalid distribution wallet transaction",
	)
)

type DistributionWalletTransaction struct {
	ID                       int64     `json:"id"`
	ChannelOrgID             int64     `json:"channel_org_id"`
	OrganizationName         string    `json:"organization_name"`
	OrganizationType         string    `json:"organization_type"`
	TransactionType          string    `json:"transaction_type"`
	Amount                   float64   `json:"amount"`
	PrepaidBalanceBefore     float64   `json:"prepaid_balance_before"`
	PrepaidBalanceAfter      float64   `json:"prepaid_balance_after"`
	CommissionReservedBefore float64   `json:"commission_reserved_before"`
	CommissionReservedAfter  float64   `json:"commission_reserved_after"`
	ReferenceNo              string    `json:"reference_no"`
	Note                     string    `json:"note"`
	OperatorUserID           *int64    `json:"operator_user_id,omitempty"`
	CreatedAt                time.Time `json:"created_at"`
}

type DistributionWalletRechargeInput struct {
	Amount         float64
	ReferenceNo    string
	Note           string
	OperatorUserID *int64
}

type DistributionWalletTransactionListRepository interface {
	ListTransactions(ctx context.Context, filter DistributionAdminListFilter, params pagination.PaginationParams) ([]DistributionWalletTransaction, *pagination.PaginationResult, error)
	ListTransactionsByChannelOrgID(ctx context.Context, channelOrgID int64, params pagination.PaginationParams, transactionType string) ([]DistributionWalletTransaction, *pagination.PaginationResult, error)
}

func normalizeDistributionWalletTransactionType(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "recharge", "refund", "consume", "commission_reserve", "commission_release", "commission_settle", "commission_deduct", "commission_refund":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return ""
	}
}
