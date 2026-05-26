package service

import (
	"context"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

var (
	ErrInvalidDistributionStats = infraerrors.BadRequest("INVALID_DISTRIBUTION_STATS", "invalid distribution stats")
)

type DistributionAdminStats struct {
	OrganizationCount         int64   `json:"organization_count"`
	PlatformCount             int64   `json:"platform_count"`
	ResellerCount             int64   `json:"reseller_count"`
	OemCount                  int64   `json:"oem_count"`
	MemberCount               int64   `json:"member_count"`
	AgentCount                int64   `json:"agent_count"`
	Kol1Count                 int64   `json:"kol1_count"`
	Kol2Count                 int64   `json:"kol2_count"`
	PromotionLinkCount        int64   `json:"promotion_link_count"`
	AttributionCount          int64   `json:"attribution_count"`
	CommissionCount           int64   `json:"commission_count"`
	WalletCount               int64   `json:"wallet_count"`
	PrepaidBalanceTotal       float64 `json:"prepaid_balance_total"`
	CommissionReservedTotal   float64 `json:"commission_reserved_total"`
	TotalRecharged            float64 `json:"total_recharged"`
	TotalConsumed             float64 `json:"total_consumed"`
	FrozenCommissionAmount    float64 `json:"frozen_commission_amount"`
	AvailableCommissionAmount float64 `json:"available_commission_amount"`
	SettledCommissionAmount   float64 `json:"settled_commission_amount"`
	CommissionExpenseRatio    float64 `json:"commission_expense_ratio"`
	CommissionUpperRatio      float64 `json:"commission_upper_ratio"`
}

type DistributionChannelSummary struct {
	Organization              DistributionOrganization `json:"organization"`
	Wallet                    DistributionWallet       `json:"wallet"`
	MemberCount               int64                    `json:"member_count"`
	AgentCount                int64                    `json:"agent_count"`
	Kol1Count                 int64                    `json:"kol1_count"`
	Kol2Count                 int64                    `json:"kol2_count"`
	PromotionLinkCount        int64                    `json:"promotion_link_count"`
	AttributionCount          int64                    `json:"attribution_count"`
	CommissionCount           int64                    `json:"commission_count"`
	FrozenCommissionAmount    float64                  `json:"frozen_commission_amount"`
	AvailableCommissionAmount float64                  `json:"available_commission_amount"`
	SettledCommissionAmount   float64                  `json:"settled_commission_amount"`
}

type DistributionStatsRepository interface {
	GetAdminStats(ctx context.Context) (*DistributionAdminStats, error)
	GetChannelSummary(ctx context.Context, channelOrgID int64) (*DistributionChannelSummary, error)
}
