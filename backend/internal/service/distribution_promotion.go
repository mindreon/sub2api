package service

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"errors"
	"sort"
	"strings"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

type DistributionPromotionLinkInput struct {
	ChannelOrgID int64
	MemberID     int64
	Code         string
	TargetType   string
	Status       string
}

type DistributionPromotionLinkRepository interface {
	ListByChannelOrgID(ctx context.Context, channelOrgID int64, params pagination.PaginationParams) ([]DistributionPromotionLink, *pagination.PaginationResult, error)
	ListAdmin(ctx context.Context, filter DistributionAdminListFilter, params pagination.PaginationParams) ([]DistributionPromotionLink, *pagination.PaginationResult, error)
	Create(ctx context.Context, input DistributionPromotionLinkInput) (*DistributionPromotionLink, error)
}

type DistributionPromotionMemberRepository interface {
	ListByUserID(ctx context.Context, userID int64) ([]DistributionMemberView, error)
	GetByID(ctx context.Context, memberID int64) (*DistributionMemberView, error)
}

type DistributionPromotionAttributionRepository interface {
	GetByUserID(ctx context.Context, userID int64) (*DistributionAttribution, error)
}

type DistributionPromotionService struct {
	repo             DistributionPromotionLinkRepository
	memberRepo       DistributionPromotionMemberRepository
	attributionRepo  DistributionPromotionAttributionRepository
	organizationRepo DistributionUserChannelOrganizationRepository
}

func NewDistributionPromotionService(
	repo DistributionPromotionLinkRepository,
	memberRepo DistributionPromotionMemberRepository,
	attributionRepo DistributionPromotionAttributionRepository,
) *DistributionPromotionService {
	return &DistributionPromotionService{
		repo:            repo,
		memberRepo:      memberRepo,
		attributionRepo: attributionRepo,
	}
}

func (s *DistributionPromotionService) SetOrganizationRepository(repo DistributionUserChannelOrganizationRepository) {
	if s == nil {
		return
	}
	s.organizationRepo = repo
}

func (s *DistributionPromotionService) ListLinks(ctx context.Context, filter DistributionAdminListFilter, params pagination.PaginationParams) ([]DistributionPromotionLink, *pagination.PaginationResult, error) {
	if s == nil || s.repo == nil {
		return nil, nil, ErrInvalidDistributionPromotionLink
	}
	return s.repo.ListAdmin(ctx, filter.normalized(), params)
}

func (s *DistributionPromotionService) ListLinksForUser(ctx context.Context, userID int64, params pagination.PaginationParams) ([]DistributionPromotionLink, *pagination.PaginationResult, error) {
	if s == nil || s.repo == nil {
		return nil, nil, ErrInvalidDistributionPromotionLink
	}
	channelOrgID, err := resolveDistributionUserChannelOrgID(ctx, userID, s.memberRepo, s.organizationRepo, s.attributionRepo)
	if err != nil {
		return nil, nil, err
	}
	canManage, err := distributionCanManageChannelWithPromotionRepos(ctx, userID, channelOrgID, s.memberRepo, s.organizationRepo)
	if err != nil {
		return nil, nil, err
	}
	if canManage {
		return s.repo.ListByChannelOrgID(ctx, channelOrgID, params)
	}

	myMembers, err := s.memberRepo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	rootMemberIDs := make([]int64, 0, len(myMembers))
	for _, member := range myMembers {
		if member.ChannelOrgID != channelOrgID || !strings.EqualFold(strings.TrimSpace(member.Status), "active") {
			continue
		}
		rootMemberIDs = append(rootMemberIDs, member.MemberID)
	}
	if len(rootMemberIDs) == 0 {
		return []DistributionPromotionLink{}, distributionPromotionPaginationResult(0, params), nil
	}

	allLinks, err := s.listAllLinksByChannel(ctx, channelOrgID)
	if err != nil {
		return nil, nil, err
	}
	memberCache := make(map[int64]*DistributionMemberView, len(rootMemberIDs))
	filtered := make([]DistributionPromotionLink, 0, len(allLinks))
	for _, link := range allLinks {
		ok, err := s.distributionMemberIsSelfOrDescendant(ctx, channelOrgID, link.MemberID, rootMemberIDs, memberCache)
		if err != nil {
			return nil, nil, err
		}
		if ok {
			filtered = append(filtered, link)
		}
	}
	return paginateDistributionPromotionLinks(filtered, params), distributionPromotionPaginationResult(int64(len(filtered)), params), nil
}

func (s *DistributionPromotionService) CreateLink(ctx context.Context, input DistributionPromotionLinkInput) (*DistributionPromotionLink, error) {
	if s == nil || s.repo == nil || s.memberRepo == nil {
		return nil, ErrInvalidDistributionPromotionLink
	}
	normalized, err := normalizeDistributionPromotionLinkInput(input)
	if err != nil {
		return nil, err
	}
	member, err := s.memberRepo.GetByID(ctx, normalized.MemberID)
	if err != nil {
		if errors.Is(err, ErrDistributionMemberNotFound) {
			return nil, ErrInvalidDistributionPromotionLink
		}
		return nil, err
	}
	if member == nil || member.ChannelOrgID <= 0 || !strings.EqualFold(strings.TrimSpace(member.Status), "active") {
		return nil, ErrInvalidDistributionPromotionLink
	}
	normalized.ChannelOrgID = member.ChannelOrgID
	return s.repo.Create(ctx, normalized)
}

func (s *DistributionPromotionService) CreateLinkForUser(ctx context.Context, userID int64, input DistributionPromotionLinkInput) (*DistributionPromotionLink, error) {
	if s == nil || s.repo == nil || s.memberRepo == nil {
		return nil, ErrInvalidDistributionPromotionLink
	}
	if userID <= 0 {
		return nil, ErrInvalidDistributionPromotionLink
	}
	channelOrgID, err := resolveDistributionUserChannelOrgID(ctx, userID, s.memberRepo, s.organizationRepo, s.attributionRepo)
	if err != nil {
		if errors.Is(err, ErrDistributionAttributionNotFound) {
			return nil, ErrInvalidDistributionPromotionLink
		}
		return nil, err
	}

	normalized, err := normalizeDistributionPromotionLinkInput(input)
	if err != nil {
		return nil, err
	}
	member, err := s.memberRepo.GetByID(ctx, normalized.MemberID)
	if err != nil {
		if errors.Is(err, ErrDistributionMemberNotFound) {
			return nil, ErrInvalidDistributionPromotionLink
		}
		return nil, err
	}
	if member == nil || member.ChannelOrgID != channelOrgID || !strings.EqualFold(strings.TrimSpace(member.Status), "active") {
		return nil, ErrInvalidDistributionPromotionLink
	}
	canManage, err := distributionCanManageChannelWithPromotionRepos(ctx, userID, channelOrgID, s.memberRepo, s.organizationRepo)
	if err != nil {
		return nil, err
	}
	if !canManage {
		myMembers, err := s.memberRepo.ListByUserID(ctx, userID)
		if err != nil {
			return nil, err
		}
		rootMemberIDs := make([]int64, 0, len(myMembers))
		for _, my := range myMembers {
			if my.ChannelOrgID != channelOrgID || !strings.EqualFold(strings.TrimSpace(my.Status), "active") {
				continue
			}
			rootMemberIDs = append(rootMemberIDs, my.MemberID)
		}
		memberCache := make(map[int64]*DistributionMemberView, len(rootMemberIDs)+1)
		ok, err := s.distributionMemberIsSelfOrDescendant(ctx, channelOrgID, member.MemberID, rootMemberIDs, memberCache)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, ErrInvalidDistributionPromotionLink
		}
	}
	normalized.ChannelOrgID = channelOrgID
	return s.repo.Create(ctx, normalized)
}

func normalizeDistributionPromotionLinkInput(input DistributionPromotionLinkInput) (DistributionPromotionLinkInput, error) {
	input.Code = strings.ToUpper(strings.TrimSpace(input.Code))
	input.TargetType = strings.ToLower(strings.TrimSpace(input.TargetType))
	input.Status = strings.ToLower(strings.TrimSpace(input.Status))
	if input.TargetType == "" {
		input.TargetType = "registration"
	}
	if input.Status == "" {
		input.Status = "active"
	}
	if input.MemberID <= 0 || !isDistributionPromotionLinkTargetType(input.TargetType) || !isDistributionPromotionLinkStatus(input.Status) {
		return DistributionPromotionLinkInput{}, ErrInvalidDistributionPromotionLink
	}
	if input.Code == "" {
		code, err := generateDistributionPromotionCode()
		if err != nil {
			return DistributionPromotionLinkInput{}, infraerrors.InternalServer("DISTRIBUTION_PROMOTION_CODE_GEN_FAILED", "failed to generate distribution promotion code").WithCause(err)
		}
		input.Code = code
	}
	return input, nil
}

func generateDistributionPromotionCode() (string, error) {
	raw := make([]byte, 10)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return strings.ToUpper(base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(raw)), nil
}

func isDistributionPromotionLinkTargetType(targetType string) bool {
	switch strings.ToLower(strings.TrimSpace(targetType)) {
	case "registration", "oauth", "manual":
		return true
	default:
		return false
	}
}

func isDistributionPromotionLinkStatus(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "active", "inactive", "disabled":
		return true
	default:
		return false
	}
}

func distributionCanManageChannelWithPromotionRepos(
	ctx context.Context,
	userID int64,
	channelOrgID int64,
	memberRepo DistributionPromotionMemberRepository,
	organizationRepo DistributionUserChannelOrganizationRepository,
) (bool, error) {
	if organizationRepo != nil {
		org, err := organizationRepo.GetByOwnerUserID(ctx, userID)
		if err != nil {
			return false, err
		}
		if org != nil && org.ID == channelOrgID && strings.EqualFold(strings.TrimSpace(org.Status), "active") {
			return true, nil
		}
	}
	if memberRepo == nil {
		return false, nil
	}
	members, err := memberRepo.ListByUserID(ctx, userID)
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

func (s *DistributionPromotionService) listAllLinksByChannel(ctx context.Context, channelOrgID int64) ([]DistributionPromotionLink, error) {
	const pageSize = 500
	page := 1
	out := make([]DistributionPromotionLink, 0, pageSize)
	for {
		items, meta, err := s.repo.ListByChannelOrgID(ctx, channelOrgID, pagination.PaginationParams{Page: page, PageSize: pageSize})
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
		return out[i].ID < out[j].ID
	})
	return out, nil
}

func (s *DistributionPromotionService) distributionMemberIsSelfOrDescendant(
	ctx context.Context,
	channelOrgID int64,
	memberID int64,
	rootMemberIDs []int64,
	cache map[int64]*DistributionMemberView,
) (bool, error) {
	if memberID <= 0 {
		return false, nil
	}
	rootSet := make(map[int64]struct{}, len(rootMemberIDs))
	for _, rootID := range rootMemberIDs {
		rootSet[rootID] = struct{}{}
	}
	currentID := memberID
	for currentID > 0 {
		if _, ok := rootSet[currentID]; ok {
			return true, nil
		}
		member, ok := cache[currentID]
		if !ok {
			resolved, err := s.memberRepo.GetByID(ctx, currentID)
			if err != nil {
				if errors.Is(err, ErrDistributionMemberNotFound) {
					return false, nil
				}
				return false, err
			}
			cache[currentID] = resolved
			member = resolved
		}
		if member == nil || member.ChannelOrgID != channelOrgID || !strings.EqualFold(strings.TrimSpace(member.Status), "active") || member.ParentMemberID == nil {
			return false, nil
		}
		currentID = *member.ParentMemberID
	}
	return false, nil
}

func distributionPromotionPaginationResult(total int64, params pagination.PaginationParams) *pagination.PaginationResult {
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

func paginateDistributionPromotionLinks(items []DistributionPromotionLink, params pagination.PaginationParams) []DistributionPromotionLink {
	offset := params.Offset()
	if offset >= len(items) {
		return []DistributionPromotionLink{}
	}
	limit := params.Limit()
	end := offset + limit
	if end > len(items) {
		end = len(items)
	}
	return items[offset:end]
}
