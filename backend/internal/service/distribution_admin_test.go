package service

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

type distributionAdminOrgRepoStub struct {
	params pagination.PaginationParams
	item   *DistributionOrganization
}

func (s *distributionAdminOrgRepoStub) List(ctx context.Context, params pagination.PaginationParams) ([]DistributionOrganization, *pagination.PaginationResult, error) {
	s.params = params
	now := time.Now().UTC()
	return []DistributionOrganization{{ID: 88, Type: "reseller", Name: "Agent", Status: "active", CreatedAt: now, UpdatedAt: now}}, &pagination.PaginationResult{Total: 1, Page: params.Page, PageSize: params.PageSize, Pages: 1}, nil
}

func (s *distributionAdminOrgRepoStub) GetByID(ctx context.Context, id int64) (*DistributionOrganization, error) {
	if s.item != nil {
		return s.item, nil
	}
	now := time.Now().UTC()
	return &DistributionOrganization{ID: id, Type: "reseller", Name: "Agent", Status: "active", CreatedAt: now, UpdatedAt: now}, nil
}

type distributionAdminMemberRepoStub struct {
	filter DistributionAdminListFilter
	params pagination.PaginationParams
}

func (s *distributionAdminMemberRepoStub) ListAdmin(ctx context.Context, filter DistributionAdminListFilter, params pagination.PaginationParams) ([]DistributionMemberView, *pagination.PaginationResult, error) {
	s.filter = filter
	s.params = params
	return []DistributionMemberView{{MemberID: 42, ChannelOrgID: filter.ChannelOrgID, RoleType: filter.RoleType}}, &pagination.PaginationResult{Total: 1, Page: params.Page, PageSize: params.PageSize, Pages: 1}, nil
}

type distributionAdminAttributionRepoStub struct {
	filter DistributionAdminListFilter
	params pagination.PaginationParams
	item   *DistributionAttribution
	input  *DistributionAttributionAdminUpdateInput
	audits []DistributionAttributionAuditView
}

func (s *distributionAdminAttributionRepoStub) ListAdmin(ctx context.Context, filter DistributionAdminListFilter, params pagination.PaginationParams) ([]DistributionAttributionView, *pagination.PaginationResult, error) {
	s.filter = filter
	s.params = params
	return []DistributionAttributionView{{UserID: 7, ChannelOrgID: filter.ChannelOrgID}}, &pagination.PaginationResult{Total: 1, Page: params.Page, PageSize: params.PageSize, Pages: 1}, nil
}

func (s *distributionAdminAttributionRepoStub) ListAuditsAdmin(ctx context.Context, filter DistributionAdminListFilter, params pagination.PaginationParams) ([]DistributionAttributionAuditView, *pagination.PaginationResult, error) {
	s.filter = filter
	s.params = params
	if s.audits != nil {
		return s.audits, &pagination.PaginationResult{Total: int64(len(s.audits)), Page: params.Page, PageSize: params.PageSize, Pages: 1}, nil
	}
	return []DistributionAttributionAuditView{{ID: 1, UserID: filter.UserID, NewChannelOrgID: filter.ChannelOrgID}}, &pagination.PaginationResult{Total: 1, Page: params.Page, PageSize: params.PageSize, Pages: 1}, nil
}

func (s *distributionAdminAttributionRepoStub) GetByUserID(ctx context.Context, userID int64) (*DistributionAttribution, error) {
	if s.item != nil {
		return s.item, nil
	}
	now := time.Now().UTC()
	return &DistributionAttribution{
		UserID:       userID,
		ChannelOrgID: 88,
		BoundAt:      now,
		BoundSource:  "registration",
		BoundBy:      "system",
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

func (s *distributionAdminAttributionRepoStub) UpdateByAdmin(ctx context.Context, input DistributionAttributionAdminUpdateInput) (*DistributionAttribution, error) {
	s.input = &input
	now := time.Now().UTC()
	return &DistributionAttribution{
		UserID:           input.UserID,
		ChannelOrgID:     input.ChannelOrgID,
		ReferrerMemberID: input.ReferrerMemberID,
		PromotionLinkID:  input.PromotionLinkID,
		BoundAt:          now,
		BoundSource:      input.BoundSource,
		BoundBy:          input.BoundBy,
		CreatedAt:        now,
		UpdatedAt:        now,
	}, nil
}

type distributionAdminCommissionRepoStub struct {
	filter DistributionAdminListFilter
	params pagination.PaginationParams
}

func (s *distributionAdminCommissionRepoStub) ListAdmin(ctx context.Context, filter DistributionAdminListFilter, params pagination.PaginationParams) ([]DistributionCommissionLedgerView, *pagination.PaginationResult, error) {
	s.filter = filter
	s.params = params
	return []DistributionCommissionLedgerView{{ID: 1001, ChannelOrgID: filter.ChannelOrgID}}, &pagination.PaginationResult{Total: 1, Page: params.Page, PageSize: params.PageSize, Pages: 1}, nil
}

type distributionAdminWalletRepoStub struct {
	wallet                  *DistributionWallet
	rechargeInput           *DistributionWalletRechargeInput
	refundInput             *DistributionWalletRefundInput
	reserveCalls            []distributionWalletMutationCall
	releaseCalls            []distributionWalletMutationCall
	settleReservedCalls     []distributionWalletMutationCall
	deductCalls             []distributionWalletMutationCall
	refundCalls             []distributionWalletMutationCall
	warningThresholdUpdated float64
}

func (s *distributionAdminWalletRepoStub) List(ctx context.Context, filter DistributionAdminListFilter, params pagination.PaginationParams) ([]DistributionWallet, *pagination.PaginationResult, error) {
	if s.wallet == nil {
		return nil, &pagination.PaginationResult{}, nil
	}
	return []DistributionWallet{*s.wallet}, &pagination.PaginationResult{Total: 1, Page: params.Page, PageSize: params.PageSize, Pages: 1}, nil
}

func (s *distributionAdminWalletRepoStub) GetByChannelOrgID(ctx context.Context, channelOrgID int64) (*DistributionWallet, error) {
	if s.wallet == nil {
		return nil, ErrInvalidDistributionWallet
	}
	return s.wallet, nil
}

func (s *distributionAdminWalletRepoStub) UpdateWarningThreshold(ctx context.Context, channelOrgID int64, warningThreshold float64) (*DistributionWallet, error) {
	s.warningThresholdUpdated = warningThreshold
	if s.wallet == nil {
		s.wallet = &DistributionWallet{ChannelOrgID: channelOrgID}
	}
	s.wallet.WarningThreshold = warningThreshold
	return s.wallet, nil
}

func (s *distributionAdminWalletRepoStub) Recharge(ctx context.Context, channelOrgID int64, input DistributionWalletRechargeInput) (*DistributionWallet, error) {
	s.rechargeInput = &input
	if s.wallet == nil {
		s.wallet = &DistributionWallet{ChannelOrgID: channelOrgID}
	}
	s.wallet.PrepaidBalance += input.Amount
	s.wallet.TotalRecharged += input.Amount
	return s.wallet, nil
}

func (s *distributionAdminWalletRepoStub) RefundPrepaidBalance(ctx context.Context, channelOrgID int64, input DistributionWalletRefundInput) (*DistributionWallet, error) {
	s.refundInput = &input
	if s.wallet == nil {
		s.wallet = &DistributionWallet{ChannelOrgID: channelOrgID}
	}
	s.wallet.PrepaidBalance -= input.Amount
	return s.wallet, nil
}

func (s *distributionAdminWalletRepoStub) ReserveCommission(ctx context.Context, channelOrgID int64, amount float64) (*DistributionWallet, error) {
	s.reserveCalls = append(s.reserveCalls, distributionWalletMutationCall{channelOrgID: channelOrgID, amount: amount})
	return s.wallet, nil
}

func (s *distributionAdminWalletRepoStub) ReleaseCommission(ctx context.Context, channelOrgID int64, amount float64) (*DistributionWallet, error) {
	s.releaseCalls = append(s.releaseCalls, distributionWalletMutationCall{channelOrgID: channelOrgID, amount: amount})
	return s.wallet, nil
}

func (s *distributionAdminWalletRepoStub) SettleReservedCommission(ctx context.Context, channelOrgID int64, amount float64) (*DistributionWallet, error) {
	s.settleReservedCalls = append(s.settleReservedCalls, distributionWalletMutationCall{channelOrgID: channelOrgID, amount: amount})
	return s.wallet, nil
}

func (s *distributionAdminWalletRepoStub) DeductCommission(ctx context.Context, channelOrgID int64, amount float64) (*DistributionWallet, error) {
	s.deductCalls = append(s.deductCalls, distributionWalletMutationCall{channelOrgID: channelOrgID, amount: amount})
	return s.wallet, nil
}

func (s *distributionAdminWalletRepoStub) RefundCommission(ctx context.Context, channelOrgID int64, amount float64) (*DistributionWallet, error) {
	s.refundCalls = append(s.refundCalls, distributionWalletMutationCall{channelOrgID: channelOrgID, amount: amount})
	return s.wallet, nil
}

type distributionAdminSettlementRepoStub struct {
	ledgerByID            *DistributionCommissionLedger
	settleResult          *DistributionCommissionLedger
	settleToBalanceResult *DistributionCommissionLedger
	reverseResult         *DistributionCommissionLedger
	reverseBalanceResult  *DistributionCommissionLedger
	settleID              int64
	settleInput           DistributionCommissionSettlementInput
	settleToBalanceID     int64
	settleToBalanceInput  DistributionCommissionSettlementInput
	reverseID             int64
	reverseBalanceID      int64
}

type distributionAdminWalletTransactionRepoStub struct {
	filter DistributionAdminListFilter
	params pagination.PaginationParams
}

func (s *distributionAdminWalletTransactionRepoStub) ListTransactions(ctx context.Context, filter DistributionAdminListFilter, params pagination.PaginationParams) ([]DistributionWalletTransaction, *pagination.PaginationResult, error) {
	s.filter = filter
	s.params = params
	return []DistributionWalletTransaction{{ID: 1, ChannelOrgID: filter.ChannelOrgID, TransactionType: filter.TransactionType}}, &pagination.PaginationResult{Total: 1, Page: params.Page, PageSize: params.PageSize, Pages: 1}, nil
}

func (s *distributionAdminWalletTransactionRepoStub) ListTransactionsByChannelOrgID(ctx context.Context, channelOrgID int64, params pagination.PaginationParams, transactionType string) ([]DistributionWalletTransaction, *pagination.PaginationResult, error) {
	return nil, nil, nil
}

func (s *distributionAdminSettlementRepoStub) GetByID(ctx context.Context, commissionID int64) (*DistributionCommissionLedger, error) {
	return s.ledgerByID, nil
}

func (s *distributionAdminSettlementRepoStub) Settle(ctx context.Context, commissionID int64, input DistributionCommissionSettlementInput) (*DistributionCommissionLedger, error) {
	s.settleID = commissionID
	s.settleInput = input
	return s.settleResult, nil
}

func (s *distributionAdminSettlementRepoStub) SettleToBalance(ctx context.Context, commissionID int64, input DistributionCommissionSettlementInput) (*DistributionCommissionLedger, error) {
	s.settleToBalanceID = commissionID
	s.settleToBalanceInput = input
	return s.settleToBalanceResult, nil
}

func (s *distributionAdminSettlementRepoStub) Reverse(ctx context.Context, commissionID int64) (*DistributionCommissionLedger, error) {
	s.reverseID = commissionID
	return s.reverseResult, nil
}

func (s *distributionAdminSettlementRepoStub) ReverseBalanceSettlement(ctx context.Context, commissionID int64) (*DistributionCommissionLedger, error) {
	s.reverseBalanceID = commissionID
	return s.reverseBalanceResult, nil
}

func TestDistributionAdminServiceListsAdminResources(t *testing.T) {
	orgRepo := &distributionAdminOrgRepoStub{}
	memberRepo := &distributionAdminMemberRepoStub{}
	attrRepo := &distributionAdminAttributionRepoStub{}
	commissionRepo := &distributionAdminCommissionRepoStub{}
	svc := NewDistributionAdminService(orgRepo, memberRepo, attrRepo, commissionRepo, nil, nil, nil)

	params := pagination.PaginationParams{Page: 2, PageSize: 10}
	filter := DistributionAdminListFilter{ChannelOrgID: 88, RoleType: "agent"}

	orgs, orgPage, err := svc.ListOrganizations(context.Background(), params)
	require.NoError(t, err)
	require.Len(t, orgs, 1)
	require.Equal(t, 2, orgRepo.params.Page)
	require.Equal(t, int64(1), orgPage.Total)

	members, _, err := svc.ListMembers(context.Background(), filter, params)
	require.NoError(t, err)
	require.Len(t, members, 1)
	require.Equal(t, filter, memberRepo.filter)

	attributions, _, err := svc.ListAttributions(context.Background(), filter, params)
	require.NoError(t, err)
	require.Len(t, attributions, 1)
	require.Equal(t, filter, attrRepo.filter)

	commissions, _, err := svc.ListCommissions(context.Background(), filter, params)
	require.NoError(t, err)
	require.Len(t, commissions, 1)
	require.Equal(t, filter, commissionRepo.filter)
}

func TestDistributionAdminServiceUpdateAttributionForcesManualAdminBinding(t *testing.T) {
	attrRepo := &distributionAdminAttributionRepoStub{}
	svc := NewDistributionAdminService(nil, nil, attrRepo, nil, nil, nil, nil)
	operatorUserID := int64(9)
	referrerMemberID := int64(101)
	promotionLinkID := int64(202)

	out, err := svc.UpdateAttribution(context.Background(), 7, DistributionAttributionAdminUpdateInput{
		ChannelOrgID:     88,
		ReferrerMemberID: &referrerMemberID,
		PromotionLinkID:  &promotionLinkID,
		OperatorUserID:   &operatorUserID,
		Note:             "manual reassignment",
		BoundSource:      "oauth",
		BoundBy:          "system",
	})
	require.NoError(t, err)
	require.NotNil(t, out)
	require.NotNil(t, attrRepo.input)
	require.Equal(t, int64(7), attrRepo.input.UserID)
	require.Equal(t, int64(88), attrRepo.input.ChannelOrgID)
	require.Equal(t, "manual", attrRepo.input.BoundSource)
	require.Equal(t, "admin", attrRepo.input.BoundBy)
	require.Equal(t, "manual", out.BoundSource)
	require.Equal(t, "admin", out.BoundBy)
}

func TestDistributionAdminServiceListAttributionAudits(t *testing.T) {
	attrRepo := &distributionAdminAttributionRepoStub{}
	svc := NewDistributionAdminService(nil, nil, attrRepo, nil, nil, nil, nil)
	params := pagination.PaginationParams{Page: 1, PageSize: 20}
	filter := DistributionAdminListFilter{ChannelOrgID: 88, UserID: 7}

	items, page, err := svc.ListAttributionAudits(context.Background(), filter, params)
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, filter, attrRepo.filter)
	require.Equal(t, int64(1), page.Total)
}

func TestDistributionAdminService_ListWalletTransactionsAndRechargeWallet(t *testing.T) {
	walletRepo := &distributionAdminWalletRepoStub{
		wallet: &DistributionWallet{ChannelOrgID: 88, PrepaidBalance: 20},
	}
	txRepo := &distributionAdminWalletTransactionRepoStub{}
	svc := NewDistributionAdminService(nil, nil, nil, nil, walletRepo, nil, nil)
	svc.SetWalletTransactionRepository(txRepo)

	items, page, err := svc.ListWalletTransactions(context.Background(), DistributionAdminListFilter{ChannelOrgID: 88, TransactionType: "recharge"}, pagination.PaginationParams{Page: 1, PageSize: 20})
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, "recharge", txRepo.filter.TransactionType)
	require.Equal(t, int64(1), page.Total)

	out, err := svc.RechargeWallet(context.Background(), 88, DistributionWalletRechargeInput{Amount: 30, ReferenceNo: "BANK-1", Note: "confirmed"})
	require.NoError(t, err)
	require.NotNil(t, out)
	require.NotNil(t, walletRepo.rechargeInput)
	require.InDelta(t, 30, walletRepo.rechargeInput.Amount, 0.0001)
	require.Equal(t, "BANK-1", walletRepo.rechargeInput.ReferenceNo)
	require.InDelta(t, 50, out.PrepaidBalance, 0.0001)
}

func TestDistributionAdminService_RefundWalletCalculatesFeeAndDeductsGrossAmount(t *testing.T) {
	orgRepo := &distributionAdminOrgRepoStub{
		item: &DistributionOrganization{
			ID:   88,
			Type: "reseller",
			Config: map[string]any{
				"refund_fee_rate": 0.1,
			},
		},
	}
	walletRepo := &distributionAdminWalletRepoStub{
		wallet: &DistributionWallet{ChannelOrgID: 88, PrepaidBalance: 200},
	}
	svc := NewDistributionAdminService(orgRepo, nil, nil, nil, walletRepo, nil, nil)

	operatorUserID := int64(9)
	out, err := svc.RefundWallet(context.Background(), 88, DistributionWalletRefundInput{
		Amount:         100,
		ReferenceNo:    "RF-1",
		Note:           "manual refund",
		OperatorUserID: &operatorUserID,
	})
	require.NoError(t, err)
	require.NotNil(t, out)
	require.NotNil(t, out.Wallet)
	require.NotNil(t, walletRepo.refundInput)
	require.InDelta(t, 100, walletRepo.refundInput.Amount, 0.0001)
	require.InDelta(t, 0.1, out.FeeRate, 0.0001)
	require.InDelta(t, 10, out.FeeAmount, 0.0001)
	require.InDelta(t, 90, out.NetAmount, 0.0001)
	require.True(t, out.ProcessedMock)
	require.Contains(t, walletRepo.refundInput.Note, "mock_refund")
	require.InDelta(t, 100, out.RefundAmount, 0.0001)
	require.InDelta(t, 100, out.Wallet.PrepaidBalance, 0.0001)
}

func TestDistributionAdminService_SettleCommissionConsumesReservedWalletBalance(t *testing.T) {
	walletRepo := &distributionAdminWalletRepoStub{
		wallet: &DistributionWallet{ChannelOrgID: 88, OrganizationType: "reseller"},
	}
	settlementRepo := &distributionAdminSettlementRepoStub{
		ledgerByID: &DistributionCommissionLedger{
			ID:               1001,
			ChannelOrgID:     88,
			Amount:           12.5,
			Status:           "available",
			SettlementMethod: "manual",
		},
		settleResult: &DistributionCommissionLedger{
			ID:               1001,
			ChannelOrgID:     88,
			Amount:           12.5,
			Status:           "settled",
			SettlementMethod: "manual",
		},
	}
	svc := NewDistributionAdminService(nil, nil, nil, nil, walletRepo, nil, settlementRepo)

	out, err := svc.SettleCommission(context.Background(), 1001, DistributionCommissionSettlementInput{SettlementMethod: "manual"})
	require.NoError(t, err)
	require.NotNil(t, out)
	require.Equal(t, int64(1001), settlementRepo.settleID)
	require.Equal(t, "manual", settlementRepo.settleInput.SettlementMethod)
	require.Len(t, walletRepo.settleReservedCalls, 1)
	require.InDelta(t, 12.5, walletRepo.settleReservedCalls[0].amount, 0.0001)
	require.Empty(t, walletRepo.releaseCalls)
}

func TestDistributionAdminService_SettleCommissionReleasesReservedWalletBalanceForOfflineSettlement(t *testing.T) {
	walletRepo := &distributionAdminWalletRepoStub{
		wallet: &DistributionWallet{ChannelOrgID: 88, OrganizationType: "reseller"},
	}
	settlementRepo := &distributionAdminSettlementRepoStub{
		ledgerByID: &DistributionCommissionLedger{
			ID:               1001,
			ChannelOrgID:     88,
			Amount:           12.5,
			Status:           "available",
			SettlementMethod: "manual",
		},
		settleResult: &DistributionCommissionLedger{
			ID:               1001,
			ChannelOrgID:     88,
			Amount:           12.5,
			Status:           "settled",
			SettlementMethod: "offline",
		},
	}
	svc := NewDistributionAdminService(nil, nil, nil, nil, walletRepo, nil, settlementRepo)

	out, err := svc.SettleCommission(context.Background(), 1001, DistributionCommissionSettlementInput{SettlementMethod: "offline"})
	require.NoError(t, err)
	require.NotNil(t, out)
	require.Len(t, walletRepo.releaseCalls, 1)
	require.InDelta(t, 12.5, walletRepo.releaseCalls[0].amount, 0.0001)
	require.Empty(t, walletRepo.settleReservedCalls)
}

func TestDistributionAdminService_SettleCommissionCreditsUserBalanceForBalanceSettlement(t *testing.T) {
	walletRepo := &distributionAdminWalletRepoStub{
		wallet: &DistributionWallet{ChannelOrgID: 88, OrganizationType: "platform"},
	}
	settlementRepo := &distributionAdminSettlementRepoStub{
		ledgerByID: &DistributionCommissionLedger{
			ID:               1001,
			ChannelOrgID:     88,
			UserID:           77,
			Amount:           12.5,
			Status:           "available",
			SettlementMethod: "balance",
		},
		settleToBalanceResult: &DistributionCommissionLedger{
			ID:               1001,
			ChannelOrgID:     88,
			UserID:           77,
			Amount:           12.5,
			Status:           "settled",
			SettlementMethod: "balance",
		},
	}
	svc := NewDistributionAdminService(nil, nil, nil, nil, walletRepo, nil, settlementRepo)

	out, err := svc.SettleCommission(context.Background(), 1001, DistributionCommissionSettlementInput{SettlementMethod: "balance"})
	require.NoError(t, err)
	require.NotNil(t, out)
	require.Equal(t, int64(1001), settlementRepo.settleToBalanceID)
	require.Equal(t, "balance", settlementRepo.settleToBalanceInput.SettlementMethod)
	require.Zero(t, settlementRepo.settleID)
}

func TestDistributionAdminService_ReverseCommissionReleasesReservedWalletBalance(t *testing.T) {
	walletRepo := &distributionAdminWalletRepoStub{
		wallet: &DistributionWallet{ChannelOrgID: 88, OrganizationType: "reseller"},
	}
	settlementRepo := &distributionAdminSettlementRepoStub{
		ledgerByID: &DistributionCommissionLedger{
			ID:               1001,
			ChannelOrgID:     88,
			Amount:           12.5,
			Status:           "available",
			SettlementMethod: "manual",
		},
		reverseResult: &DistributionCommissionLedger{
			ID:               1001,
			ChannelOrgID:     88,
			Amount:           12.5,
			Status:           "cancelled",
			SettlementMethod: "manual",
		},
	}
	svc := NewDistributionAdminService(nil, nil, nil, nil, walletRepo, nil, settlementRepo)

	out, err := svc.ReverseCommission(context.Background(), 1001)
	require.NoError(t, err)
	require.NotNil(t, out)
	require.Len(t, walletRepo.releaseCalls, 1)
	require.InDelta(t, 12.5, walletRepo.releaseCalls[0].amount, 0.0001)
	require.Empty(t, walletRepo.refundCalls)
}

func TestDistributionAdminService_ReverseCommissionRefundsSettledWalletBalance(t *testing.T) {
	walletRepo := &distributionAdminWalletRepoStub{
		wallet: &DistributionWallet{ChannelOrgID: 88, OrganizationType: "reseller"},
	}
	settlementRepo := &distributionAdminSettlementRepoStub{
		ledgerByID: &DistributionCommissionLedger{
			ID:               1001,
			ChannelOrgID:     88,
			Amount:           12.5,
			Status:           "settled",
			SettlementMethod: "manual",
		},
		reverseResult: &DistributionCommissionLedger{
			ID:               1002,
			ChannelOrgID:     88,
			Amount:           -12.5,
			Status:           "reversed",
			SettlementMethod: "manual",
		},
	}
	svc := NewDistributionAdminService(nil, nil, nil, nil, walletRepo, nil, settlementRepo)

	out, err := svc.ReverseCommission(context.Background(), 1001)
	require.NoError(t, err)
	require.NotNil(t, out)
	require.Len(t, walletRepo.refundCalls, 1)
	require.InDelta(t, 12.5, walletRepo.refundCalls[0].amount, 0.0001)
	require.Empty(t, walletRepo.releaseCalls)
}

func TestDistributionAdminService_ReverseCommissionDeductsUserBalanceForBalanceSettlement(t *testing.T) {
	walletRepo := &distributionAdminWalletRepoStub{
		wallet: &DistributionWallet{ChannelOrgID: 88, OrganizationType: "platform"},
	}
	settlementRepo := &distributionAdminSettlementRepoStub{
		ledgerByID: &DistributionCommissionLedger{
			ID:               1001,
			ChannelOrgID:     88,
			UserID:           77,
			Amount:           12.5,
			Status:           "settled",
			SettlementMethod: "balance",
		},
		reverseBalanceResult: &DistributionCommissionLedger{
			ID:               1002,
			ChannelOrgID:     88,
			UserID:           77,
			Amount:           -12.5,
			Status:           "reversed",
			SettlementMethod: "balance",
		},
	}
	svc := NewDistributionAdminService(nil, nil, nil, nil, walletRepo, nil, settlementRepo)

	out, err := svc.ReverseCommission(context.Background(), 1001)
	require.NoError(t, err)
	require.NotNil(t, out)
	require.Equal(t, int64(1001), settlementRepo.reverseBalanceID)
	require.Zero(t, settlementRepo.reverseID)
}
