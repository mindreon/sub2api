package service

import (
	"context"
	"errors"
	"fmt"
	"math"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/paymentorder"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

type DistributionPaymentRuleResolver interface {
	ValidateBalanceRecharge(ctx context.Context, userID int64, amount float64) error
	ApplyMethodLimits(ctx context.Context, userID int64, input MethodLimitsResponse) (MethodLimitsResponse, error)
}

type DistributionPaymentOrderHistoryRepository interface {
	HasCompletedBalanceOrder(ctx context.Context, userID int64) (bool, error)
}

type distributionPaymentOrderHistoryRepository struct {
	client *dbent.Client
}

func NewDistributionPaymentOrderHistoryRepository(client *dbent.Client) DistributionPaymentOrderHistoryRepository {
	return &distributionPaymentOrderHistoryRepository{client: client}
}

func (r *distributionPaymentOrderHistoryRepository) HasCompletedBalanceOrder(ctx context.Context, userID int64) (bool, error) {
	if r == nil || r.client == nil || userID <= 0 {
		return false, nil
	}
	return r.client.PaymentOrder.Query().
		Where(
			paymentorder.UserIDEQ(userID),
			paymentorder.OrderTypeEQ(payment.OrderTypeBalance),
			paymentorder.StatusIn(
				OrderStatusCompleted,
				OrderStatusRefundRequested,
				OrderStatusRefunding,
				OrderStatusPartiallyRefunded,
				OrderStatusRefunded,
				OrderStatusRefundFailed,
			),
		).
		Exist(ctx)
}

type DistributionPaymentRuleService struct {
	attributionRepo  DistributionAttributionLookupRepository
	organizationRepo DistributionOrganizationLookupRepository
	orderHistoryRepo DistributionPaymentOrderHistoryRepository
}

func NewDistributionPaymentRuleService(
	attributionRepo DistributionAttributionLookupRepository,
	organizationRepo DistributionOrganizationLookupRepository,
	orderHistoryRepo DistributionPaymentOrderHistoryRepository,
) *DistributionPaymentRuleService {
	return &DistributionPaymentRuleService{
		attributionRepo:  attributionRepo,
		organizationRepo: organizationRepo,
		orderHistoryRepo: orderHistoryRepo,
	}
}

func (s *DistributionPaymentRuleService) ValidateBalanceRecharge(ctx context.Context, userID int64, amount float64) error {
	minAmount, err := s.resolveBalanceRechargeMinimum(ctx, userID)
	if err != nil || minAmount <= 0 {
		return err
	}
	if amount+1e-9 < minAmount {
		return infraerrors.BadRequest("INVALID_AMOUNT", "amount out of range").
			WithMetadata(map[string]string{"min": fmt.Sprintf("%.2f", minAmount)})
	}
	return nil
}

func (s *DistributionPaymentRuleService) ApplyMethodLimits(ctx context.Context, userID int64, input MethodLimitsResponse) (MethodLimitsResponse, error) {
	minAmount, err := s.resolveBalanceRechargeMinimum(ctx, userID)
	if err != nil || minAmount <= 0 || len(input.Methods) == 0 {
		return input, err
	}

	out := MethodLimitsResponse{
		Methods:   make(map[string]MethodLimits, len(input.Methods)),
		GlobalMin: input.GlobalMin,
		GlobalMax: input.GlobalMax,
	}
	for key, limit := range input.Methods {
		if limit.SingleMin < minAmount {
			limit.SingleMin = minAmount
		}
		out.Methods[key] = limit
	}
	out.GlobalMin, out.GlobalMax = pcComputeGlobalRange(out.Methods)
	return out, nil
}

func (s *DistributionPaymentRuleService) resolveBalanceRechargeMinimum(ctx context.Context, userID int64) (float64, error) {
	if s == nil || s.attributionRepo == nil || s.organizationRepo == nil || userID <= 0 {
		return 0, nil
	}

	attribution, err := s.attributionRepo.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrDistributionAttributionNotFound) {
			return 0, nil
		}
		return 0, err
	}
	if attribution == nil || attribution.ChannelOrgID <= 0 {
		return 0, nil
	}

	org, err := s.organizationRepo.GetByID(ctx, attribution.ChannelOrgID)
	if err != nil {
		return 0, err
	}
	if org == nil {
		return 0, nil
	}

	firstMin := distributionOrganizationConfigFloat(org.Config, "first_recharge_min_amount")
	rechargeMin := distributionOrganizationConfigFloat(org.Config, "recharge_min_amount", "single_recharge_min_amount")
	if firstMin <= 0 && rechargeMin <= 0 {
		return 0, nil
	}

	hasCompleted := false
	if s.orderHistoryRepo != nil {
		completed, err := s.orderHistoryRepo.HasCompletedBalanceOrder(ctx, userID)
		if err != nil {
			return 0, err
		}
		hasCompleted = completed
	}
	if hasCompleted {
		return rechargeMin, nil
	}
	return math.Max(firstMin, rechargeMin), nil
}
