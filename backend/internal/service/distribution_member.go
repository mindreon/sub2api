package service

import (
	"context"
	"errors"
	"strings"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

var (
	ErrDistributionMemberNotFound         = infraerrors.NotFound("DISTRIBUTION_MEMBER_NOT_FOUND", "distribution member not found")
	ErrInvalidDistributionMember          = infraerrors.BadRequest("INVALID_DISTRIBUTION_MEMBER", "invalid distribution member")
	ErrDistributionMemberChannelConflict  = infraerrors.Conflict("DISTRIBUTION_MEMBER_CHANNEL_CONFLICT", "distribution member already belongs to another channel")
	ErrDistributionMemberParentForbidden  = infraerrors.Forbidden("DISTRIBUTION_MEMBER_PARENT_FORBIDDEN", "distribution member parent is not allowed")
	ErrDistributionMemberPermissionDenied = infraerrors.Forbidden("DISTRIBUTION_MEMBER_PERMISSION_DENIED", "distribution member permission denied")
	ErrDistributionMemberLimitExceeded    = infraerrors.Forbidden("DISTRIBUTION_MEMBER_LIMIT_EXCEEDED", "distribution member limit exceeded")
)

type DistributionMemberInput struct {
	ChannelOrgID   int64
	UserID         int64
	RoleType       string
	ParentMemberID *int64
	LevelCode      string
	CommissionRate float64
	Status         string
}

type DistributionMemberCreateRepository interface {
	ListByUserID(ctx context.Context, userID int64) ([]DistributionMemberView, error)
	GetByID(ctx context.Context, memberID int64) (*DistributionMemberView, error)
	CountByChannelOrgIDAndRole(ctx context.Context, channelOrgID int64, roleType string) (int64, error)
	Create(ctx context.Context, input DistributionMemberInput) (*DistributionMemberView, error)
}

type DistributionMemberService struct {
	repo             DistributionMemberCreateRepository
	settingService   *SettingService
	organizationRepo DistributionOrganizationLookupRepository
}

func NewDistributionMemberService(repo DistributionMemberCreateRepository) *DistributionMemberService {
	return &DistributionMemberService{repo: repo}
}

func (s *DistributionMemberService) SetSettingService(settingService *SettingService) {
	if s == nil {
		return
	}
	s.settingService = settingService
}

func (s *DistributionMemberService) SetOrganizationRepository(repo DistributionOrganizationLookupRepository) {
	if s == nil {
		return
	}
	s.organizationRepo = repo
}

func (s *DistributionMemberService) CreateMember(ctx context.Context, input DistributionMemberInput) (*DistributionMemberView, error) {
	if s == nil || s.repo == nil {
		return nil, ErrInvalidDistributionMember
	}
	input.RoleType = strings.ToLower(strings.TrimSpace(input.RoleType))
	input.Status = strings.ToLower(strings.TrimSpace(input.Status))
	input.LevelCode = strings.TrimSpace(input.LevelCode)
	if input.Status == "" {
		input.Status = "active"
	}
	if input.ChannelOrgID <= 0 || input.UserID <= 0 || input.CommissionRate < 0 || !isDistributionMemberRole(input.RoleType) {
		return nil, ErrInvalidDistributionMember
	}
	if !isDistributionMemberStatus(input.Status) {
		return nil, ErrInvalidDistributionMember
	}
	if input.LevelCode != "" {
		if rate, ok := s.resolveMemberLevelRate(ctx, input.ChannelOrgID, input.RoleType, input.LevelCode); ok {
			input.CommissionRate = rate
		}
	}
	if err := s.enforceMemberRoleLimit(ctx, input.ChannelOrgID, input.RoleType); err != nil {
		return nil, err
	}

	existing, err := s.repo.ListByUserID(ctx, input.UserID)
	if err != nil {
		return nil, err
	}
	for _, member := range existing {
		if member.ChannelOrgID > 0 && member.ChannelOrgID != input.ChannelOrgID {
			return nil, ErrDistributionMemberChannelConflict
		}
		if strings.EqualFold(strings.TrimSpace(member.RoleType), input.RoleType) {
			return nil, ErrDistributionMemberChannelConflict
		}
	}

	var parent *DistributionMemberView
	if input.ParentMemberID != nil && *input.ParentMemberID > 0 {
		parent, err = s.repo.GetByID(ctx, *input.ParentMemberID)
		if err != nil {
			if errors.Is(err, ErrDistributionMemberNotFound) {
				return nil, ErrDistributionMemberParentForbidden
			}
			return nil, err
		}
		if parent == nil || parent.ChannelOrgID != input.ChannelOrgID || !strings.EqualFold(strings.TrimSpace(parent.Status), "active") {
			return nil, ErrDistributionMemberParentForbidden
		}
		if !distributionMemberParentAllowsRole(parent.RoleType, input.RoleType) {
			return nil, ErrDistributionMemberParentForbidden
		}
		if input.LevelCode == "" {
			input.LevelCode = distributionMemberChildLevelCode(parent.LevelCode, input.RoleType)
		}
	} else {
		if input.RoleType != "manager" && input.RoleType != "agent" {
			return nil, ErrDistributionMemberParentForbidden
		}
		if input.LevelCode == "" {
			input.LevelCode = input.RoleType
		}
	}

	return s.repo.Create(ctx, input)
}

func (s *DistributionMemberService) CreateMemberForUser(ctx context.Context, currentUserID int64, input DistributionMemberInput) (*DistributionMemberView, error) {
	if s == nil || s.repo == nil {
		return nil, ErrInvalidDistributionMember
	}
	channelOrgID, err := resolveDistributionUserChannelOrgID(ctx, currentUserID, s.repo, s.organizationRepo, nil)
	if err != nil {
		return nil, err
	}

	callerMembers, err := s.repo.ListByUserID(ctx, currentUserID)
	if err != nil {
		return nil, err
	}

	canCreateAgent := false
	if s.organizationRepo != nil {
		org, err := s.organizationRepo.GetByOwnerUserID(ctx, currentUserID)
		if err != nil {
			return nil, err
		}
		if org != nil && org.ID == channelOrgID && strings.EqualFold(strings.TrimSpace(org.Status), "active") {
			canCreateAgent = true
		}
	}
	for _, member := range callerMembers {
		if member.ChannelOrgID != channelOrgID || !strings.EqualFold(strings.TrimSpace(member.Status), "active") {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(member.RoleType), "manager") {
			canCreateAgent = true
		}
	}

	input.ChannelOrgID = channelOrgID
	input.RoleType = strings.ToLower(strings.TrimSpace(input.RoleType))

	switch input.RoleType {
	case "agent":
		if input.ParentMemberID != nil && *input.ParentMemberID > 0 {
			return nil, ErrDistributionMemberPermissionDenied
		}
		if !canCreateAgent {
			return nil, ErrDistributionMemberPermissionDenied
		}
	case "kol1", "kol2":
		if input.ParentMemberID == nil || *input.ParentMemberID <= 0 {
			return nil, ErrDistributionMemberPermissionDenied
		}
		parent, err := s.repo.GetByID(ctx, *input.ParentMemberID)
		if err != nil {
			if errors.Is(err, ErrDistributionMemberNotFound) {
				return nil, ErrDistributionMemberPermissionDenied
			}
			return nil, err
		}
		if parent == nil || parent.ChannelOrgID != channelOrgID || parent.UserID != currentUserID || !strings.EqualFold(strings.TrimSpace(parent.Status), "active") {
			return nil, ErrDistributionMemberPermissionDenied
		}
		if !distributionMemberParentAllowsRole(parent.RoleType, input.RoleType) {
			return nil, ErrDistributionMemberPermissionDenied
		}
	default:
		return nil, ErrDistributionMemberPermissionDenied
	}

	return s.CreateMember(ctx, input)
}

func (s *DistributionMemberService) resolveMemberLevelRate(ctx context.Context, channelOrgID int64, roleType, levelCode string) (float64, bool) {
	levelCode = strings.TrimSpace(levelCode)
	if levelCode == "" {
		return 0, false
	}

	if strings.EqualFold(strings.TrimSpace(roleType), "kol2") && s.settingService != nil {
		return distributionLevelRateToPercent(s.settingService.GetDistributionKol2Rate(ctx)), true
	}

	if s.organizationRepo != nil && channelOrgID > 0 {
		if org, err := s.organizationRepo.GetByID(ctx, channelOrgID); err == nil && org != nil {
			if levels, err := parseDistributionLevelConfigs(org.Config["distribution_levels"]); err == nil {
				if cfg, ok := findDistributionLevelConfig(levels, levelCode); ok {
					return distributionLevelRateToPercent(cfg.CommissionRate), true
				}
			}
		}
	}
	if s.settingService != nil {
		if cfg, ok := findDistributionLevelConfig(s.settingService.GetDistributionGlobalLevels(ctx), levelCode); ok {
			return distributionLevelRateToPercent(cfg.CommissionRate), true
		}
	}
	return 0, false
}

func isDistributionMemberRole(role string) bool {
	switch role {
	case "manager", "agent", "kol1", "kol2":
		return true
	default:
		return false
	}
}

func isDistributionMemberStatus(status string) bool {
	switch status {
	case "active", "inactive", "disabled":
		return true
	default:
		return false
	}
}

func distributionMemberParentAllowsRole(parentRole, childRole string) bool {
	switch strings.ToLower(strings.TrimSpace(parentRole)) {
	case "manager":
		return childRole == "agent"
	case "agent":
		return childRole == "kol1"
	case "kol1":
		return childRole == "kol2"
	default:
		return false
	}
}

func distributionMemberChildLevelCode(parentLevel, childRole string) string {
	parentLevel = strings.TrimSpace(parentLevel)
	if parentLevel == "" {
		return childRole
	}
	next := parentLevel + "/" + childRole
	if len(next) > 20 {
		return childRole
	}
	return next
}

func (s *DistributionMemberService) enforceMemberRoleLimit(ctx context.Context, channelOrgID int64, roleType string) error {
	if s == nil || s.repo == nil || s.organizationRepo == nil || channelOrgID <= 0 {
		return nil
	}

	org, err := s.organizationRepo.GetByID(ctx, channelOrgID)
	if err != nil || org == nil {
		return err
	}

	limit := distributionMemberRoleLimit(org.Config, roleType)
	if limit <= 0 {
		return nil
	}

	currentCount, err := s.countMembersForLimit(ctx, channelOrgID, roleType)
	if err != nil {
		return err
	}
	if currentCount >= limit {
		return ErrDistributionMemberLimitExceeded
	}
	return nil
}

func (s *DistributionMemberService) countMembersForLimit(ctx context.Context, channelOrgID int64, roleType string) (int64, error) {
	switch roleType {
	case "kol1", "kol2":
		kol1Count, err := s.repo.CountByChannelOrgIDAndRole(ctx, channelOrgID, "kol1")
		if err != nil {
			return 0, err
		}
		kol2Count, err := s.repo.CountByChannelOrgIDAndRole(ctx, channelOrgID, "kol2")
		if err != nil {
			return 0, err
		}
		return kol1Count + kol2Count, nil
	default:
		return s.repo.CountByChannelOrgIDAndRole(ctx, channelOrgID, roleType)
	}
}

func distributionMemberRoleLimit(config map[string]any, roleType string) int64 {
	switch roleType {
	case "manager":
		return int64(distributionOrganizationConfigFloat(config, "max_manager_count"))
	case "agent":
		return int64(distributionOrganizationConfigFloat(config, "max_agent_count"))
	case "kol1", "kol2":
		return int64(distributionOrganizationConfigFloat(config, "max_kol_count"))
	default:
		return 0
	}
}
