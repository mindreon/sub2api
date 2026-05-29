package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type distributionAnalyticsRepoStub struct {
	channelSummary     *DistributionAnalyticsSummary
	channelTrend       []DistributionAnalyticsTrendPoint
	channelRanking     []DistributionAnalyticsRankingItem
	personalSummary    *DistributionAnalyticsSummary
	childMemberRanking []DistributionAnalyticsRankingItem

	channelSummaryOrgID int64
	personalMemberIDs   []int64
	childParentIDs      []int64
	userStatsMemberIDs  []int64
	attributedUserStats *DistributionAttributedUserStats
}

func (s *distributionAnalyticsRepoStub) GetChannelAnalyticsSummary(ctx context.Context, channelOrgID int64, filter DistributionAnalyticsFilter) (*DistributionAnalyticsSummary, error) {
	s.channelSummaryOrgID = channelOrgID
	return s.channelSummary, nil
}

func (s *distributionAnalyticsRepoStub) ListChannelTrend(ctx context.Context, channelOrgID int64, filter DistributionAnalyticsFilter) ([]DistributionAnalyticsTrendPoint, error) {
	return append([]DistributionAnalyticsTrendPoint(nil), s.channelTrend...), nil
}

func (s *distributionAnalyticsRepoStub) ListChannelMemberRanking(ctx context.Context, channelOrgID int64, filter DistributionAnalyticsFilter, limit int) ([]DistributionAnalyticsRankingItem, error) {
	return append([]DistributionAnalyticsRankingItem(nil), s.channelRanking...), nil
}

func (s *distributionAnalyticsRepoStub) GetMemberAnalyticsSummary(ctx context.Context, memberIDs []int64, filter DistributionAnalyticsFilter) (*DistributionAnalyticsSummary, error) {
	s.personalMemberIDs = append([]int64(nil), memberIDs...)
	return s.personalSummary, nil
}

func (s *distributionAnalyticsRepoStub) ListChildMemberRanking(ctx context.Context, parentMemberIDs []int64, filter DistributionAnalyticsFilter, limit int) ([]DistributionAnalyticsRankingItem, error) {
	s.childParentIDs = append([]int64(nil), parentMemberIDs...)
	return append([]DistributionAnalyticsRankingItem(nil), s.childMemberRanking...), nil
}

func (s *distributionAnalyticsRepoStub) GetAttributedUserStats(ctx context.Context, memberIDs []int64, filter DistributionAnalyticsFilter) (*DistributionAttributedUserStats, error) {
	s.userStatsMemberIDs = append([]int64(nil), memberIDs...)
	if s.attributedUserStats != nil {
		return s.attributedUserStats, nil
	}
	return &DistributionAttributedUserStats{}, nil
}

func TestDistributionAnalyticsService_GetAnalyticsForChannelManager(t *testing.T) {
	ownerUserID := int64(9)
	now := time.Date(2026, 5, 24, 0, 0, 0, 0, time.UTC)
	orgRepo := &distributionUserManageOrgRepoStub{
		orgByID: map[int64]*DistributionOrganization{
			88: {ID: 88, Type: "reseller", Name: "Channel A", OwnerUserID: &ownerUserID, Status: "active"},
		},
		orgByOwner: map[int64]*DistributionOrganization{
			9: {ID: 88, Type: "reseller", Name: "Channel A", OwnerUserID: &ownerUserID, Status: "active"},
		},
	}
	memberRepo := &distributionUserManageMemberRepoStub{
		byUser: map[int64][]DistributionMemberView{
			9: {
				{MemberID: 11, UserID: 9, ChannelOrgID: 88, RoleType: "manager", Status: "active"},
				{MemberID: 12, UserID: 9, ChannelOrgID: 88, RoleType: "kol1", Status: "active"},
			},
		},
	}
	analyticsRepo := &distributionAnalyticsRepoStub{
		channelSummary:  &DistributionAnalyticsSummary{RegisteredUsers: 18, ConsumptionAmount: 120.5, AgentCount: 3, Kol1Count: 4, Kol2Count: 2},
		channelTrend:    []DistributionAnalyticsTrendPoint{{Date: "2026-05-24", ConsumptionAmount: 12.5}},
		channelRanking:  []DistributionAnalyticsRankingItem{{MemberID: 21, ConsumptionAmount: 33.3}},
		personalSummary: &DistributionAnalyticsSummary{RegisteredUsers: 6, ConsumptionAmount: 22.2},
		childMemberRanking: []DistributionAnalyticsRankingItem{
			{MemberID: 31, RoleType: "kol2", ConsumptionAmount: 8.8},
		},
		attributedUserStats: &DistributionAttributedUserStats{TotalUsers: 33, NewUsers: 6},
	}

	svc := NewDistributionAnalyticsService(memberRepo, orgRepo, analyticsRepo)
	out, err := svc.GetAnalyticsForUser(context.Background(), 9, DistributionAnalyticsFilter{
		StartTime:   now.AddDate(0, 0, -6),
		EndTime:     now.AddDate(0, 0, 1),
		Granularity: "day",
		Limit:       10,
	})
	require.NoError(t, err)
	require.NotNil(t, out)
	require.True(t, out.CanManageChannel)
	require.NotNil(t, out.Channel)
	require.NotNil(t, out.Personal)
	require.Equal(t, int64(88), analyticsRepo.channelSummaryOrgID)
	require.Equal(t, []int64{12}, analyticsRepo.personalMemberIDs)
	require.Equal(t, []int64{12}, analyticsRepo.childParentIDs)
	require.Equal(t, []int64{12}, analyticsRepo.userStatsMemberIDs)
	require.Equal(t, int64(18), out.Channel.Summary.RegisteredUsers)
	require.Len(t, out.Channel.Trend, 1)
	require.Len(t, out.Channel.MemberRanking, 1)
	require.Empty(t, out.Channel.RoleBreakdown)
	require.Equal(t, int64(6), out.Personal.Summary.RegisteredUsers)
	require.Len(t, out.Personal.ChildMemberRanking, 1)
	require.Equal(t, int64(33), out.Personal.UserStats.TotalUsers)
	require.Equal(t, int64(6), out.Personal.UserStats.NewUsers)
	require.Equal(t, []string{"kol1"}, out.Personal.RoleTypes)
}

func TestDistributionAnalyticsService_GetAnalyticsForPromoterOnly(t *testing.T) {
	memberRepo := &distributionUserManageMemberRepoStub{
		byUser: map[int64][]DistributionMemberView{
			7: {
				{MemberID: 15, UserID: 7, ChannelOrgID: 88, RoleType: "agent", Status: "active"},
			},
		},
	}
	orgRepo := &distributionUserManageOrgRepoStub{
		orgByID: map[int64]*DistributionOrganization{
			88: {ID: 88, Type: "reseller", Name: "Channel A", Status: "active"},
		},
	}
	analyticsRepo := &distributionAnalyticsRepoStub{
		personalSummary: &DistributionAnalyticsSummary{RegisteredUsers: 4, ConsumptionAmount: 9.9},
		attributedUserStats: &DistributionAttributedUserStats{TotalUsers: 12, NewUsers: 4},
	}

	svc := NewDistributionAnalyticsService(memberRepo, orgRepo, analyticsRepo)
	out, err := svc.GetAnalyticsForUser(context.Background(), 7, DistributionAnalyticsFilter{
		StartTime:   time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
		EndTime:     time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		Granularity: "day",
		Limit:       5,
	})
	require.NoError(t, err)
	require.NotNil(t, out)
	require.False(t, out.CanManageChannel)
	require.Nil(t, out.Channel)
	require.NotNil(t, out.Personal)
	require.Equal(t, []int64{15}, analyticsRepo.personalMemberIDs)
	require.Equal(t, []int64{15}, analyticsRepo.userStatsMemberIDs)
	require.Empty(t, out.Personal.ChildMemberRanking)
	require.Equal(t, int64(12), out.Personal.UserStats.TotalUsers)
	require.Equal(t, int64(4), out.Personal.UserStats.NewUsers)
	require.Equal(t, []string{"agent"}, out.Personal.RoleTypes)
	require.Equal(t, []int64{15}, analyticsRepo.childParentIDs)
}

func TestDistributionAnalyticsService_GetAnalyticsForKol2HasNoChildRanking(t *testing.T) {
	memberRepo := &distributionUserManageMemberRepoStub{
		byUser: map[int64][]DistributionMemberView{
			7: {
				{MemberID: 18, UserID: 7, ChannelOrgID: 88, RoleType: "kol2", Status: "active"},
			},
		},
	}
	orgRepo := &distributionUserManageOrgRepoStub{
		orgByID: map[int64]*DistributionOrganization{
			88: {ID: 88, Type: "reseller", Name: "Channel A", Status: "active"},
		},
	}
	analyticsRepo := &distributionAnalyticsRepoStub{
		personalSummary: &DistributionAnalyticsSummary{RegisteredUsers: 2, ConsumptionAmount: 3.3},
	}

	svc := NewDistributionAnalyticsService(memberRepo, orgRepo, analyticsRepo)
	out, err := svc.GetAnalyticsForUser(context.Background(), 7, DistributionAnalyticsFilter{
		StartTime:   time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
		EndTime:     time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		Granularity: "day",
		Limit:       5,
	})
	require.NoError(t, err)
	require.NotNil(t, out)
	require.NotNil(t, out.Personal)
	require.Equal(t, []int64{18}, analyticsRepo.personalMemberIDs)
	require.Equal(t, []int64{18}, analyticsRepo.userStatsMemberIDs)
	require.Empty(t, analyticsRepo.childParentIDs)
	require.Empty(t, out.Personal.ChildMemberRanking)
}

func TestBuildDistributionRoleBreakdown(t *testing.T) {
	out := buildDistributionRoleBreakdown([]DistributionAnalyticsRankingItem{
		{RoleType: "agent", RegisteredUsers: 5, ConsumptionAmount: 100, CommissionAmount: 10, SettledCommissionAmount: 6},
		{RoleType: "agent", RegisteredUsers: 3, ConsumptionAmount: 80, CommissionAmount: 8, SettledCommissionAmount: 4},
		{RoleType: "kol1", RegisteredUsers: 4, ConsumptionAmount: 50, CommissionAmount: 5, SettledCommissionAmount: 3},
	})
	require.Len(t, out, 2)
	require.Equal(t, "agent", out[0].RoleType)
	require.Equal(t, int64(2), out[0].MemberCount)
	require.Equal(t, int64(8), out[0].RegisteredUsers)
	require.InDelta(t, 180, out[0].ConsumptionAmount, 0.0001)
	require.Equal(t, "kol1", out[1].RoleType)
	require.Equal(t, int64(1), out[1].MemberCount)
}
