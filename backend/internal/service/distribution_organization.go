package service

import (
	"context"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

var (
	ErrInvalidDistributionOrganization = infraerrors.BadRequest("INVALID_DISTRIBUTION_ORGANIZATION", "invalid distribution organization")
)

type DistributionOrganization struct {
	ID          int64          `json:"id"`
	Type        string         `json:"type"`
	Name        string         `json:"name"`
	OwnerUserID *int64         `json:"owner_user_id,omitempty"`
	Status      string         `json:"status"`
	Config      map[string]any `json:"config"`
	BrandConfig map[string]any `json:"brand_config"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

type DistributionOrganizationInput struct {
	Type        string
	Name        string
	OwnerUserID *int64
	Status      string
	Config      map[string]any
	BrandConfig map[string]any
}

type DistributionOrganizationRepository interface {
	Create(ctx context.Context, input DistributionOrganizationInput) (*DistributionOrganization, error)
	GetByID(ctx context.Context, id int64) (*DistributionOrganization, error)
	GetByOwnerUserID(ctx context.Context, userID int64) (*DistributionOrganization, error)
	Update(ctx context.Context, id int64, input DistributionOrganizationInput) (*DistributionOrganization, error)
}

type DistributionOrganizationService struct {
	repo DistributionOrganizationRepository
}

func NewDistributionOrganizationService(repo DistributionOrganizationRepository) *DistributionOrganizationService {
	return &DistributionOrganizationService{repo: repo}
}

func (s *DistributionOrganizationService) CreateOrganization(ctx context.Context, input DistributionOrganizationInput) (*DistributionOrganization, error) {
	if s == nil || s.repo == nil {
		return nil, ErrInvalidDistributionOrganization
	}
	input.Type = strings.ToLower(strings.TrimSpace(input.Type))
	input.Name = strings.TrimSpace(input.Name)
	input.Status = strings.ToLower(strings.TrimSpace(input.Status))
	if input.Status == "" {
		input.Status = "active"
	}
	input.Config = normalizeDistributionOrganizationConfig(input.Config)
	if input.BrandConfig == nil {
		input.BrandConfig = map[string]any{}
	}

	if input.Name == "" || !isDistributionOrganizationType(input.Type) || !isDistributionOrganizationStatus(input.Status) {
		return nil, ErrInvalidDistributionOrganization
	}
	if input.OwnerUserID != nil && *input.OwnerUserID <= 0 {
		return nil, ErrInvalidDistributionOrganization
	}

	return s.repo.Create(ctx, input)
}

func (s *DistributionOrganizationService) UpdateOrganization(ctx context.Context, id int64, input DistributionOrganizationInput) (*DistributionOrganization, error) {
	if s == nil || s.repo == nil {
		return nil, ErrInvalidDistributionOrganization
	}
	input.Type = strings.ToLower(strings.TrimSpace(input.Type))
	input.Name = strings.TrimSpace(input.Name)
	input.Status = strings.ToLower(strings.TrimSpace(input.Status))
	if input.Status == "" {
		input.Status = "active"
	}
	input.Config = normalizeDistributionOrganizationConfig(input.Config)
	if input.BrandConfig == nil {
		input.BrandConfig = map[string]any{}
	}
	if id <= 0 || input.Name == "" || !isDistributionOrganizationType(input.Type) || !isDistributionOrganizationStatus(input.Status) {
		return nil, ErrInvalidDistributionOrganization
	}
	if input.OwnerUserID != nil && *input.OwnerUserID <= 0 {
		return nil, ErrInvalidDistributionOrganization
	}
	return s.repo.Update(ctx, id, input)
}

func normalizeDistributionOrganizationConfig(config map[string]any) map[string]any {
	if config == nil {
		return map[string]any{}
	}

	normalized := make(map[string]any, len(config))
	for key, value := range config {
		normalized[key] = value
	}

	rawLevels, ok := normalized["distribution_levels"]
	if !ok {
		return normalized
	}

	levels, err := parseDistributionLevelConfigs(rawLevels)
	if err != nil {
		delete(normalized, "distribution_levels")
		return normalized
	}
	normalized["distribution_levels"] = levels
	return normalized
}

func isDistributionOrganizationType(orgType string) bool {
	switch orgType {
	case "platform", "reseller", "oem":
		return true
	default:
		return false
	}
}

func isDistributionOrganizationStatus(status string) bool {
	switch status {
	case "active", "inactive", "disabled":
		return true
	default:
		return false
	}
}
