package service

import (
	"context"
	"testing"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/stretchr/testify/require"
)

type distributionPaymentRulesAttributionRepoStub struct {
	item *DistributionAttribution
	err  error
}

func (s *distributionPaymentRulesAttributionRepoStub) GetByUserID(ctx context.Context, userID int64) (*DistributionAttribution, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.item, nil
}

type distributionPaymentRulesOrganizationRepoStub struct {
	item *DistributionOrganization
	err  error
}

func (s *distributionPaymentRulesOrganizationRepoStub) GetByID(ctx context.Context, id int64) (*DistributionOrganization, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.item, nil
}

func (s *distributionPaymentRulesOrganizationRepoStub) GetByOwnerUserID(ctx context.Context, userID int64) (*DistributionOrganization, error) {
	return nil, nil
}

type distributionPaymentRulesOrderHistoryStub struct {
	hasCompleted bool
	err          error
}

func (s *distributionPaymentRulesOrderHistoryStub) HasCompletedBalanceOrder(ctx context.Context, userID int64) (bool, error) {
	if s.err != nil {
		return false, s.err
	}
	return s.hasCompleted, nil
}

func TestDistributionPaymentRuleServiceValidateBalanceRechargeRejectsBelowFirstRechargeMinimum(t *testing.T) {
	svc := NewDistributionPaymentRuleService(
		&distributionPaymentRulesAttributionRepoStub{
			item: &DistributionAttribution{UserID: 7, ChannelOrgID: 88},
		},
		&distributionPaymentRulesOrganizationRepoStub{
			item: &DistributionOrganization{
				ID: 88,
				Config: map[string]any{
					"first_recharge_min_amount": 100.0,
					"recharge_min_amount":       60.0,
				},
			},
		},
		&distributionPaymentRulesOrderHistoryStub{},
	)

	err := svc.ValidateBalanceRecharge(context.Background(), 7, 99)
	require.Error(t, err)
	appErr := infraerrors.FromError(err)
	require.Equal(t, "INVALID_AMOUNT", appErr.Reason)
	require.Equal(t, "100.00", appErr.Metadata["min"])
}

func TestDistributionPaymentRuleServiceValidateBalanceRechargeUsesRepeatRechargeMinimumAfterFirstOrder(t *testing.T) {
	svc := NewDistributionPaymentRuleService(
		&distributionPaymentRulesAttributionRepoStub{
			item: &DistributionAttribution{UserID: 7, ChannelOrgID: 88},
		},
		&distributionPaymentRulesOrganizationRepoStub{
			item: &DistributionOrganization{
				ID: 88,
				Config: map[string]any{
					"first_recharge_min_amount": 100.0,
					"recharge_min_amount":       60.0,
				},
			},
		},
		&distributionPaymentRulesOrderHistoryStub{hasCompleted: true},
	)

	require.NoError(t, svc.ValidateBalanceRecharge(context.Background(), 7, 60))
}

func TestDistributionPaymentRuleServiceApplyMethodLimitsRaisesSingleMinimums(t *testing.T) {
	svc := NewDistributionPaymentRuleService(
		&distributionPaymentRulesAttributionRepoStub{
			item: &DistributionAttribution{UserID: 7, ChannelOrgID: 88},
		},
		&distributionPaymentRulesOrganizationRepoStub{
			item: &DistributionOrganization{
				ID: 88,
				Config: map[string]any{
					"first_recharge_min_amount": 80.0,
					"recharge_min_amount":       50.0,
				},
			},
		},
		&distributionPaymentRulesOrderHistoryStub{},
	)

	out, err := svc.ApplyMethodLimits(context.Background(), 7, MethodLimitsResponse{
		Methods: map[string]MethodLimits{
			"wxpay":  {PaymentType: "wxpay", SingleMin: 10, SingleMax: 1000},
			"alipay": {PaymentType: "alipay", SingleMin: 30, SingleMax: 1000},
		},
		GlobalMin: 10,
		GlobalMax: 1000,
	})
	require.NoError(t, err)
	require.Equal(t, 80.0, out.Methods["wxpay"].SingleMin)
	require.Equal(t, 80.0, out.Methods["alipay"].SingleMin)
	require.Equal(t, 80.0, out.GlobalMin)
	require.Equal(t, 1000.0, out.GlobalMax)
}
