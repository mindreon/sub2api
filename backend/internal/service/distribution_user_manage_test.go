package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

type distributionUserManageOrgRepoStub struct {
	orgByID      map[int64]*DistributionOrganization
	orgByOwner   map[int64]*DistributionOrganization
	updatedID    int64
	updatedInput DistributionOrganizationInput
}

func (s *distributionUserManageOrgRepoStub) GetByID(ctx context.Context, id int64) (*DistributionOrganization, error) {
	if s.orgByID == nil {
		return nil, nil
	}
	return s.orgByID[id], nil
}

func (s *distributionUserManageOrgRepoStub) GetByOwnerUserID(ctx context.Context, userID int64) (*DistributionOrganization, error) {
	if s.orgByOwner == nil {
		return nil, nil
	}
	return s.orgByOwner[userID], nil
}

func (s *distributionUserManageOrgRepoStub) Update(ctx context.Context, id int64, input DistributionOrganizationInput) (*DistributionOrganization, error) {
	s.updatedID = id
	s.updatedInput = input
	current := s.orgByID[id]
	if current == nil {
		return nil, ErrInvalidDistributionOrganization
	}
	return &DistributionOrganization{
		ID:          current.ID,
		Type:        current.Type,
		Name:        input.Name,
		OwnerUserID: current.OwnerUserID,
		Status:      current.Status,
		Config:      input.Config,
		BrandConfig: input.BrandConfig,
	}, nil
}

type distributionUserManageMemberRepoStub struct {
	byUser map[int64][]DistributionMemberView
}

func (s *distributionUserManageMemberRepoStub) ListByUserID(ctx context.Context, userID int64) ([]DistributionMemberView, error) {
	if s.byUser == nil {
		return nil, nil
	}
	return s.byUser[userID], nil
}

type distributionWholesalePricingCatalogStub struct {
	items []ModelPricingCatalogEntry
}

func (s *distributionWholesalePricingCatalogStub) ListModelPricings() []ModelPricingCatalogEntry {
	return append([]ModelPricingCatalogEntry(nil), s.items...)
}

func TestDistributionUserManageService_UpdateOrganizationForOwner(t *testing.T) {
	ownerUserID := int64(9)
	orgRepo := &distributionUserManageOrgRepoStub{
		orgByID: map[int64]*DistributionOrganization{
			88: {
				ID:          88,
				Type:        "reseller",
				Name:        "Channel A",
				OwnerUserID: &ownerUserID,
				Status:      "active",
				Config: map[string]any{
					"commission_settlement_method": "manual",
					"consumption_limit":            1000,
				},
				BrandConfig: map[string]any{"logo_url": "old"},
			},
		},
		orgByOwner: map[int64]*DistributionOrganization{
			9: {ID: 88, Type: "reseller", Name: "Channel A", OwnerUserID: &ownerUserID, Status: "active"},
		},
	}
	memberRepo := &distributionUserManageMemberRepoStub{}
	svc := NewDistributionUserManageService(memberRepo, orgRepo, nil, nil)

	out, err := svc.UpdateOrganizationForUser(context.Background(), 9, DistributionOrganizationInput{
		Name: "Channel A+",
		Config: map[string]any{
			"commission_settlement_method": "offline",
			"distribution_levels": []map[string]any{
				{"code": "agent", "name": "Agent", "commission_rate": 0.2},
			},
			"consumption_limit": 9999,
		},
		BrandConfig: map[string]any{
			"logo_url":    "new",
			"domain":      "brand.example.com",
			"unknown_key": "ignored",
		},
	})
	require.NoError(t, err)
	require.NotNil(t, out)
	require.Equal(t, int64(88), orgRepo.updatedID)
	require.Equal(t, "Channel A+", orgRepo.updatedInput.Name)
	require.Equal(t, "offline", orgRepo.updatedInput.Config["commission_settlement_method"])
	require.Equal(t, 1000, orgRepo.updatedInput.Config["consumption_limit"])
	require.Equal(t, "new", orgRepo.updatedInput.BrandConfig["logo_url"])
	require.Equal(t, "brand.example.com", orgRepo.updatedInput.BrandConfig["domain"])
	_, exists := orgRepo.updatedInput.BrandConfig["unknown_key"]
	require.False(t, exists)
}

func TestDistributionUserManageService_SettleCommissionForManager(t *testing.T) {
	memberRepo := &distributionUserManageMemberRepoStub{
		byUser: map[int64][]DistributionMemberView{
			7: {{MemberID: 11, UserID: 7, ChannelOrgID: 88, RoleType: "manager", Status: "active"}},
		},
	}
	orgRepo := &distributionUserManageOrgRepoStub{
		orgByID: map[int64]*DistributionOrganization{
			88: {ID: 88, Type: "reseller", Name: "Channel A", Status: "active"},
		},
	}
	adminSvc := NewDistributionAdminService(nil, nil, nil, nil, &distributionAdminWalletRepoStub{
		wallet: &DistributionWallet{ChannelOrgID: 88, OrganizationType: "reseller"},
	}, nil, &distributionAdminSettlementRepoStub{
		ledgerByID: &DistributionCommissionLedger{
			ID:               1001,
			ChannelOrgID:     88,
			Amount:           12.5,
			Status:           "available",
			SettlementMethod: "manual",
		},
		settleResult: &DistributionCommissionLedger{ID: 1001, ChannelOrgID: 88, Status: "settled", SettlementMethod: "offline"},
	})
	svc := NewDistributionUserManageService(memberRepo, orgRepo, adminSvc, nil)

	out, err := svc.SettleCommissionForUser(context.Background(), 7, 1001, DistributionCommissionSettlementInput{
		SettlementMethod:      "offline",
		SettlementReferenceNo: "VCH-1",
		SettlementNote:        "paid by bank",
	})
	require.NoError(t, err)
	require.NotNil(t, out)
	require.Equal(t, "settled", out.Status)
}

func TestDistributionUserManageService_ListWholesalePricingForOwner(t *testing.T) {
	ownerUserID := int64(9)
	orgRepo := &distributionUserManageOrgRepoStub{
		orgByID: map[int64]*DistributionOrganization{
			88: {
				ID:          88,
				Type:        "reseller",
				Name:        "Channel A",
				OwnerUserID: &ownerUserID,
				Status:      "active",
				Config: map[string]any{
					"wholesale_discount_rate": 0.4,
				},
			},
		},
		orgByOwner: map[int64]*DistributionOrganization{
			9: {ID: 88, Type: "reseller", Name: "Channel A", OwnerUserID: &ownerUserID, Status: "active"},
		},
	}
	catalog := &distributionWholesalePricingCatalogStub{
		items: []ModelPricingCatalogEntry{
			{
				Model: "claude-sonnet-4",
				Pricing: LiteLLMModelPricing{
					InputCostPerToken:     3e-6,
					OutputCostPerToken:    15e-6,
					LiteLLMProvider:       "anthropic",
					Mode:                  "chat",
					SupportsPromptCaching: true,
				},
			},
			{
				Model: "free-model",
				Pricing: LiteLLMModelPricing{
					LiteLLMProvider: "openai",
					Mode:            "chat",
				},
			},
		},
	}

	svc := NewDistributionUserManageService(&distributionUserManageMemberRepoStub{}, orgRepo, nil, catalog)
	items, pageInfo, discountRate, err := svc.ListWholesalePricingForUser(context.Background(), 9, pagination.PaginationParams{Page: 1, PageSize: 20}, "claude")
	require.NoError(t, err)
	require.NotNil(t, pageInfo)
	require.Equal(t, 0.4, discountRate)
	require.Equal(t, int64(1), pageInfo.Total)
	require.Len(t, items, 1)
	require.Equal(t, "claude-sonnet-4", items[0].Model)
	require.Equal(t, "anthropic", items[0].Provider)
	require.Equal(t, "chat", items[0].BillingMode)
	require.Equal(t, 3e-6, items[0].OfficialInputPrice)
	require.InDelta(t, 1.2e-6, items[0].WholesaleInputPrice, 1e-12)
	require.Equal(t, 15e-6, items[0].OfficialOutputPrice)
	require.InDelta(t, 6e-6, items[0].WholesaleOutputPrice, 1e-12)
}
