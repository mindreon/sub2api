package service

import (
	"context"
	"errors"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

var (
	ErrDistributionAttributionNotFound   = infraerrors.NotFound("DISTRIBUTION_ATTRIBUTION_NOT_FOUND", "distribution attribution not found")
	ErrInvalidDistributionAttribution    = infraerrors.BadRequest("INVALID_DISTRIBUTION_ATTRIBUTION", "invalid distribution attribution")
	ErrDistributionPromotionLinkNotFound = infraerrors.NotFound("DISTRIBUTION_PROMOTION_LINK_NOT_FOUND", "distribution promotion link not found")
	ErrInvalidDistributionPromotionLink  = infraerrors.BadRequest("INVALID_DISTRIBUTION_PROMOTION_LINK", "invalid distribution promotion link")
)

type DistributionAttribution struct {
	UserID           int64     `json:"user_id"`
	ChannelOrgID     int64     `json:"channel_org_id"`
	ReferrerMemberID *int64    `json:"referrer_member_id,omitempty"`
	PromotionLinkID  *int64    `json:"promotion_link_id,omitempty"`
	BoundAt          time.Time `json:"bound_at"`
	BoundSource      string    `json:"bound_source"`
	BoundBy          string    `json:"bound_by"`
	AuditID          *int64    `json:"audit_id,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type DistributionAttributionAuditView struct {
	ID                       int64     `json:"id"`
	UserID                   int64     `json:"user_id"`
	UserEmail                string    `json:"user_email"`
	Username                 string    `json:"username"`
	PreviousChannelOrgID     *int64    `json:"previous_channel_org_id,omitempty"`
	PreviousReferrerMemberID *int64    `json:"previous_referrer_member_id,omitempty"`
	PreviousPromotionLinkID  *int64    `json:"previous_promotion_link_id,omitempty"`
	PreviousBoundSource      string    `json:"previous_bound_source"`
	PreviousBoundBy          string    `json:"previous_bound_by"`
	NewChannelOrgID          int64     `json:"new_channel_org_id"`
	NewReferrerMemberID      *int64    `json:"new_referrer_member_id,omitempty"`
	NewPromotionLinkID       *int64    `json:"new_promotion_link_id,omitempty"`
	NewBoundSource           string    `json:"new_bound_source"`
	NewBoundBy               string    `json:"new_bound_by"`
	Note                     string    `json:"note"`
	OperatorUserID           *int64    `json:"operator_user_id,omitempty"`
	OperatorUserEmail        string    `json:"operator_user_email"`
	OperatorUsername         string    `json:"operator_username"`
	CreatedAt                time.Time `json:"created_at"`
}

type DistributionAttributionInput struct {
	UserID           int64
	ChannelOrgID     int64
	ReferrerMemberID *int64
	PromotionLinkID  *int64
	BoundAt          time.Time
	BoundSource      string
	BoundBy          string
	AuditID          *int64
}

type DistributionAttributionAdminUpdateInput struct {
	UserID           int64
	ChannelOrgID     int64
	ReferrerMemberID *int64
	PromotionLinkID  *int64
	BoundSource      string
	BoundBy          string
	OperatorUserID   *int64
	Note             string
}

type DistributionPromotionLink struct {
	ID           int64     `json:"id"`
	ChannelOrgID int64     `json:"channel_org_id"`
	MemberID     int64     `json:"member_id"`
	UserID       int64     `json:"user_id"`
	UserEmail    string    `json:"user_email"`
	Username     string    `json:"username"`
	RoleType     string    `json:"role_type"`
	Code         string    `json:"code"`
	TargetType   string    `json:"target_type"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type DistributionPromotionRepository interface {
	GetByCode(ctx context.Context, code string) (*DistributionPromotionLink, error)
}

type DistributionAttributionRepository interface {
	GetByUserID(ctx context.Context, userID int64) (*DistributionAttribution, error)
	Create(ctx context.Context, input DistributionAttributionInput) (*DistributionAttribution, error)
}

type DistributionAttributionService struct {
	repo          DistributionAttributionRepository
	promotionRepo DistributionPromotionRepository
}

func NewDistributionAttributionService(repo DistributionAttributionRepository) *DistributionAttributionService {
	return &DistributionAttributionService{repo: repo}
}

func (s *DistributionAttributionService) SetPromotionRepository(repo DistributionPromotionRepository) {
	if s == nil {
		return
	}
	s.promotionRepo = repo
}

func (s *DistributionAttributionService) EnsureUserAttribution(ctx context.Context, input DistributionAttributionInput) (*DistributionAttribution, error) {
	if s == nil || s.repo == nil {
		return nil, ErrInvalidDistributionAttribution
	}
	if input.UserID <= 0 || input.ChannelOrgID <= 0 {
		return nil, ErrInvalidDistributionAttribution
	}

	existing, err := s.repo.GetByUserID(ctx, input.UserID)
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, ErrDistributionAttributionNotFound) {
		return nil, err
	}

	if input.BoundAt.IsZero() {
		input.BoundAt = time.Now().UTC()
	}
	if strings.TrimSpace(input.BoundSource) == "" {
		input.BoundSource = "registration"
	}
	if strings.TrimSpace(input.BoundBy) == "" {
		input.BoundBy = "system"
	}

	return s.repo.Create(ctx, input)
}

func (s *DistributionAttributionService) EnsureUserAttributionFromPromotionCode(
	ctx context.Context,
	userID int64,
	promotionCode string,
	boundSource string,
	boundBy string,
) (*DistributionAttribution, error) {
	if strings.TrimSpace(promotionCode) == "" {
		return nil, nil
	}
	if s == nil || s.repo == nil {
		return nil, ErrInvalidDistributionAttribution
	}
	if s.promotionRepo == nil {
		return nil, nil
	}

	link, err := s.promotionRepo.GetByCode(ctx, strings.TrimSpace(promotionCode))
	if err != nil {
		if errors.Is(err, ErrDistributionPromotionLinkNotFound) {
			return nil, nil
		}
		return nil, err
	}
	if link == nil || link.ChannelOrgID <= 0 || link.MemberID <= 0 {
		return nil, ErrInvalidDistributionPromotionLink
	}
	if !strings.EqualFold(strings.TrimSpace(link.Status), "active") {
		return nil, nil
	}

	memberID := link.MemberID
	linkID := link.ID
	return s.EnsureUserAttribution(ctx, DistributionAttributionInput{
		UserID:           userID,
		ChannelOrgID:     link.ChannelOrgID,
		ReferrerMemberID: &memberID,
		PromotionLinkID:  &linkID,
		BoundSource:      boundSource,
		BoundBy:          boundBy,
	})
}
