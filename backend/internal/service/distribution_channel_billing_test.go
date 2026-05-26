package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type distributionChannelBillingWalletRepoStub struct {
	wallet         *DistributionWallet
	consumedAmount float64
	consumedLimit  float64
	consumedOrgID  int64
	consumeErr     error
	syncedLimit    float64
	syncedOrgID    int64
}

func (s *distributionChannelBillingWalletRepoStub) GetByChannelOrgID(ctx context.Context, channelOrgID int64) (*DistributionWallet, error) {
	return s.wallet, nil
}

func (s *distributionChannelBillingWalletRepoStub) ConsumeUsage(ctx context.Context, channelOrgID int64, amount float64, consumptionLimit float64) (*DistributionWallet, error) {
	s.consumedOrgID = channelOrgID
	s.consumedAmount = amount
	s.consumedLimit = consumptionLimit
	if s.consumeErr != nil {
		return nil, s.consumeErr
	}
	if s.wallet != nil {
		s.wallet.PrepaidBalance -= amount
		s.wallet.TotalConsumed += amount
	}
	return s.wallet, nil
}

func (s *distributionChannelBillingWalletRepoStub) SyncStatus(ctx context.Context, channelOrgID int64, consumptionLimit float64) (*DistributionWallet, error) {
	s.syncedOrgID = channelOrgID
	s.syncedLimit = consumptionLimit
	if s.wallet != nil {
		s.wallet.Status = ResolveDistributionWalletStatus(
			s.wallet.Status,
			s.wallet.PrepaidBalance,
			s.wallet.CommissionReserved,
			s.wallet.TotalConsumed,
			consumptionLimit,
		)
	}
	return s.wallet, nil
}

func TestDistributionChannelBillingService_CheckUserAccessRejectsExhaustedWallet(t *testing.T) {
	memberID := int64(42)
	svc := NewDistributionChannelBillingService(
		&distributionCommissionAttributionRepoStub{
			attribution: &DistributionAttribution{
				UserID:           7,
				ChannelOrgID:     99,
				ReferrerMemberID: &memberID,
			},
		},
		&distributionCommissionOrgRepoStub{
			org: &DistributionOrganization{
				ID:     99,
				Type:   "reseller",
				Status: "active",
			},
		},
		&distributionChannelBillingWalletRepoStub{
			wallet: &DistributionWallet{
				ChannelOrgID:       99,
				OrganizationType:   "reseller",
				PrepaidBalance:     10,
				CommissionReserved: 10,
				Status:             "active",
			},
		},
	)

	err := svc.CheckUserAccess(context.Background(), 7)
	require.ErrorIs(t, err, ErrDistributionWalletInsufficientBalance)
}

func TestDistributionChannelBillingService_CheckUserAccessRejectsConsumptionLimitExceeded(t *testing.T) {
	memberID := int64(42)
	svc := NewDistributionChannelBillingService(
		&distributionCommissionAttributionRepoStub{
			attribution: &DistributionAttribution{
				UserID:           7,
				ChannelOrgID:     99,
				ReferrerMemberID: &memberID,
			},
		},
		&distributionCommissionOrgRepoStub{
			org: &DistributionOrganization{
				ID:     99,
				Type:   "reseller",
				Status: "active",
				Config: map[string]any{"consumption_limit": 100.0},
			},
		},
		&distributionChannelBillingWalletRepoStub{
			wallet: &DistributionWallet{
				ChannelOrgID:       99,
				OrganizationType:   "reseller",
				PrepaidBalance:     200,
				CommissionReserved: 10,
				TotalConsumed:      100,
				Status:             "active",
			},
		},
	)

	err := svc.CheckUserAccess(context.Background(), 7)
	require.ErrorIs(t, err, ErrDistributionChannelConsumptionLimitExceeded)
}

func TestDistributionChannelBillingService_ConsumeUsageUsesAccountStatsCostAndConsumptionLimit(t *testing.T) {
	memberID := int64(42)
	walletRepo := &distributionChannelBillingWalletRepoStub{
		wallet: &DistributionWallet{
			ChannelOrgID:       99,
			OrganizationType:   "reseller",
			PrepaidBalance:     200,
			CommissionReserved: 10,
			TotalConsumed:      30,
			Status:             "active",
		},
	}
	accountStatsCost := 12.5
	svc := NewDistributionChannelBillingService(
		&distributionCommissionAttributionRepoStub{
			attribution: &DistributionAttribution{
				UserID:           7,
				ChannelOrgID:     99,
				ReferrerMemberID: &memberID,
			},
		},
		&distributionCommissionOrgRepoStub{
			org: &DistributionOrganization{
				ID:     99,
				Type:   "reseller",
				Status: "active",
				Config: map[string]any{"consumption_limit": 100.0},
			},
		},
		walletRepo,
	)

	err := svc.ConsumeUsage(context.Background(), &UsageLog{
		UserID:           7,
		TotalCost:        20,
		AccountStatsCost: &accountStatsCost,
	})
	require.NoError(t, err)
	require.Equal(t, int64(99), walletRepo.consumedOrgID)
	require.InDelta(t, 12.5, walletRepo.consumedAmount, 0.0001)
	require.InDelta(t, 100.0, walletRepo.consumedLimit, 0.0001)
}
