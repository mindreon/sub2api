package service

import (
	"context"
	"errors"
	"strings"
)

type DistributionChannelBillingWalletRepository interface {
	GetByChannelOrgID(ctx context.Context, channelOrgID int64) (*DistributionWallet, error)
	ConsumeUsage(ctx context.Context, channelOrgID int64, amount float64, consumptionLimit float64) (*DistributionWallet, error)
	SyncStatus(ctx context.Context, channelOrgID int64, consumptionLimit float64) (*DistributionWallet, error)
}

type distributionChannelUsageConsumer interface {
	ConsumeUsage(ctx context.Context, usageLog *UsageLog) error
}

type DistributionChannelBillingService struct {
	attributionRepo DistributionAttributionLookupRepository
	orgRepo         DistributionOrganizationLookupRepository
	walletRepo      DistributionChannelBillingWalletRepository
}

func NewDistributionChannelBillingService(
	attributionRepo DistributionAttributionLookupRepository,
	orgRepo DistributionOrganizationLookupRepository,
	walletRepo DistributionChannelBillingWalletRepository,
) *DistributionChannelBillingService {
	return &DistributionChannelBillingService{
		attributionRepo: attributionRepo,
		orgRepo:         orgRepo,
		walletRepo:      walletRepo,
	}
}

func (s *DistributionChannelBillingService) CheckUserAccess(ctx context.Context, userID int64) error {
	channelOrgID, _, wallet, consumptionLimit, err := s.resolveUserChannelBilling(ctx, userID)
	if err != nil || wallet == nil {
		return err
	}
	if synced, syncErr := s.walletRepo.SyncStatus(ctx, channelOrgID, consumptionLimit); syncErr == nil && synced != nil {
		wallet = synced
	}
	if consumptionLimit > 0 && wallet.TotalConsumed >= consumptionLimit {
		return ErrDistributionChannelConsumptionLimitExceeded
	}
	if available := wallet.PrepaidBalance - wallet.CommissionReserved; available <= 0 {
		return ErrDistributionWalletInsufficientBalance
	}
	if NormalizeDistributionWalletStatus(wallet.Status) != "active" {
		return ErrDistributionChannelSuspended
	}
	return nil
}

func (s *DistributionChannelBillingService) ConsumeUsage(ctx context.Context, usageLog *UsageLog) error {
	if usageLog == nil || usageLog.UserID <= 0 {
		return nil
	}
	channelOrgID, _, _, consumptionLimit, err := s.resolveUserChannelBilling(ctx, usageLog.UserID)
	if err != nil || channelOrgID <= 0 {
		return err
	}

	amount := distributionCommissionBaseAmount(usageLog)
	if amount <= 0 {
		return nil
	}

	_, err = s.walletRepo.ConsumeUsage(ctx, channelOrgID, amount, consumptionLimit)
	if err != nil {
		if errors.Is(err, ErrDistributionWalletInsufficientBalance) || errors.Is(err, ErrDistributionChannelConsumptionLimitExceeded) {
			_, _ = s.walletRepo.SyncStatus(ctx, channelOrgID, consumptionLimit)
		}
		return err
	}
	return nil
}

func (s *DistributionChannelBillingService) resolveUserChannelBilling(ctx context.Context, userID int64) (int64, *DistributionOrganization, *DistributionWallet, float64, error) {
	if s == nil || s.attributionRepo == nil || s.orgRepo == nil || s.walletRepo == nil || userID <= 0 {
		return 0, nil, nil, 0, nil
	}
	attribution, err := s.attributionRepo.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrDistributionAttributionNotFound) {
			return 0, nil, nil, 0, nil
		}
		return 0, nil, nil, 0, err
	}
	if attribution == nil || attribution.ChannelOrgID <= 0 {
		return 0, nil, nil, 0, nil
	}

	org, err := s.orgRepo.GetByID(ctx, attribution.ChannelOrgID)
	if err != nil {
		return 0, nil, nil, 0, err
	}
	if !shouldAffectDistributionWallet(org) {
		return 0, org, nil, 0, nil
	}
	if org == nil || !strings.EqualFold(strings.TrimSpace(org.Status), "active") {
		return attribution.ChannelOrgID, org, nil, 0, ErrDistributionChannelSuspended
	}

	wallet, err := s.walletRepo.GetByChannelOrgID(ctx, attribution.ChannelOrgID)
	if err != nil {
		return 0, nil, nil, 0, err
	}
	consumptionLimit := distributionOrganizationConfigFloat(org.Config, "consumption_limit", "credit_limit", "usage_limit")
	return attribution.ChannelOrgID, org, wallet, consumptionLimit, nil
}
