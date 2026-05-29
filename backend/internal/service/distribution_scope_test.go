//go:build unit

package service

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

type distributionScopeAttributionRepoStub struct {
	attribution *DistributionAttribution
	channelID   int64
	params      pagination.PaginationParams
	rows        []DistributionAttributionView
}

func (s *distributionScopeAttributionRepoStub) GetByUserID(ctx context.Context, userID int64) (*DistributionAttribution, error) {
	return s.attribution, nil
}

func (s *distributionScopeAttributionRepoStub) ListByChannelOrgID(ctx context.Context, channelOrgID int64, params pagination.PaginationParams) ([]DistributionAttributionView, *pagination.PaginationResult, error) {
	s.channelID = channelOrgID
	s.params = params
	items := s.rows
	if len(items) == 0 {
		items = []DistributionAttributionView{{UserID: 2, ChannelOrgID: channelOrgID}}
	}
	offset := params.Offset()
	if offset > len(items) {
		offset = len(items)
	}
	limit := params.Limit()
	end := offset + limit
	if end > len(items) {
		end = len(items)
	}
	return items[offset:end], &pagination.PaginationResult{Total: int64(len(items)), Page: params.Page, PageSize: params.PageSize, Pages: (len(items) + params.PageSize - 1) / params.PageSize}, nil
}

type distributionScopeMemberRepoStub struct {
	channelID int64
	roleType  string
	params    pagination.PaginationParams
	byUser    []DistributionMemberView
	rows      []DistributionMemberView
}

func (s *distributionScopeMemberRepoStub) ListByChannelOrgID(ctx context.Context, channelOrgID int64, params pagination.PaginationParams, roleType string) ([]DistributionMemberView, *pagination.PaginationResult, error) {
	s.channelID = channelOrgID
	s.roleType = roleType
	s.params = params
	items := s.rows
	if len(items) == 0 {
		items = []DistributionMemberView{{MemberID: 3, ChannelOrgID: channelOrgID}}
	}
	if roleType != "" {
		filtered := make([]DistributionMemberView, 0, len(items))
		for _, item := range items {
			if item.RoleType == roleType {
				filtered = append(filtered, item)
			}
		}
		items = filtered
	}
	offset := params.Offset()
	if offset > len(items) {
		offset = len(items)
	}
	limit := params.Limit()
	end := offset + limit
	if end > len(items) {
		end = len(items)
	}
	return items[offset:end], &pagination.PaginationResult{Total: int64(len(items)), Page: params.Page, PageSize: params.PageSize, Pages: (len(items) + params.PageSize - 1) / params.PageSize}, nil
}

func (s *distributionScopeMemberRepoStub) ListByUserID(ctx context.Context, userID int64) ([]DistributionMemberView, error) {
	return s.byUser, nil
}

type distributionScopeOrganizationRepoStub struct {
	orgByOwner map[int64]*DistributionOrganization
}

func (s *distributionScopeOrganizationRepoStub) GetByOwnerUserID(ctx context.Context, userID int64) (*DistributionOrganization, error) {
	if s.orgByOwner == nil {
		return nil, nil
	}
	return s.orgByOwner[userID], nil
}

type distributionScopeCommissionRepoStub struct {
	channelID int64
	params    pagination.PaginationParams
	rows      []DistributionCommissionLedgerView
}

func (s *distributionScopeCommissionRepoStub) ListByChannelOrgID(ctx context.Context, channelOrgID int64, params pagination.PaginationParams) ([]DistributionCommissionLedgerView, *pagination.PaginationResult, error) {
	s.channelID = channelOrgID
	s.params = params
	items := s.rows
	if len(items) == 0 {
		now := time.Now().UTC()
		items = []DistributionCommissionLedgerView{{ID: 4, ChannelOrgID: channelOrgID, MemberID: 3, CreatedAt: now, UpdatedAt: now}}
	}
	offset := params.Offset()
	if offset > len(items) {
		offset = len(items)
	}
	limit := params.Limit()
	end := offset + limit
	if end > len(items) {
		end = len(items)
	}
	return items[offset:end], &pagination.PaginationResult{Total: int64(len(items)), Page: params.Page, PageSize: params.PageSize, Pages: (len(items) + params.PageSize - 1) / params.PageSize}, nil
}

type distributionScopeWalletTransactionRepoStub struct {
	channelID       int64
	transactionType string
	params          pagination.PaginationParams
}

func (s *distributionScopeWalletTransactionRepoStub) ListTransactions(ctx context.Context, filter DistributionAdminListFilter, params pagination.PaginationParams) ([]DistributionWalletTransaction, *pagination.PaginationResult, error) {
	return nil, nil, nil
}

func (s *distributionScopeWalletTransactionRepoStub) ListTransactionsByChannelOrgID(ctx context.Context, channelOrgID int64, params pagination.PaginationParams, transactionType string) ([]DistributionWalletTransaction, *pagination.PaginationResult, error) {
	s.channelID = channelOrgID
	s.transactionType = transactionType
	s.params = params
	return []DistributionWalletTransaction{{ID: 5, ChannelOrgID: channelOrgID, TransactionType: transactionType}}, &pagination.PaginationResult{Total: 1, Page: params.Page, PageSize: params.PageSize, Pages: 1}, nil
}

func TestDistributionScopeService_ScopesListsByUserChannel(t *testing.T) {
	attrRepo := &distributionScopeAttributionRepoStub{
		attribution: &DistributionAttribution{UserID: 1, ChannelOrgID: 77},
	}
	memberRepo := &distributionScopeMemberRepoStub{}
	commissionRepo := &distributionScopeCommissionRepoStub{}
	walletTxRepo := &distributionScopeWalletTransactionRepoStub{}

	svc := NewDistributionScopeService(attrRepo, memberRepo, commissionRepo, nil, nil)
	svc.SetWalletTransactionRepository(walletTxRepo)

	channelID, err := svc.ResolveUserChannelOrgID(context.Background(), 1)
	require.NoError(t, err)
	require.Equal(t, int64(77), channelID)

	members, _, err := svc.ListMembersForUser(context.Background(), 1, pagination.PaginationParams{Page: 2, PageSize: 10}, "kol1")
	require.NoError(t, err)
	require.Len(t, members, 1)
	require.Equal(t, int64(77), memberRepo.channelID)
	require.Equal(t, "kol1", memberRepo.roleType)
	require.Equal(t, 2, memberRepo.params.Page)

	attributions, _, err := svc.ListAttributionsForUser(context.Background(), 1, pagination.PaginationParams{Page: 1, PageSize: 5})
	require.NoError(t, err)
	require.Len(t, attributions, 1)
	require.Equal(t, int64(77), attrRepo.channelID)

	commissions, _, err := svc.ListCommissionsForUser(context.Background(), 1, pagination.PaginationParams{Page: 1, PageSize: 5})
	require.NoError(t, err)
	require.Len(t, commissions, 1)
	require.Equal(t, int64(77), commissionRepo.channelID)

	transactions, _, err := svc.ListWalletTransactionsForUser(context.Background(), 1, pagination.PaginationParams{Page: 3, PageSize: 5}, "consume")
	require.NoError(t, err)
	require.Len(t, transactions, 1)
	require.Equal(t, int64(77), walletTxRepo.channelID)
	require.Equal(t, "consume", walletTxRepo.transactionType)
	require.Equal(t, 3, walletTxRepo.params.Page)
}

func TestDistributionScopeService_ResolvesChannelFromMemberBeforeAttribution(t *testing.T) {
	attrRepo := &distributionScopeAttributionRepoStub{
		attribution: &DistributionAttribution{UserID: 1, ChannelOrgID: 77},
	}
	memberRepo := &distributionScopeMemberRepoStub{
		byUser: []DistributionMemberView{
			{MemberID: 3, UserID: 1, ChannelOrgID: 88, RoleType: "agent", Status: "active"},
		},
	}

	svc := NewDistributionScopeService(attrRepo, memberRepo, nil, nil, nil)

	channelID, err := svc.ResolveUserChannelOrgID(context.Background(), 1)
	require.NoError(t, err)
	require.Equal(t, int64(88), channelID)
}

func TestDistributionScopeService_ResolvesChannelFromOwnerWhenNoMemberOrAttribution(t *testing.T) {
	attrRepo := &distributionScopeAttributionRepoStub{}
	memberRepo := &distributionScopeMemberRepoStub{}
	svc := NewDistributionScopeService(attrRepo, memberRepo, nil, nil, nil)
	svc.SetOrganizationRepository(&distributionScopeOrganizationRepoStub{
		orgByOwner: map[int64]*DistributionOrganization{
			9: {ID: 66, Type: "reseller", Name: "Channel Owner"},
		},
	})

	channelID, err := svc.ResolveUserChannelOrgID(context.Background(), 9)
	require.NoError(t, err)
	require.Equal(t, int64(66), channelID)
}

func TestDistributionScopeService_CanManageChannelForUser(t *testing.T) {
	attrRepo := &distributionScopeAttributionRepoStub{}
	memberRepo := &distributionScopeMemberRepoStub{
		byUser: []DistributionMemberView{
			{MemberID: 3, UserID: 1, ChannelOrgID: 88, RoleType: "manager", Status: "active"},
		},
	}
	svc := NewDistributionScopeService(attrRepo, memberRepo, nil, nil, nil)
	svc.SetOrganizationRepository(&distributionScopeOrganizationRepoStub{})

	ok, err := svc.CanManageChannelForUser(context.Background(), 1)
	require.NoError(t, err)
	require.True(t, ok)
}

func TestDistributionScopeService_FiltersByMemberTreeForNonManager(t *testing.T) {
	attrRepo := &distributionScopeAttributionRepoStub{
		attribution: &DistributionAttribution{UserID: 9, ChannelOrgID: 88},
		rows: []DistributionAttributionView{
			{UserID: 101, ChannelOrgID: 88, ReferrerMemberID: ptrInt64(20)},
			{UserID: 102, ChannelOrgID: 88, ReferrerMemberID: ptrInt64(30)},
		},
	}
	memberRepo := &distributionScopeMemberRepoStub{
		byUser: []DistributionMemberView{
			{MemberID: 20, UserID: 9, ChannelOrgID: 88, RoleType: "kol1", Status: "active"},
		},
		rows: []DistributionMemberView{
			{MemberID: 10, UserID: 1, ChannelOrgID: 88, RoleType: "agent", Status: "active"},
			{MemberID: 20, UserID: 9, ChannelOrgID: 88, RoleType: "kol1", ParentMemberID: ptrInt64(10), Status: "active"},
			{MemberID: 21, UserID: 10, ChannelOrgID: 88, RoleType: "kol2", ParentMemberID: ptrInt64(20), Status: "active"},
			{MemberID: 30, UserID: 11, ChannelOrgID: 88, RoleType: "kol1", ParentMemberID: ptrInt64(10), Status: "active"},
		},
	}
	commissionRepo := &distributionScopeCommissionRepoStub{
		rows: []DistributionCommissionLedgerView{
			{ID: 1, ChannelOrgID: 88, MemberID: 20, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()},
			{ID: 2, ChannelOrgID: 88, MemberID: 21, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()},
			{ID: 3, ChannelOrgID: 88, MemberID: 30, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()},
		},
	}
	svc := NewDistributionScopeService(attrRepo, memberRepo, commissionRepo, nil, nil)
	svc.SetOrganizationRepository(&distributionScopeOrganizationRepoStub{})

	members, memberPage, err := svc.ListMembersForUser(context.Background(), 9, pagination.PaginationParams{Page: 1, PageSize: 20}, "")
	require.NoError(t, err)
	require.Len(t, members, 2)
	require.Equal(t, int64(2), memberPage.Total)

	attributions, attrPage, err := svc.ListAttributionsForUser(context.Background(), 9, pagination.PaginationParams{Page: 1, PageSize: 20})
	require.NoError(t, err)
	require.Len(t, attributions, 1)
	require.Equal(t, int64(1), attrPage.Total)
	require.Equal(t, int64(20), *attributions[0].ReferrerMemberID)

	commissions, comPage, err := svc.ListCommissionsForUser(context.Background(), 9, pagination.PaginationParams{Page: 1, PageSize: 20})
	require.NoError(t, err)
	require.Len(t, commissions, 2)
	require.Equal(t, int64(2), comPage.Total)
}

func TestDistributionScopeService_ListWalletTransactionsRequiresManagerScope(t *testing.T) {
	attrRepo := &distributionScopeAttributionRepoStub{
		attribution: &DistributionAttribution{UserID: 9, ChannelOrgID: 88},
	}
	memberRepo := &distributionScopeMemberRepoStub{
		byUser: []DistributionMemberView{
			{MemberID: 20, UserID: 9, ChannelOrgID: 88, RoleType: "kol1", Status: "active"},
		},
	}
	walletTxRepo := &distributionScopeWalletTransactionRepoStub{}
	svc := NewDistributionScopeService(attrRepo, memberRepo, nil, nil, nil)
	svc.SetOrganizationRepository(&distributionScopeOrganizationRepoStub{})
	svc.SetWalletTransactionRepository(walletTxRepo)

	_, _, err := svc.ListWalletTransactionsForUser(context.Background(), 9, pagination.PaginationParams{Page: 1, PageSize: 20}, "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "DISTRIBUTION_CHANNEL_PERMISSION_DENIED")
}
