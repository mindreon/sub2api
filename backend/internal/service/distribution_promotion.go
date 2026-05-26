package service

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"errors"
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
	organizationRepo distributionUserChannelOrganizationRepository
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

func (s *DistributionPromotionService) SetOrganizationRepository(repo distributionUserChannelOrganizationRepository) {
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
	return s.repo.ListByChannelOrgID(ctx, channelOrgID, params)
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
