package service

import (
	"context"
	"sort"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

type DistributionAnalyticsSummary struct {
	RegisteredUsers         int64   `json:"registered_users"`
	RechargeAmount          float64 `json:"recharge_amount"`
	ConsumptionAmount       float64 `json:"consumption_amount"`
	CommissionAmount        float64 `json:"commission_amount"`
	SettledCommissionAmount float64 `json:"settled_commission_amount"`
	MemberCount             int64   `json:"member_count"`
	AgentCount              int64   `json:"agent_count"`
	Kol1Count               int64   `json:"kol1_count"`
	Kol2Count               int64   `json:"kol2_count"`
	CommissionExpenseRatio  float64 `json:"commission_expense_ratio"`
	CommissionUpperRatio    float64 `json:"commission_upper_ratio"`
}

type DistributionAnalyticsTrendPoint struct {
	Date                    string  `json:"date"`
	RegisteredUsers         int64   `json:"registered_users"`
	RechargeAmount          float64 `json:"recharge_amount"`
	ConsumptionAmount       float64 `json:"consumption_amount"`
	CommissionAmount        float64 `json:"commission_amount"`
	SettledCommissionAmount float64 `json:"settled_commission_amount"`
}

type DistributionAnalyticsRankingItem struct {
	MemberID                int64   `json:"member_id"`
	UserID                  int64   `json:"user_id"`
	UserEmail               string  `json:"user_email"`
	Username                string  `json:"username"`
	RoleType                string  `json:"role_type"`
	RegisteredUsers         int64   `json:"registered_users"`
	RechargeAmount          float64 `json:"recharge_amount"`
	ConsumptionAmount       float64 `json:"consumption_amount"`
	CommissionAmount        float64 `json:"commission_amount"`
	SettledCommissionAmount float64 `json:"settled_commission_amount"`
}

type DistributionAnalyticsRoleBreakdownItem struct {
	RoleType                string  `json:"role_type"`
	MemberCount             int64   `json:"member_count"`
	RegisteredUsers         int64   `json:"registered_users"`
	ConsumptionAmount       float64 `json:"consumption_amount"`
	CommissionAmount        float64 `json:"commission_amount"`
	SettledCommissionAmount float64 `json:"settled_commission_amount"`
}

type DistributionAttributedUserStats struct {
	TotalUsers int64 `json:"total_users"`
	NewUsers   int64 `json:"new_users"`
}

type DistributionAnalyticsFilter struct {
	StartTime   time.Time `json:"-"`
	EndTime     time.Time `json:"-"`
	Granularity string    `json:"granularity"`
	Limit       int       `json:"limit"`
}

type DistributionAnalyticsFilterView struct {
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
	Granularity string `json:"granularity"`
	Limit       int    `json:"limit"`
}

type DistributionAnalyticsChannel struct {
	Summary       DistributionAnalyticsSummary       `json:"summary"`
	Trend         []DistributionAnalyticsTrendPoint  `json:"trend"`
	MemberRanking []DistributionAnalyticsRankingItem `json:"member_ranking"`
	RoleBreakdown []DistributionAnalyticsRoleBreakdownItem `json:"role_breakdown"`
}

type DistributionAnalyticsPersonal struct {
	RoleTypes          []string                           `json:"role_types"`
	Summary            DistributionAnalyticsSummary       `json:"summary"`
	ChildMemberRanking []DistributionAnalyticsRankingItem `json:"child_member_ranking"`
	UserStats          DistributionAttributedUserStats    `json:"user_stats"`
}

type DistributionAnalyticsResponse struct {
	CanManageChannel bool                            `json:"can_manage_channel"`
	Filter           DistributionAnalyticsFilterView `json:"filter"`
	Channel          *DistributionAnalyticsChannel   `json:"channel,omitempty"`
	Personal         *DistributionAnalyticsPersonal  `json:"personal,omitempty"`
}

type DistributionAnalyticsRepository interface {
	GetChannelAnalyticsSummary(ctx context.Context, channelOrgID int64, filter DistributionAnalyticsFilter) (*DistributionAnalyticsSummary, error)
	ListChannelTrend(ctx context.Context, channelOrgID int64, filter DistributionAnalyticsFilter) ([]DistributionAnalyticsTrendPoint, error)
	ListChannelMemberRanking(ctx context.Context, channelOrgID int64, filter DistributionAnalyticsFilter, limit int) ([]DistributionAnalyticsRankingItem, error)
	GetMemberAnalyticsSummary(ctx context.Context, memberIDs []int64, filter DistributionAnalyticsFilter) (*DistributionAnalyticsSummary, error)
	ListChildMemberRanking(ctx context.Context, parentMemberIDs []int64, filter DistributionAnalyticsFilter, limit int) ([]DistributionAnalyticsRankingItem, error)
	GetAttributedUserStats(ctx context.Context, memberIDs []int64, filter DistributionAnalyticsFilter) (*DistributionAttributedUserStats, error)
}

type DistributionAnalyticsService struct {
	memberRepo       DistributionUserManageMemberRepository
	organizationRepo DistributionUserManageOrganizationRepository
	analyticsRepo    DistributionAnalyticsRepository
}

func NewDistributionAnalyticsService(
	memberRepo DistributionUserManageMemberRepository,
	organizationRepo DistributionUserManageOrganizationRepository,
	analyticsRepo DistributionAnalyticsRepository,
) *DistributionAnalyticsService {
	return &DistributionAnalyticsService{
		memberRepo:       memberRepo,
		organizationRepo: organizationRepo,
		analyticsRepo:    analyticsRepo,
	}
}

func (s *DistributionAnalyticsService) GetAnalyticsForUser(
	ctx context.Context,
	userID int64,
	filter DistributionAnalyticsFilter,
) (*DistributionAnalyticsResponse, error) {
	if s == nil || s.analyticsRepo == nil {
		return nil, infraerrors.ServiceUnavailable("DISTRIBUTION_ANALYTICS_NOT_READY", "distribution analytics is not ready")
	}

	filter, err := normalizeDistributionAnalyticsFilter(filter)
	if err != nil {
		return nil, err
	}

	channelOrgID, err := resolveDistributionUserChannelOrgID(ctx, userID, s.memberRepo, s.organizationRepo, nil)
	if err != nil {
		return nil, err
	}

	canManageChannel, err := s.canManageChannelOrgID(ctx, userID, channelOrgID)
	if err != nil {
		return nil, err
	}

	resp := &DistributionAnalyticsResponse{
		CanManageChannel: canManageChannel,
		Filter: DistributionAnalyticsFilterView{
			StartDate:   filter.StartTime.Format("2006-01-02"),
			EndDate:     filter.EndTime.Add(-time.Nanosecond).Format("2006-01-02"),
			Granularity: filter.Granularity,
			Limit:       filter.Limit,
		},
	}

	members, err := s.listUserMembers(ctx, userID)
	if err != nil {
		return nil, err
	}
	promoterMemberIDs, roleTypes, childRankingParentIDs := collectDistributionAnalyticsMemberIDs(members, channelOrgID)

	if canManageChannel {
		channel := &DistributionAnalyticsChannel{}
		channel.Summary, err = s.channelSummary(ctx, channelOrgID, filter)
		if err != nil {
			return nil, err
		}
		channel.Trend, err = s.analyticsRepo.ListChannelTrend(ctx, channelOrgID, filter)
		if err != nil {
			return nil, err
		}
		channel.MemberRanking, err = s.analyticsRepo.ListChannelMemberRanking(ctx, channelOrgID, filter, filter.Limit)
		if err != nil {
			return nil, err
		}
		channel.RoleBreakdown = buildDistributionRoleBreakdown(channel.MemberRanking)
		resp.Channel = channel
	}

	if len(promoterMemberIDs) > 0 {
		personal := &DistributionAnalyticsPersonal{
			RoleTypes: roleTypes,
		}
		personal.Summary, err = s.memberSummary(ctx, promoterMemberIDs, filter)
		if err != nil {
			return nil, err
		}
		userStats, err := s.analyticsRepo.GetAttributedUserStats(ctx, promoterMemberIDs, filter)
		if err != nil {
			return nil, err
		}
		if userStats != nil {
			personal.UserStats = *userStats
		}
		if len(childRankingParentIDs) > 0 {
			personal.ChildMemberRanking, err = s.analyticsRepo.ListChildMemberRanking(ctx, childRankingParentIDs, filter, filter.Limit)
			if err != nil {
				return nil, err
			}
		} else {
			personal.ChildMemberRanking = []DistributionAnalyticsRankingItem{}
		}
		resp.Personal = personal
	}

	return resp, nil
}

func (s *DistributionAnalyticsService) listUserMembers(ctx context.Context, userID int64) ([]DistributionMemberView, error) {
	if s == nil || s.memberRepo == nil {
		return nil, nil
	}
	return s.memberRepo.ListByUserID(ctx, userID)
}

func (s *DistributionAnalyticsService) canManageChannelOrgID(ctx context.Context, userID int64, channelOrgID int64) (bool, error) {
	if userID <= 0 || channelOrgID <= 0 {
		return false, ErrInvalidDistributionOrganization
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
	if s.memberRepo == nil {
		return false, nil
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

func (s *DistributionAnalyticsService) channelSummary(ctx context.Context, channelOrgID int64, filter DistributionAnalyticsFilter) (DistributionAnalyticsSummary, error) {
	summary, err := s.analyticsRepo.GetChannelAnalyticsSummary(ctx, channelOrgID, filter)
	if err != nil {
		return DistributionAnalyticsSummary{}, err
	}
	if summary == nil {
		return DistributionAnalyticsSummary{}, nil
	}
	return *summary, nil
}

func (s *DistributionAnalyticsService) memberSummary(ctx context.Context, memberIDs []int64, filter DistributionAnalyticsFilter) (DistributionAnalyticsSummary, error) {
	summary, err := s.analyticsRepo.GetMemberAnalyticsSummary(ctx, memberIDs, filter)
	if err != nil {
		return DistributionAnalyticsSummary{}, err
	}
	if summary == nil {
		return DistributionAnalyticsSummary{}, nil
	}
	return *summary, nil
}

func normalizeDistributionAnalyticsFilter(filter DistributionAnalyticsFilter) (DistributionAnalyticsFilter, error) {
	if filter.StartTime.IsZero() || filter.EndTime.IsZero() || !filter.EndTime.After(filter.StartTime) {
		return DistributionAnalyticsFilter{}, ErrInvalidDistributionStats
	}
	granularity := strings.ToLower(strings.TrimSpace(filter.Granularity))
	switch granularity {
	case "", "day":
		granularity = "day"
	case "hour", "week", "month":
	default:
		return DistributionAnalyticsFilter{}, ErrInvalidDistributionStats
	}
	filter.Granularity = granularity
	if filter.Limit <= 0 {
		filter.Limit = 10
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}
	return filter, nil
}

func collectDistributionAnalyticsMemberIDs(members []DistributionMemberView, channelOrgID int64) ([]int64, []string, []int64) {
	memberIDs := make([]int64, 0, len(members))
	childRankingParentIDs := make([]int64, 0, len(members))
	roleSet := make(map[string]struct{}, 3)
	for _, member := range members {
		if member.ChannelOrgID != channelOrgID || !strings.EqualFold(strings.TrimSpace(member.Status), "active") {
			continue
		}
		role := strings.ToLower(strings.TrimSpace(member.RoleType))
		switch role {
		case "agent", "kol1", "kol2":
			memberIDs = append(memberIDs, member.MemberID)
			roleSet[role] = struct{}{}
			if role == "agent" || role == "kol1" {
				childRankingParentIDs = append(childRankingParentIDs, member.MemberID)
			}
		}
	}
	sort.Slice(memberIDs, func(i, j int) bool { return memberIDs[i] < memberIDs[j] })
	sort.Slice(childRankingParentIDs, func(i, j int) bool { return childRankingParentIDs[i] < childRankingParentIDs[j] })
	roleTypes := make([]string, 0, len(roleSet))
	for _, role := range []string{"agent", "kol1", "kol2"} {
		if _, ok := roleSet[role]; ok {
			roleTypes = append(roleTypes, role)
		}
	}
	return memberIDs, roleTypes, childRankingParentIDs
}

func buildDistributionRoleBreakdown(items []DistributionAnalyticsRankingItem) []DistributionAnalyticsRoleBreakdownItem {
	if len(items) == 0 {
		return []DistributionAnalyticsRoleBreakdownItem{}
	}
	byRole := map[string]*DistributionAnalyticsRoleBreakdownItem{
		"agent": {RoleType: "agent"},
		"kol1":  {RoleType: "kol1"},
		"kol2":  {RoleType: "kol2"},
	}
	for _, item := range items {
		role := strings.ToLower(strings.TrimSpace(item.RoleType))
		bucket, ok := byRole[role]
		if !ok {
			continue
		}
		bucket.MemberCount++
		bucket.RegisteredUsers += item.RegisteredUsers
		bucket.ConsumptionAmount += item.ConsumptionAmount
		bucket.CommissionAmount += item.CommissionAmount
		bucket.SettledCommissionAmount += item.SettledCommissionAmount
	}
	out := make([]DistributionAnalyticsRoleBreakdownItem, 0, 3)
	for _, role := range []string{"agent", "kol1", "kol2"} {
		bucket := byRole[role]
		if bucket.MemberCount == 0 {
			continue
		}
		out = append(out, *bucket)
	}
	return out
}
