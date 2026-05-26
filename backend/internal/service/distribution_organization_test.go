package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type distributionOrganizationRepoStub struct {
	createdInput *DistributionOrganizationInput
	updatedID    int64
	updatedInput *DistributionOrganizationInput
}

func (s *distributionOrganizationRepoStub) Create(ctx context.Context, input DistributionOrganizationInput) (*DistributionOrganization, error) {
	s.createdInput = &input
	now := time.Now().UTC()
	return &DistributionOrganization{
		ID:          88,
		Type:        input.Type,
		Name:        input.Name,
		OwnerUserID: input.OwnerUserID,
		Status:      input.Status,
		Config:      input.Config,
		BrandConfig: input.BrandConfig,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

func (s *distributionOrganizationRepoStub) GetByID(ctx context.Context, id int64) (*DistributionOrganization, error) {
	return &DistributionOrganization{ID: id, Type: "reseller", Name: "Independent Agent", Status: "active"}, nil
}

func (s *distributionOrganizationRepoStub) GetByOwnerUserID(ctx context.Context, userID int64) (*DistributionOrganization, error) {
	return nil, nil
}

func (s *distributionOrganizationRepoStub) Update(ctx context.Context, id int64, input DistributionOrganizationInput) (*DistributionOrganization, error) {
	s.updatedID = id
	s.updatedInput = &input
	now := time.Now().UTC()
	return &DistributionOrganization{
		ID:          id,
		Type:        input.Type,
		Name:        input.Name,
		OwnerUserID: input.OwnerUserID,
		Status:      input.Status,
		Config:      input.Config,
		BrandConfig: input.BrandConfig,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

func TestDistributionOrganizationServiceCreateOrganizationDefaultsStatusAndConfig(t *testing.T) {
	ownerID := int64(7)
	repo := &distributionOrganizationRepoStub{}
	svc := NewDistributionOrganizationService(repo)

	out, err := svc.CreateOrganization(context.Background(), DistributionOrganizationInput{
		Type:        "reseller",
		Name:        " Independent Agent ",
		OwnerUserID: &ownerID,
	})

	require.NoError(t, err)
	require.Equal(t, int64(88), out.ID)
	require.Equal(t, "Independent Agent", repo.createdInput.Name)
	require.Equal(t, "active", repo.createdInput.Status)
	require.NotNil(t, repo.createdInput.Config)
	require.NotNil(t, repo.createdInput.BrandConfig)
}

func TestDistributionOrganizationServiceCreateOrganizationRejectsInvalidType(t *testing.T) {
	repo := &distributionOrganizationRepoStub{}
	svc := NewDistributionOrganizationService(repo)

	_, err := svc.CreateOrganization(context.Background(), DistributionOrganizationInput{
		Type: "agency",
		Name: "bad",
	})

	require.ErrorIs(t, err, ErrInvalidDistributionOrganization)
	require.Nil(t, repo.createdInput)
}

func TestDistributionOrganizationServiceUpdateOrganizationNormalizesInput(t *testing.T) {
	repo := &distributionOrganizationRepoStub{}
	svc := NewDistributionOrganizationService(repo)
	ownerID := int64(9)

	out, err := svc.UpdateOrganization(context.Background(), 88, DistributionOrganizationInput{
		Type:        "reseller",
		Name:        " Agent ",
		OwnerUserID: &ownerID,
		Config:      nil,
		BrandConfig: nil,
	})

	require.NoError(t, err)
	require.Equal(t, int64(88), out.ID)
	require.Equal(t, "Agent", repo.updatedInput.Name)
	require.Equal(t, "active", repo.updatedInput.Status)
	require.NotNil(t, repo.updatedInput.Config)
	require.NotNil(t, repo.updatedInput.BrandConfig)
}

func TestDistributionOrganizationServiceCreateOrganizationNormalizesDistributionLevels(t *testing.T) {
	repo := &distributionOrganizationRepoStub{}
	svc := NewDistributionOrganizationService(repo)

	_, err := svc.CreateOrganization(context.Background(), DistributionOrganizationInput{
		Type: "reseller",
		Name: "Agent",
		Config: map[string]any{
			"distribution_levels": []map[string]any{
				{
					"code":            " vip ",
					"name":            "",
					"commission_rate": 18.5,
					"active":          true,
					"sort_order":      2,
					"note":            " default ",
				},
			},
		},
	})

	require.NoError(t, err)
	rawLevels, ok := repo.createdInput.Config["distribution_levels"]
	require.True(t, ok)

	levels, ok := rawLevels.([]DistributionLevelConfig)
	require.True(t, ok)
	require.Equal(t, []DistributionLevelConfig{
		{
			Code:           "VIP",
			Name:           "VIP",
			CommissionRate: 18.5,
			Active:         true,
			SortOrder:      2,
			Note:           "default",
		},
	}, levels)
}
