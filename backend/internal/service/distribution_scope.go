package service

import (
	"context"
	"sort"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

// DistributionAttributionView is a user-facing attribution row with the linked
// user identity for scoped distribution pages.
type DistributionAttributionView struct {
	UserID           int64     `json:"user_id"`
	UserEmail        string    `json:"user_email"`
	Username         string    `json:"username"`
	ChannelOrgID     int64     `json:"channel_org_id"`
	ReferrerMemberID *int64    `json:"referrer_member_id,omitempty"`
	PromotionLinkID  *int64    `json:"promotion_link_id,omitempty"`
	BoundAt          time.Time `json:"bound_at"`
	BoundSource      string    `json:"bound_source"`
	BoundBy          string    `json:"bound_by"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// DistributionMemberView is a user-facing channel member row.
type DistributionMemberView struct {
	MemberID       int64     `json:"member_id"`
	UserID         int64     `json:"user_id"`
	UserEmail      string    `json:"user_email"`
	Username       string    `json:"username"`
	ChannelOrgID   int64     `json:"channel_org_id"`
	RoleType       string    `json:"role_type"`
	ParentMemberID *int64    `json:"parent_member_id,omitempty"`
	LevelCode      string    `json:"level_code"`
	CommissionRate float64   `json:"commission_rate"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// DistributionCommissionLedgerView is a user-facing commission ledger row.
type DistributionCommissionLedgerView struct {
	ID                    int64      `json:"id"`
	ChannelOrgID          int64      `json:"channel_org_id"`
	MemberID              int64      `json:"member_id"`
	UserID                int64      `json:"user_id"`
	UserEmail             string     `json:"user_email"`
	Username              string     `json:"username"`
	UsageLogID            *int64     `json:"usage_log_id,omitempty"`
	CommissionType        string     `json:"commission_type"`
	BaseAmount            float64    `json:"base_amount"`
	Rate                  float64    `json:"rate"`
	Amount                float64    `json:"amount"`
	Status                string     `json:"status"`
	SettlementMethod      string     `json:"settlement_method"`
	SettlementReferenceNo string     `json:"settlement_reference_no"`
	SettlementNote        string     `json:"settlement_note"`
	FrozenUntil           *time.Time `json:"frozen_until,omitempty"`
	SettledAt             *time.Time `json:"settled_at,omitempty"`
	SettledByUserID       *int64     `json:"settled_by_user_id,omitempty"`
	ReversedFromID        *int64     `json:"reversed_from_id,omitempty"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

type DistributionAttributionListRepository interface {
	GetByUserID(ctx context.Context, userID int64) (*DistributionAttribution, error)
	ListByChannelOrgID(ctx context.Context, channelOrgID int64, params pagination.PaginationParams) ([]DistributionAttributionView, *pagination.PaginationResult, error)
}

type DistributionMemberListRepository interface {
	ListByChannelOrgID(ctx context.Context, channelOrgID int64, params pagination.PaginationParams, roleType string) ([]DistributionMemberView, *pagination.PaginationResult, error)
	ListByUserID(ctx context.Context, userID int64) ([]DistributionMemberView, error)
}

type DistributionCommissionListRepository interface {
	ListByChannelOrgID(ctx context.Context, channelOrgID int64, params pagination.PaginationParams) ([]DistributionCommissionLedgerView, *pagination.PaginationResult, error)
}

type DistributionWalletSummaryRepository interface {
	GetByChannelOrgID(ctx context.Context, channelOrgID int64) (*DistributionWallet, error)
}

type DistributionStatsSummaryRepository interface {
	GetChannelSummary(ctx context.Context, channelOrgID int64) (*DistributionChannelSummary, error)
}

type DistributionScopeService struct {
	attributionRepo       DistributionAttributionListRepository
	memberRepo            DistributionMemberListRepository
	commissionRepo        DistributionCommissionListRepository
	walletRepo            DistributionWalletSummaryRepository
	walletTransactionRepo DistributionWalletTransactionListRepository
	statsRepo             DistributionStatsSummaryRepository
	organizationRepo      distributionUserChannelOrganizationRepository
}

func NewDistributionScopeService(
	attributionRepo DistributionAttributionListRepository,
	memberRepo DistributionMemberListRepository,
	commissionRepo DistributionCommissionListRepository,
	walletRepo DistributionWalletSummaryRepository,
	statsRepo DistributionStatsSummaryRepository,
) *DistributionScopeService {
	return &DistributionScopeService{
		attributionRepo: attributionRepo,
		memberRepo:      memberRepo,
		commissionRepo:  commissionRepo,
		walletRepo:      walletRepo,
		statsRepo:       statsRepo,
	}
}

func (s *DistributionScopeService) SetOrganizationRepository(repo distributionUserChannelOrganizationRepository) {
	if s == nil {
		return
	}
	s.organizationRepo = repo
}

func (s *DistributionScopeService) SetWalletTransactionRepository(repo DistributionWalletTransactionListRepository) {
	if s == nil {
		return
	}
	s.walletTransactionRepo = repo
}

func (s *DistributionScopeService) ResolveUserChannelOrgID(ctx context.Context, userID int64) (int64, error) {
	if s == nil || userID <= 0 {
		return 0, ErrInvalidDistributionAttribution
	}
	return resolveDistributionUserChannelOrgID(ctx, userID, s.memberRepo, s.organizationRepo, s.attributionRepo)
}

func (s *DistributionScopeService) ListMembersForUser(
	ctx context.Context,
	userID int64,
	params pagination.PaginationParams,
	roleType string,
) ([]DistributionMemberView, *pagination.PaginationResult, error) {
	if s == nil {
		return nil, nil, infraerrors.ServiceUnavailable("DISTRIBUTION_SCOPE_NOT_READY", "distribution scope is not ready")
	}
	channelOrgID, err := s.ResolveUserChannelOrgID(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	if s.memberRepo == nil {
		return nil, nil, ErrInvalidDistributionAttribution
	}
	canManage, err := s.CanManageChannelForUser(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	if canManage {
		return s.memberRepo.ListByChannelOrgID(ctx, channelOrgID, params, roleType)
	}
	allMembers, err := s.listAllMembersByChannel(ctx, channelOrgID)
	if err != nil {
		return nil, nil, err
	}
	visibleIDs := distributionVisibleMemberIDs(channelOrgID, allMembers, userID)
	filtered := make([]DistributionMemberView, 0, len(allMembers))
	roleType = strings.ToLower(strings.TrimSpace(roleType))
	for _, member := range allMembers {
		if _, ok := visibleIDs[member.MemberID]; !ok {
			continue
		}
		if roleType != "" && !strings.EqualFold(strings.TrimSpace(member.RoleType), roleType) {
			continue
		}
		filtered = append(filtered, member)
	}
	return paginateDistributionViews(filtered, params), buildDistributionPaginationResult(int64(len(filtered)), params), nil
}

func (s *DistributionScopeService) ListAttributionsForUser(
	ctx context.Context,
	userID int64,
	params pagination.PaginationParams,
) ([]DistributionAttributionView, *pagination.PaginationResult, error) {
	if s == nil {
		return nil, nil, infraerrors.ServiceUnavailable("DISTRIBUTION_SCOPE_NOT_READY", "distribution scope is not ready")
	}
	channelOrgID, err := s.ResolveUserChannelOrgID(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	if s.attributionRepo == nil {
		return nil, nil, ErrInvalidDistributionAttribution
	}
	canManage, err := s.CanManageChannelForUser(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	if canManage {
		return s.attributionRepo.ListByChannelOrgID(ctx, channelOrgID, params)
	}
	allMembers, err := s.listAllMembersByChannel(ctx, channelOrgID)
	if err != nil {
		return nil, nil, err
	}
	visibleIDs := distributionVisibleMemberIDs(channelOrgID, allMembers, userID)
	allRows, err := s.listAllAttributionsByChannel(ctx, channelOrgID)
	if err != nil {
		return nil, nil, err
	}
	filtered := make([]DistributionAttributionView, 0, len(allRows))
	for _, row := range allRows {
		if row.ReferrerMemberID == nil {
			continue
		}
		if _, ok := visibleIDs[*row.ReferrerMemberID]; !ok {
			continue
		}
		filtered = append(filtered, row)
	}
	return paginateDistributionViews(filtered, params), buildDistributionPaginationResult(int64(len(filtered)), params), nil
}

func (s *DistributionScopeService) ListCommissionsForUser(
	ctx context.Context,
	userID int64,
	params pagination.PaginationParams,
) ([]DistributionCommissionLedgerView, *pagination.PaginationResult, error) {
	if s == nil {
		return nil, nil, infraerrors.ServiceUnavailable("DISTRIBUTION_SCOPE_NOT_READY", "distribution scope is not ready")
	}
	channelOrgID, err := s.ResolveUserChannelOrgID(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	if s.commissionRepo == nil {
		return nil, nil, ErrInvalidDistributionAttribution
	}
	canManage, err := s.CanManageChannelForUser(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	if canManage {
		return s.commissionRepo.ListByChannelOrgID(ctx, channelOrgID, params)
	}
	allMembers, err := s.listAllMembersByChannel(ctx, channelOrgID)
	if err != nil {
		return nil, nil, err
	}
	visibleIDs := distributionVisibleMemberIDs(channelOrgID, allMembers, userID)
	allRows, err := s.listAllCommissionsByChannel(ctx, channelOrgID)
	if err != nil {
		return nil, nil, err
	}
	filtered := make([]DistributionCommissionLedgerView, 0, len(allRows))
	for _, row := range allRows {
		if _, ok := visibleIDs[row.MemberID]; !ok {
			continue
		}
		filtered = append(filtered, row)
	}
	return paginateDistributionViews(filtered, params), buildDistributionPaginationResult(int64(len(filtered)), params), nil
}

func (s *DistributionScopeService) GetOverviewForUser(ctx context.Context, userID int64) (*DistributionChannelSummary, error) {
	if s == nil {
		return nil, infraerrors.ServiceUnavailable("DISTRIBUTION_SCOPE_NOT_READY", "distribution scope is not ready")
	}
	channelOrgID, err := s.ResolveUserChannelOrgID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if s.statsRepo == nil {
		return nil, ErrInvalidDistributionAttribution
	}
	return s.statsRepo.GetChannelSummary(ctx, channelOrgID)
}

func (s *DistributionScopeService) CanManageChannelForUser(ctx context.Context, userID int64) (bool, error) {
	if s == nil || s.memberRepo == nil {
		return false, infraerrors.ServiceUnavailable("DISTRIBUTION_SCOPE_NOT_READY", "distribution scope is not ready")
	}
	channelOrgID, err := s.ResolveUserChannelOrgID(ctx, userID)
	if err != nil {
		return false, err
	}
	if s.organizationRepo != nil {
		org, err := s.organizationRepo.GetByOwnerUserID(ctx, userID)
		if err != nil {
			return false, err
		}
		if org != nil && org.ID == channelOrgID && strings.EqualFold(strings.TrimSpace(org.Status), "active") {
			return true, nil
		}
	}
	members, err := s.memberRepo.ListByUserID(ctx, userID)
	if err != nil {
		return false, err
	}
	for _, member := range members {
		if member.ChannelOrgID == channelOrgID &&
			strings.EqualFold(strings.TrimSpace(member.RoleType), "manager") &&
			strings.EqualFold(strings.TrimSpace(member.Status), "active") {
			return true, nil
		}
	}
	return false, nil
}

func (s *DistributionScopeService) ListWalletTransactionsForUser(
	ctx context.Context,
	userID int64,
	params pagination.PaginationParams,
	transactionType string,
) ([]DistributionWalletTransaction, *pagination.PaginationResult, error) {
	if s == nil {
		return nil, nil, infraerrors.ServiceUnavailable("DISTRIBUTION_SCOPE_NOT_READY", "distribution scope is not ready")
	}
	channelOrgID, err := s.ResolveUserChannelOrgID(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	if s.walletTransactionRepo == nil {
		return nil, nil, ErrInvalidDistributionWalletTransaction
	}
	canManage, err := s.CanManageChannelForUser(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	if !canManage {
		return nil, nil, infraerrors.Forbidden("DISTRIBUTION_CHANNEL_PERMISSION_DENIED", "distribution channel permission denied")
	}
	return s.walletTransactionRepo.ListTransactionsByChannelOrgID(ctx, channelOrgID, params, transactionType)
}

func (s *DistributionScopeService) listAllMembersByChannel(ctx context.Context, channelOrgID int64) ([]DistributionMemberView, error) {
	const pageSize = 500
	page := 1
	out := make([]DistributionMemberView, 0, pageSize)
	for {
		items, meta, err := s.memberRepo.ListByChannelOrgID(ctx, channelOrgID, pagination.PaginationParams{Page: page, PageSize: pageSize}, "")
		if err != nil {
			return nil, err
		}
		out = append(out, items...)
		if meta == nil || page >= meta.Pages || len(items) == 0 {
			break
		}
		page++
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].MemberID < out[j].MemberID
	})
	return out, nil
}

func (s *DistributionScopeService) listAllAttributionsByChannel(ctx context.Context, channelOrgID int64) ([]DistributionAttributionView, error) {
	const pageSize = 500
	page := 1
	out := make([]DistributionAttributionView, 0, pageSize)
	for {
		items, meta, err := s.attributionRepo.ListByChannelOrgID(ctx, channelOrgID, pagination.PaginationParams{Page: page, PageSize: pageSize})
		if err != nil {
			return nil, err
		}
		out = append(out, items...)
		if meta == nil || page >= meta.Pages || len(items) == 0 {
			break
		}
		page++
	}
	return out, nil
}

func (s *DistributionScopeService) listAllCommissionsByChannel(ctx context.Context, channelOrgID int64) ([]DistributionCommissionLedgerView, error) {
	const pageSize = 500
	page := 1
	out := make([]DistributionCommissionLedgerView, 0, pageSize)
	for {
		items, meta, err := s.commissionRepo.ListByChannelOrgID(ctx, channelOrgID, pagination.PaginationParams{Page: page, PageSize: pageSize})
		if err != nil {
			return nil, err
		}
		out = append(out, items...)
		if meta == nil || page >= meta.Pages || len(items) == 0 {
			break
		}
		page++
	}
	return out, nil
}

func distributionVisibleMemberIDs(channelOrgID int64, members []DistributionMemberView, userID int64) map[int64]struct{} {
	children := make(map[int64][]int64, len(members))
	roots := make([]int64, 0, 2)
	for _, member := range members {
		if member.ChannelOrgID != channelOrgID || !strings.EqualFold(strings.TrimSpace(member.Status), "active") {
			continue
		}
		if member.ParentMemberID != nil && *member.ParentMemberID > 0 {
			children[*member.ParentMemberID] = append(children[*member.ParentMemberID], member.MemberID)
		}
		if member.UserID == userID {
			roots = append(roots, member.MemberID)
		}
	}
	visible := make(map[int64]struct{}, len(roots))
	queue := append([]int64(nil), roots...)
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if _, seen := visible[current]; seen {
			continue
		}
		visible[current] = struct{}{}
		queue = append(queue, children[current]...)
	}
	return visible
}

func buildDistributionPaginationResult(total int64, params pagination.PaginationParams) *pagination.PaginationResult {
	page := params.Page
	if page < 1 {
		page = 1
	}
	pageSize := params.Limit()
	pages := 0
	if total > 0 {
		pages = int((total + int64(pageSize) - 1) / int64(pageSize))
	}
	return &pagination.PaginationResult{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Pages:    pages,
	}
}

func paginateDistributionViews[T any](items []T, params pagination.PaginationParams) []T {
	offset := params.Offset()
	if offset >= len(items) {
		return []T{}
	}
	limit := params.Limit()
	end := offset + limit
	if end > len(items) {
		end = len(items)
	}
	return items[offset:end]
}
