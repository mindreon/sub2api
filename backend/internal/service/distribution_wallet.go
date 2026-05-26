package service

import (
	"context"
	"math"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

var (
	ErrInvalidDistributionWallet             = infraerrors.BadRequest("INVALID_DISTRIBUTION_WALLET", "invalid distribution wallet")
	ErrDistributionWalletInsufficientBalance = infraerrors.BadRequest(
		"DISTRIBUTION_WALLET_INSUFFICIENT_BALANCE",
		"distribution wallet has insufficient available balance",
	)
	ErrDistributionWalletInsufficientReserved = infraerrors.BadRequest(
		"DISTRIBUTION_WALLET_INSUFFICIENT_RESERVED",
		"distribution wallet has insufficient reserved commission",
	)
	ErrDistributionChannelSuspended = infraerrors.Forbidden(
		"DISTRIBUTION_CHANNEL_SUSPENDED",
		"distribution channel is suspended",
	)
	ErrDistributionChannelConsumptionLimitExceeded = infraerrors.Forbidden(
		"DISTRIBUTION_CHANNEL_CONSUMPTION_LIMIT_EXCEEDED",
		"distribution channel consumption limit exceeded",
	)
)

type DistributionWallet struct {
	ChannelOrgID       int64     `json:"channel_org_id"`
	OrganizationName   string    `json:"organization_name"`
	OrganizationType   string    `json:"organization_type"`
	PrepaidBalance     float64   `json:"prepaid_balance"`
	CommissionReserved float64   `json:"commission_reserved"`
	TotalRecharged     float64   `json:"total_recharged"`
	TotalConsumed      float64   `json:"total_consumed"`
	WarningThreshold   float64   `json:"warning_threshold"`
	Status             string    `json:"status"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type DistributionWalletListRepository interface {
	List(ctx context.Context, filter DistributionAdminListFilter, params pagination.PaginationParams) ([]DistributionWallet, *pagination.PaginationResult, error)
	GetByChannelOrgID(ctx context.Context, channelOrgID int64) (*DistributionWallet, error)
	UpdateWarningThreshold(ctx context.Context, channelOrgID int64, warningThreshold float64) (*DistributionWallet, error)
}

type DistributionWalletMutationRepository interface {
	ReserveCommission(ctx context.Context, channelOrgID int64, amount float64) (*DistributionWallet, error)
	ReleaseCommission(ctx context.Context, channelOrgID int64, amount float64) (*DistributionWallet, error)
	SettleReservedCommission(ctx context.Context, channelOrgID int64, amount float64) (*DistributionWallet, error)
	DeductCommission(ctx context.Context, channelOrgID int64, amount float64) (*DistributionWallet, error)
	RefundCommission(ctx context.Context, channelOrgID int64, amount float64) (*DistributionWallet, error)
	ConsumeUsage(ctx context.Context, channelOrgID int64, amount float64, consumptionLimit float64) (*DistributionWallet, error)
}

type DistributionWalletRefundInput struct {
	Amount         float64
	ReferenceNo    string
	Note           string
	OperatorUserID *int64
}

type DistributionWalletRefundResult struct {
	Wallet        *DistributionWallet `json:"wallet"`
	RefundAmount  float64             `json:"refund_amount"`
	FeeRate       float64             `json:"fee_rate"`
	FeeAmount     float64             `json:"fee_amount"`
	NetAmount     float64             `json:"net_amount"`
	ReferenceNo   string              `json:"reference_no"`
	Note          string              `json:"note"`
	ProcessedMock bool                `json:"processed_mock"`
}

func roundDistributionWalletAmount(value float64) float64 {
	return math.Round(value*1e8) / 1e8
}

func NormalizeDistributionWalletStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "active", "inactive", "disabled":
		return strings.ToLower(strings.TrimSpace(status))
	default:
		return "active"
	}
}

func ResolveDistributionWalletStatus(currentStatus string, prepaidBalance float64, commissionReserved float64, totalConsumed float64, consumptionLimit float64) string {
	status := NormalizeDistributionWalletStatus(currentStatus)
	if status == "disabled" {
		return status
	}
	if roundDistributionWalletAmount(prepaidBalance-commissionReserved) <= 0 {
		return "inactive"
	}
	if consumptionLimit > 0 && roundDistributionWalletAmount(totalConsumed) >= roundDistributionWalletAmount(consumptionLimit) {
		return "inactive"
	}
	return "active"
}
