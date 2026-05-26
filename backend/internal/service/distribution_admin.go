package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

type DistributionAdminListFilter struct {
	ChannelOrgID    int64
	UserID          int64
	RoleType        string
	TransactionType string
	Q               string
}

func (f DistributionAdminListFilter) normalized() DistributionAdminListFilter {
	f.RoleType = strings.ToLower(strings.TrimSpace(f.RoleType))
	f.TransactionType = normalizeDistributionWalletTransactionType(f.TransactionType)
	f.Q = strings.TrimSpace(f.Q)
	if f.ChannelOrgID < 0 {
		f.ChannelOrgID = 0
	}
	if f.UserID < 0 {
		f.UserID = 0
	}
	return f
}

type DistributionOrganizationListRepository interface {
	List(ctx context.Context, params pagination.PaginationParams) ([]DistributionOrganization, *pagination.PaginationResult, error)
	GetByID(ctx context.Context, id int64) (*DistributionOrganization, error)
}

type DistributionMemberAdminListRepository interface {
	ListAdmin(ctx context.Context, filter DistributionAdminListFilter, params pagination.PaginationParams) ([]DistributionMemberView, *pagination.PaginationResult, error)
}

type DistributionAttributionAdminListRepository interface {
	ListAdmin(ctx context.Context, filter DistributionAdminListFilter, params pagination.PaginationParams) ([]DistributionAttributionView, *pagination.PaginationResult, error)
	ListAuditsAdmin(ctx context.Context, filter DistributionAdminListFilter, params pagination.PaginationParams) ([]DistributionAttributionAuditView, *pagination.PaginationResult, error)
	GetByUserID(ctx context.Context, userID int64) (*DistributionAttribution, error)
	UpdateByAdmin(ctx context.Context, input DistributionAttributionAdminUpdateInput) (*DistributionAttribution, error)
}

type DistributionCommissionAdminListRepository interface {
	ListAdmin(ctx context.Context, filter DistributionAdminListFilter, params pagination.PaginationParams) ([]DistributionCommissionLedgerView, *pagination.PaginationResult, error)
}

type DistributionWalletAdminListRepository interface {
	List(ctx context.Context, filter DistributionAdminListFilter, params pagination.PaginationParams) ([]DistributionWallet, *pagination.PaginationResult, error)
	GetByChannelOrgID(ctx context.Context, channelOrgID int64) (*DistributionWallet, error)
	UpdateWarningThreshold(ctx context.Context, channelOrgID int64, warningThreshold float64) (*DistributionWallet, error)
	Recharge(ctx context.Context, channelOrgID int64, input DistributionWalletRechargeInput) (*DistributionWallet, error)
	RefundPrepaidBalance(ctx context.Context, channelOrgID int64, input DistributionWalletRefundInput) (*DistributionWallet, error)
	ReserveCommission(ctx context.Context, channelOrgID int64, amount float64) (*DistributionWallet, error)
	ReleaseCommission(ctx context.Context, channelOrgID int64, amount float64) (*DistributionWallet, error)
	SettleReservedCommission(ctx context.Context, channelOrgID int64, amount float64) (*DistributionWallet, error)
	DeductCommission(ctx context.Context, channelOrgID int64, amount float64) (*DistributionWallet, error)
	RefundCommission(ctx context.Context, channelOrgID int64, amount float64) (*DistributionWallet, error)
}

type DistributionStatsAdminRepository interface {
	GetAdminStats(ctx context.Context) (*DistributionAdminStats, error)
}

type DistributionCommissionSettlementRepository interface {
	GetByID(ctx context.Context, commissionID int64) (*DistributionCommissionLedger, error)
	Settle(ctx context.Context, commissionID int64, input DistributionCommissionSettlementInput) (*DistributionCommissionLedger, error)
	SettleToBalance(ctx context.Context, commissionID int64, input DistributionCommissionSettlementInput) (*DistributionCommissionLedger, error)
	Reverse(ctx context.Context, commissionID int64) (*DistributionCommissionLedger, error)
	ReverseBalanceSettlement(ctx context.Context, commissionID int64) (*DistributionCommissionLedger, error)
}

type DistributionCommissionSettlementInput struct {
	SettlementMethod      string
	SettlementReferenceNo string
	SettlementNote        string
	SettledByUserID       *int64
}

type DistributionAdminService struct {
	organizationRepo      DistributionOrganizationListRepository
	memberRepo            DistributionMemberAdminListRepository
	attributionRepo       DistributionAttributionAdminListRepository
	commissionRepo        DistributionCommissionAdminListRepository
	walletRepo            DistributionWalletAdminListRepository
	walletRequestRepo     DistributionWalletRequestRepository
	alertEventRepo        DistributionAlertEventRepository
	walletTransactionRepo DistributionWalletTransactionListRepository
	statsRepo             DistributionStatsAdminRepository
	settlementRepo        DistributionCommissionSettlementRepository
}

func NewDistributionAdminService(
	organizationRepo DistributionOrganizationListRepository,
	memberRepo DistributionMemberAdminListRepository,
	attributionRepo DistributionAttributionAdminListRepository,
	commissionRepo DistributionCommissionAdminListRepository,
	walletRepo DistributionWalletAdminListRepository,
	statsRepo DistributionStatsAdminRepository,
	settlementRepo DistributionCommissionSettlementRepository,
) *DistributionAdminService {
	return &DistributionAdminService{
		organizationRepo: organizationRepo,
		memberRepo:       memberRepo,
		attributionRepo:  attributionRepo,
		commissionRepo:   commissionRepo,
		walletRepo:       walletRepo,
		statsRepo:        statsRepo,
		settlementRepo:   settlementRepo,
	}
}

func (s *DistributionAdminService) SetWalletTransactionRepository(repo DistributionWalletTransactionListRepository) {
	if s == nil {
		return
	}
	s.walletTransactionRepo = repo
}

func (s *DistributionAdminService) SetWalletRequestRepository(repo DistributionWalletRequestRepository) {
	if s == nil {
		return
	}
	s.walletRequestRepo = repo
}

func (s *DistributionAdminService) SetAlertEventRepository(repo DistributionAlertEventRepository) {
	if s == nil {
		return
	}
	s.alertEventRepo = repo
}

func (s *DistributionAdminService) ListOrganizations(ctx context.Context, params pagination.PaginationParams) ([]DistributionOrganization, *pagination.PaginationResult, error) {
	if s == nil || s.organizationRepo == nil {
		return nil, nil, ErrInvalidDistributionOrganization
	}
	return s.organizationRepo.List(ctx, params)
}

func (s *DistributionAdminService) ListMembers(ctx context.Context, filter DistributionAdminListFilter, params pagination.PaginationParams) ([]DistributionMemberView, *pagination.PaginationResult, error) {
	if s == nil || s.memberRepo == nil {
		return nil, nil, ErrInvalidDistributionMember
	}
	return s.memberRepo.ListAdmin(ctx, filter.normalized(), params)
}

func (s *DistributionAdminService) ListAttributions(ctx context.Context, filter DistributionAdminListFilter, params pagination.PaginationParams) ([]DistributionAttributionView, *pagination.PaginationResult, error) {
	if s == nil || s.attributionRepo == nil {
		return nil, nil, ErrInvalidDistributionAttribution
	}
	return s.attributionRepo.ListAdmin(ctx, filter.normalized(), params)
}

func (s *DistributionAdminService) ListAttributionAudits(ctx context.Context, filter DistributionAdminListFilter, params pagination.PaginationParams) ([]DistributionAttributionAuditView, *pagination.PaginationResult, error) {
	if s == nil || s.attributionRepo == nil {
		return nil, nil, ErrInvalidDistributionAttribution
	}
	return s.attributionRepo.ListAuditsAdmin(ctx, filter.normalized(), params)
}

func (s *DistributionAdminService) UpdateAttribution(ctx context.Context, userID int64, input DistributionAttributionAdminUpdateInput) (*DistributionAttribution, error) {
	if s == nil || s.attributionRepo == nil {
		return nil, ErrInvalidDistributionAttribution
	}
	if userID <= 0 || input.ChannelOrgID <= 0 {
		return nil, ErrInvalidDistributionAttribution
	}

	input.UserID = userID
	input.BoundSource = "manual"
	input.BoundBy = "admin"

	return s.attributionRepo.UpdateByAdmin(ctx, input)
}

func (s *DistributionAdminService) ListCommissions(ctx context.Context, filter DistributionAdminListFilter, params pagination.PaginationParams) ([]DistributionCommissionLedgerView, *pagination.PaginationResult, error) {
	if s == nil || s.commissionRepo == nil {
		return nil, nil, ErrInvalidDistributionCommission
	}
	return s.commissionRepo.ListAdmin(ctx, filter.normalized(), params)
}

func (s *DistributionAdminService) ListWallets(ctx context.Context, filter DistributionAdminListFilter, params pagination.PaginationParams) ([]DistributionWallet, *pagination.PaginationResult, error) {
	if s == nil || s.walletRepo == nil {
		return nil, nil, ErrInvalidDistributionWallet
	}
	return s.walletRepo.List(ctx, filter.normalized(), params)
}

func (s *DistributionAdminService) GetStats(ctx context.Context) (*DistributionAdminStats, error) {
	if s == nil || s.statsRepo == nil {
		return nil, ErrInvalidDistributionStats
	}
	return s.statsRepo.GetAdminStats(ctx)
}

func (s *DistributionAdminService) ListWalletTransactions(ctx context.Context, filter DistributionAdminListFilter, params pagination.PaginationParams) ([]DistributionWalletTransaction, *pagination.PaginationResult, error) {
	if s == nil || s.walletTransactionRepo == nil {
		return nil, nil, ErrInvalidDistributionWalletTransaction
	}
	return s.walletTransactionRepo.ListTransactions(ctx, filter.normalized(), params)
}

func (s *DistributionAdminService) ListWalletRequests(ctx context.Context, filter DistributionWalletRequestListFilter, params pagination.PaginationParams) ([]DistributionWalletRequest, *pagination.PaginationResult, error) {
	if s == nil || s.walletRequestRepo == nil {
		return nil, nil, ErrInvalidDistributionWalletRequest
	}
	return s.walletRequestRepo.ListRequests(ctx, filter.normalized(), params)
}

func (s *DistributionAdminService) ListAlertEvents(ctx context.Context, filter DistributionAlertEventListFilter, params pagination.PaginationParams) ([]DistributionAlertEvent, *pagination.PaginationResult, error) {
	if s == nil || s.alertEventRepo == nil {
		return nil, nil, ErrInvalidDistributionAlertEvent
	}
	return s.alertEventRepo.ListAlertEvents(ctx, filter.normalized(), params)
}

func (s *DistributionAdminService) CreateWalletRequest(ctx context.Context, input DistributionWalletRequestCreateInput) (*DistributionWalletRequest, error) {
	if s == nil || s.walletRequestRepo == nil {
		return nil, ErrInvalidDistributionWalletRequest
	}
	if input.ChannelOrgID <= 0 || input.CreatedByUserID <= 0 || input.Amount <= 0 {
		return nil, ErrInvalidDistributionWalletRequest
	}
	input.RequestType = normalizeDistributionWalletRequestType(input.RequestType)
	if input.RequestType == "" {
		return nil, ErrInvalidDistributionWalletRequest
	}
	return s.walletRequestRepo.CreateRequest(ctx, input)
}

func (s *DistributionAdminService) ReviewWalletRequest(ctx context.Context, requestID int64, input DistributionWalletRequestReviewInput) (*DistributionWalletRequest, error) {
	if s == nil || s.walletRequestRepo == nil {
		return nil, ErrInvalidDistributionWalletRequest
	}
	if requestID <= 0 || input.ReviewedByUserID <= 0 {
		return nil, ErrInvalidDistributionWalletRequest
	}
	input.Action = normalizeDistributionWalletRequestAction(input.Action)
	if input.Action == "" {
		return nil, ErrInvalidDistributionWalletRequest
	}

	request, err := s.walletRequestRepo.GetRequestByID(ctx, requestID)
	if err != nil {
		return nil, err
	}
	if request == nil || request.ID <= 0 || request.ChannelOrgID <= 0 || request.Amount <= 0 {
		return nil, ErrInvalidDistributionWalletRequest
	}
	if !strings.EqualFold(strings.TrimSpace(request.Status), "pending") {
		return nil, ErrInvalidDistributionWalletRequest
	}

	if input.Action == "reject" {
		return s.walletRequestRepo.RejectRequest(ctx, requestID, input)
	}

	switch normalizeDistributionWalletRequestType(request.RequestType) {
	case "recharge":
		note := strings.TrimSpace(request.Note)
		if note != "" {
			note += " "
		}
		note += fmt.Sprintf("request_id=%d", requestID)
		return s.walletRequestRepo.ApproveRechargeRequest(ctx, requestID, input, DistributionWalletRechargeInput{
			Amount:         request.Amount,
			ReferenceNo:    request.ReferenceNo,
			Note:           note,
			OperatorUserID: &input.ReviewedByUserID,
		})
	case "refund":
		if s.organizationRepo == nil {
			return nil, ErrInvalidDistributionWalletRequest
		}
		org, err := s.organizationRepo.GetByID(ctx, request.ChannelOrgID)
		if err != nil {
			return nil, err
		}
		feeRate := normalizeDistributionRefundFeeRate(distributionOrganizationConfigFloat(nilSafeDistributionOrgConfig(org), "refund_fee_rate"))
		feeAmount := roundDistributionWalletAmount(request.Amount * feeRate)
		netAmount := roundDistributionWalletAmount(request.Amount - feeAmount)
		note := strings.TrimSpace(request.Note)
		if note != "" {
			note += " "
		}
		note += fmt.Sprintf("request_id=%d", requestID)
		refundInput := DistributionWalletRefundInput{
			Amount:         request.Amount,
			ReferenceNo:    request.ReferenceNo,
			Note:           buildDistributionWalletRefundNote(note, feeRate, feeAmount, netAmount),
			OperatorUserID: &input.ReviewedByUserID,
		}
		return s.walletRequestRepo.ApproveRefundRequest(ctx, requestID, input, refundInput)
	default:
		return nil, ErrInvalidDistributionWalletRequest
	}
}

func (s *DistributionAdminService) GetCommission(ctx context.Context, commissionID int64) (*DistributionCommissionLedger, error) {
	if s == nil || s.settlementRepo == nil {
		return nil, ErrInvalidDistributionCommission
	}
	return s.settlementRepo.GetByID(ctx, commissionID)
}

func (s *DistributionAdminService) RechargeWallet(ctx context.Context, channelOrgID int64, input DistributionWalletRechargeInput) (*DistributionWallet, error) {
	if s == nil || s.walletRepo == nil {
		return nil, ErrInvalidDistributionWallet
	}
	return s.walletRepo.Recharge(ctx, channelOrgID, input)
}

func (s *DistributionAdminService) RefundWallet(ctx context.Context, channelOrgID int64, input DistributionWalletRefundInput) (*DistributionWalletRefundResult, error) {
	if s == nil || s.walletRepo == nil || s.organizationRepo == nil {
		return nil, ErrInvalidDistributionWallet
	}
	if channelOrgID <= 0 || input.Amount <= 0 {
		return nil, ErrInvalidDistributionWallet
	}

	org, err := s.organizationRepo.GetByID(ctx, channelOrgID)
	if err != nil {
		return nil, err
	}
	feeRate := normalizeDistributionRefundFeeRate(distributionOrganizationConfigFloat(nilSafeDistributionOrgConfig(org), "refund_fee_rate"))
	feeAmount := roundDistributionWalletAmount(input.Amount * feeRate)
	netAmount := roundDistributionWalletAmount(input.Amount - feeAmount)

	input.Note = buildDistributionWalletRefundNote(input.Note, feeRate, feeAmount, netAmount)
	wallet, err := s.walletRepo.RefundPrepaidBalance(ctx, channelOrgID, input)
	if err != nil {
		return nil, err
	}

	return &DistributionWalletRefundResult{
		Wallet:        wallet,
		RefundAmount:  roundDistributionWalletAmount(input.Amount),
		FeeRate:       feeRate,
		FeeAmount:     feeAmount,
		NetAmount:     netAmount,
		ReferenceNo:   input.ReferenceNo,
		Note:          input.Note,
		ProcessedMock: true,
	}, nil
}

func (s *DistributionAdminService) UpdateWalletWarningThreshold(ctx context.Context, channelOrgID int64, warningThreshold float64) (*DistributionWallet, error) {
	if s == nil || s.walletRepo == nil {
		return nil, ErrInvalidDistributionWallet
	}
	return s.walletRepo.UpdateWarningThreshold(ctx, channelOrgID, warningThreshold)
}

func (s *DistributionAdminService) SettleCommission(ctx context.Context, commissionID int64, input DistributionCommissionSettlementInput) (*DistributionCommissionLedger, error) {
	if s == nil || s.settlementRepo == nil {
		return nil, ErrInvalidDistributionCommission
	}
	original, err := s.settlementRepo.GetByID(ctx, commissionID)
	if err != nil {
		return nil, err
	}
	if original == nil {
		return nil, ErrInvalidDistributionCommission
	}
	if strings.EqualFold(strings.TrimSpace(original.Status), "settled") {
		return original, nil
	}
	method := normalizeDistributionSettlementMethod(input.SettlementMethod, original.SettlementMethod)
	if err := s.applyWalletOnSettle(ctx, original, method); err != nil {
		return nil, err
	}
	input.SettlementMethod = method
	if requiresDistributionBalancePayout(method) {
		return s.settlementRepo.SettleToBalance(ctx, commissionID, input)
	}
	return s.settlementRepo.Settle(ctx, commissionID, input)
}

func (s *DistributionAdminService) ReverseCommission(ctx context.Context, commissionID int64) (*DistributionCommissionLedger, error) {
	if s == nil || s.settlementRepo == nil {
		return nil, ErrInvalidDistributionCommission
	}
	original, err := s.settlementRepo.GetByID(ctx, commissionID)
	if err != nil {
		return nil, err
	}
	if original == nil {
		return nil, ErrInvalidDistributionCommission
	}
	if err := s.applyWalletOnReverse(ctx, original); err != nil {
		return nil, err
	}
	if strings.EqualFold(strings.TrimSpace(original.Status), "settled") && requiresDistributionBalancePayout(original.SettlementMethod) {
		return s.settlementRepo.ReverseBalanceSettlement(ctx, commissionID)
	}
	return s.settlementRepo.Reverse(ctx, commissionID)
}

func (s *DistributionAdminService) applyWalletOnSettle(ctx context.Context, original *DistributionCommissionLedger, settlementMethod string) error {
	if !s.shouldAdjustWallet(ctx, original) || original == nil || original.Amount <= 0 {
		return nil
	}
	wasReserved := requiresDistributionWalletReserve(original.SettlementMethod)
	needsDeduction := requiresDistributionWalletReserve(settlementMethod)

	switch {
	case wasReserved && needsDeduction:
		_, err := s.walletRepo.SettleReservedCommission(ctx, original.ChannelOrgID, original.Amount)
		return err
	case wasReserved && !needsDeduction:
		_, err := s.walletRepo.ReleaseCommission(ctx, original.ChannelOrgID, original.Amount)
		return err
	case !wasReserved && needsDeduction:
		_, err := s.walletRepo.DeductCommission(ctx, original.ChannelOrgID, original.Amount)
		return err
	default:
		return nil
	}
}

func (s *DistributionAdminService) applyWalletOnReverse(ctx context.Context, original *DistributionCommissionLedger) error {
	if !s.shouldAdjustWallet(ctx, original) || original == nil || original.Amount <= 0 {
		return nil
	}
	switch {
	case strings.EqualFold(strings.TrimSpace(original.Status), "settled") && requiresDistributionWalletReserve(original.SettlementMethod):
		_, err := s.walletRepo.RefundCommission(ctx, original.ChannelOrgID, original.Amount)
		return err
	case requiresDistributionWalletReserve(original.SettlementMethod):
		_, err := s.walletRepo.ReleaseCommission(ctx, original.ChannelOrgID, original.Amount)
		return err
	default:
		return nil
	}
}

func (s *DistributionAdminService) shouldAdjustWallet(ctx context.Context, ledger *DistributionCommissionLedger) bool {
	if s == nil || s.walletRepo == nil || ledger == nil || ledger.ChannelOrgID <= 0 {
		return false
	}
	wallet, err := s.walletRepo.GetByChannelOrgID(ctx, ledger.ChannelOrgID)
	if err != nil || wallet == nil {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(wallet.OrganizationType)) {
	case "reseller", "oem":
		return true
	default:
		return false
	}
}

func normalizeDistributionSettlementMethod(candidate string, fallback string) string {
	for _, value := range []string{candidate, fallback} {
		switch strings.ToLower(strings.TrimSpace(value)) {
		case "balance", "auto", "manual", "offline":
			return strings.ToLower(strings.TrimSpace(value))
		}
	}
	return "manual"
}

func requiresDistributionBalancePayout(settlementMethod string) bool {
	switch strings.ToLower(strings.TrimSpace(settlementMethod)) {
	case "balance", "auto":
		return true
	default:
		return false
	}
}

func normalizeDistributionRefundFeeRate(rate float64) float64 {
	if rate > 1 && rate <= 100 {
		rate = rate / 100
	}
	if rate < 0 {
		return 0
	}
	if rate > 1 {
		return 1
	}
	return rate
}

func buildDistributionWalletRefundNote(note string, feeRate float64, feeAmount float64, netAmount float64) string {
	base := strings.TrimSpace(note)
	meta := fmt.Sprintf("[mock_refund fee_rate=%.4f fee_amount=%.8f net_amount=%.8f]", feeRate, feeAmount, netAmount)
	if base == "" {
		return meta
	}
	return base + " " + meta
}
