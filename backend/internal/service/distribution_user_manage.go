package service

import (
	"context"
	"strings"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

type DistributionUserManageOrganizationRepository interface {
	GetByID(ctx context.Context, id int64) (*DistributionOrganization, error)
	GetByOwnerUserID(ctx context.Context, userID int64) (*DistributionOrganization, error)
	Update(ctx context.Context, id int64, input DistributionOrganizationInput) (*DistributionOrganization, error)
}

type DistributionUserManageMemberRepository interface {
	ListByUserID(ctx context.Context, userID int64) ([]DistributionMemberView, error)
}

type DistributionWholesalePricingCatalog interface {
	ListModelPricings() []ModelPricingCatalogEntry
}

type DistributionWholesalePricingItem struct {
	Model                    string  `json:"model"`
	Provider                 string  `json:"provider"`
	BillingMode              string  `json:"billing_mode"`
	OfficialInputPrice       float64 `json:"official_input_price"`
	OfficialOutputPrice      float64 `json:"official_output_price"`
	OfficialCacheWritePrice  float64 `json:"official_cache_write_price"`
	OfficialCacheReadPrice   float64 `json:"official_cache_read_price"`
	OfficialImagePrice       float64 `json:"official_image_price"`
	WholesaleInputPrice      float64 `json:"wholesale_input_price"`
	WholesaleOutputPrice     float64 `json:"wholesale_output_price"`
	WholesaleCacheWritePrice float64 `json:"wholesale_cache_write_price"`
	WholesaleCacheReadPrice  float64 `json:"wholesale_cache_read_price"`
	WholesaleImagePrice      float64 `json:"wholesale_image_price"`
}

type DistributionUserManageService struct {
	memberRepo       DistributionUserManageMemberRepository
	organizationRepo DistributionUserManageOrganizationRepository
	adminService     *DistributionAdminService
	pricingCatalog   DistributionWholesalePricingCatalog
}

func NewDistributionUserManageService(
	memberRepo DistributionUserManageMemberRepository,
	organizationRepo DistributionUserManageOrganizationRepository,
	adminService *DistributionAdminService,
	pricingCatalog DistributionWholesalePricingCatalog,
) *DistributionUserManageService {
	return &DistributionUserManageService{
		memberRepo:       memberRepo,
		organizationRepo: organizationRepo,
		adminService:     adminService,
		pricingCatalog:   pricingCatalog,
	}
}

func (s *DistributionUserManageService) ResolveChannelOrganization(ctx context.Context, userID int64) (*DistributionOrganization, error) {
	if s == nil || s.organizationRepo == nil {
		return nil, ErrInvalidDistributionOrganization
	}
	channelOrgID, err := resolveDistributionUserChannelOrgID(ctx, userID, s.memberRepo, s.organizationRepo, nil)
	if err != nil {
		return nil, err
	}
	return s.organizationRepo.GetByID(ctx, channelOrgID)
}

func (s *DistributionUserManageService) CanManageChannel(ctx context.Context, userID int64) (bool, error) {
	if s == nil || s.organizationRepo == nil {
		return false, ErrInvalidDistributionOrganization
	}
	channelOrgID, err := resolveDistributionUserChannelOrgID(ctx, userID, s.memberRepo, s.organizationRepo, nil)
	if err != nil {
		return false, err
	}
	return s.canManageChannelOrgID(ctx, userID, channelOrgID)
}

func (s *DistributionUserManageService) UpdateOrganizationForUser(ctx context.Context, userID int64, input DistributionOrganizationInput) (*DistributionOrganization, error) {
	if s == nil || s.organizationRepo == nil {
		return nil, ErrInvalidDistributionOrganization
	}
	current, err := s.ResolveChannelOrganization(ctx, userID)
	if err != nil {
		return nil, err
	}
	if current == nil {
		return nil, ErrInvalidDistributionOrganization
	}
	ok, err := s.canManageChannelOrgID(ctx, userID, current.ID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, infraerrors.Forbidden("DISTRIBUTION_CHANNEL_PERMISSION_DENIED", "distribution channel permission denied")
	}

	config := normalizeDistributionOrganizationUserConfig(current.Config, input.Config)
	brandConfig := normalizeDistributionOrganizationUserBrandConfig(current.BrandConfig, input.BrandConfig)
	name := strings.TrimSpace(input.Name)
	if name == "" {
		name = current.Name
	}

	return s.organizationRepo.Update(ctx, current.ID, DistributionOrganizationInput{
		Type:        current.Type,
		Name:        name,
		OwnerUserID: current.OwnerUserID,
		Status:      current.Status,
		Config:      config,
		BrandConfig: brandConfig,
	})
}

func (s *DistributionUserManageService) SettleCommissionForUser(ctx context.Context, userID int64, commissionID int64, input DistributionCommissionSettlementInput) (*DistributionCommissionLedger, error) {
	if s == nil || s.adminService == nil {
		return nil, ErrInvalidDistributionCommission
	}
	channelOrgID, err := resolveDistributionUserChannelOrgID(ctx, userID, s.memberRepo, s.organizationRepo, nil)
	if err != nil {
		return nil, err
	}
	ok, err := s.canManageChannelOrgID(ctx, userID, channelOrgID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, infraerrors.Forbidden("DISTRIBUTION_CHANNEL_PERMISSION_DENIED", "distribution channel permission denied")
	}
	ledger, err := s.adminService.GetCommission(ctx, commissionID)
	if err != nil {
		return nil, err
	}
	if ledger == nil || ledger.ChannelOrgID != channelOrgID {
		return nil, ErrInvalidDistributionCommission
	}
	input.SettledByUserID = &userID
	return s.adminService.SettleCommission(ctx, commissionID, input)
}

func (s *DistributionUserManageService) ListWalletRequestsForUser(
	ctx context.Context,
	userID int64,
	filter DistributionWalletRequestListFilter,
	params pagination.PaginationParams,
) ([]DistributionWalletRequest, *pagination.PaginationResult, error) {
	if s == nil || s.adminService == nil {
		return nil, nil, ErrInvalidDistributionWalletRequest
	}
	channelOrgID, err := resolveDistributionUserChannelOrgID(ctx, userID, s.memberRepo, s.organizationRepo, nil)
	if err != nil {
		return nil, nil, err
	}
	ok, err := s.canManageChannelOrgID(ctx, userID, channelOrgID)
	if err != nil {
		return nil, nil, err
	}
	if !ok {
		return nil, nil, infraerrors.Forbidden("DISTRIBUTION_CHANNEL_PERMISSION_DENIED", "distribution channel permission denied")
	}
	filter.ChannelOrgID = channelOrgID
	return s.adminService.ListWalletRequests(ctx, filter, params)
}

func (s *DistributionUserManageService) CreateWalletRequestForUser(ctx context.Context, userID int64, input DistributionWalletRequestCreateInput) (*DistributionWalletRequest, error) {
	if s == nil || s.adminService == nil {
		return nil, ErrInvalidDistributionWalletRequest
	}
	channelOrgID, err := resolveDistributionUserChannelOrgID(ctx, userID, s.memberRepo, s.organizationRepo, nil)
	if err != nil {
		return nil, err
	}
	ok, err := s.canManageChannelOrgID(ctx, userID, channelOrgID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, infraerrors.Forbidden("DISTRIBUTION_CHANNEL_PERMISSION_DENIED", "distribution channel permission denied")
	}
	input.ChannelOrgID = channelOrgID
	input.CreatedByUserID = userID
	return s.adminService.CreateWalletRequest(ctx, input)
}

func (s *DistributionUserManageService) ListAlertEventsForUser(
	ctx context.Context,
	userID int64,
	filter DistributionAlertEventListFilter,
	params pagination.PaginationParams,
) ([]DistributionAlertEvent, *pagination.PaginationResult, error) {
	if s == nil || s.adminService == nil {
		return nil, nil, ErrInvalidDistributionAlertEvent
	}
	channelOrgID, err := resolveDistributionUserChannelOrgID(ctx, userID, s.memberRepo, s.organizationRepo, nil)
	if err != nil {
		return nil, nil, err
	}
	ok, err := s.canManageChannelOrgID(ctx, userID, channelOrgID)
	if err != nil {
		return nil, nil, err
	}
	if !ok {
		return nil, nil, infraerrors.Forbidden("DISTRIBUTION_CHANNEL_PERMISSION_DENIED", "distribution channel permission denied")
	}
	filter.ChannelOrgID = channelOrgID
	return s.adminService.ListAlertEvents(ctx, filter, params)
}

func (s *DistributionUserManageService) ListWholesalePricingForUser(
	ctx context.Context,
	userID int64,
	params pagination.PaginationParams,
	keyword string,
) ([]DistributionWholesalePricingItem, *pagination.PaginationResult, float64, error) {
	if s == nil || s.organizationRepo == nil || s.pricingCatalog == nil {
		return nil, nil, 0, infraerrors.ServiceUnavailable("DISTRIBUTION_WHOLESALE_PRICING_NOT_READY", "distribution wholesale pricing is not ready")
	}

	current, err := s.ResolveChannelOrganization(ctx, userID)
	if err != nil {
		return nil, nil, 0, err
	}
	if current == nil {
		return nil, nil, 0, ErrInvalidDistributionOrganization
	}

	ok, err := s.canManageChannelOrgID(ctx, userID, current.ID)
	if err != nil {
		return nil, nil, 0, err
	}
	if !ok {
		return nil, nil, 0, infraerrors.Forbidden("DISTRIBUTION_CHANNEL_PERMISSION_DENIED", "distribution channel permission denied")
	}

	discountRate := distributionWholesaleDiscountRate(current.Config)
	keyword = strings.ToLower(strings.TrimSpace(keyword))

	allItems := s.pricingCatalog.ListModelPricings()
	filtered := make([]DistributionWholesalePricingItem, 0, len(allItems))
	for _, entry := range allItems {
		item := buildDistributionWholesalePricingItem(entry, discountRate)
		if !hasDistributionWholesalePricing(item) {
			continue
		}
		if keyword != "" &&
			!strings.Contains(strings.ToLower(item.Model), keyword) &&
			!strings.Contains(strings.ToLower(item.Provider), keyword) {
			continue
		}
		filtered = append(filtered, item)
	}

	page := params.Page
	if page < 1 {
		page = 1
	}
	pageSize := params.Limit()
	offset := params.Offset()
	total := len(filtered)
	pages := 0
	if total > 0 {
		pages = (total + pageSize - 1) / pageSize
	}

	if offset > total {
		offset = total
	}
	end := offset + pageSize
	if end > total {
		end = total
	}

	return filtered[offset:end], &pagination.PaginationResult{
		Total:    int64(total),
		Page:     page,
		PageSize: pageSize,
		Pages:    pages,
	}, discountRate, nil
}

func (s *DistributionUserManageService) canManageChannelOrgID(ctx context.Context, userID int64, channelOrgID int64) (bool, error) {
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

func normalizeDistributionOrganizationUserConfig(current map[string]any, incoming map[string]any) map[string]any {
	out := make(map[string]any, len(current)+2)
	for key, value := range current {
		out[key] = value
	}
	if incoming == nil {
		return normalizeDistributionOrganizationConfig(out)
	}
	if value, ok := incoming["commission_settlement_method"]; ok {
		out["commission_settlement_method"] = value
	}
	if value, ok := incoming["distribution_levels"]; ok {
		out["distribution_levels"] = value
	}
	return normalizeDistributionOrganizationConfig(out)
}

func normalizeDistributionOrganizationUserBrandConfig(current map[string]any, incoming map[string]any) map[string]any {
	out := make(map[string]any, len(current)+6)
	for key, value := range current {
		out[key] = value
	}
	if incoming == nil {
		return out
	}
	for _, key := range []string{"logo_url", "primary_color", "secondary_color", "theme_color", "domain", "api_domain"} {
		if value, ok := incoming[key]; ok {
			out[key] = value
		}
	}
	return out
}

func distributionWholesaleDiscountRate(config map[string]any) float64 {
	rate := distributionOrganizationConfigFloat(config, "wholesale_discount_rate", "wholesale_discount", "pricing_discount_rate")
	if rate <= 0 {
		return 0.5
	}
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

func buildDistributionWholesalePricingItem(entry ModelPricingCatalogEntry, discountRate float64) DistributionWholesalePricingItem {
	pricing := entry.Pricing
	return DistributionWholesalePricingItem{
		Model:                    entry.Model,
		Provider:                 pricing.LiteLLMProvider,
		BillingMode:              pricing.Mode,
		OfficialInputPrice:       pricing.InputCostPerToken,
		OfficialOutputPrice:      pricing.OutputCostPerToken,
		OfficialCacheWritePrice:  pricing.CacheCreationInputTokenCost,
		OfficialCacheReadPrice:   pricing.CacheReadInputTokenCost,
		OfficialImagePrice:       firstPositiveFloat(pricing.OutputCostPerImage, pricing.OutputCostPerImageToken),
		WholesaleInputPrice:      pricing.InputCostPerToken * discountRate,
		WholesaleOutputPrice:     pricing.OutputCostPerToken * discountRate,
		WholesaleCacheWritePrice: pricing.CacheCreationInputTokenCost * discountRate,
		WholesaleCacheReadPrice:  pricing.CacheReadInputTokenCost * discountRate,
		WholesaleImagePrice:      firstPositiveFloat(pricing.OutputCostPerImage, pricing.OutputCostPerImageToken) * discountRate,
	}
}

func hasDistributionWholesalePricing(item DistributionWholesalePricingItem) bool {
	return item.OfficialInputPrice > 0 ||
		item.OfficialOutputPrice > 0 ||
		item.OfficialCacheWritePrice > 0 ||
		item.OfficialCacheReadPrice > 0 ||
		item.OfficialImagePrice > 0
}

func firstPositiveFloat(values ...float64) float64 {
	for _, value := range values {
		if value > 0 {
			return value
		}
	}
	return 0
}
