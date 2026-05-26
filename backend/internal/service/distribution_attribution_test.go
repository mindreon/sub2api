package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type distributionAttributionRepoStub struct {
	getFn       func(ctx context.Context, userID int64) (*DistributionAttribution, error)
	createFn    func(ctx context.Context, input DistributionAttributionInput) (*DistributionAttribution, error)
	getCalls    int
	createCalls int
}

func (s *distributionAttributionRepoStub) GetByUserID(ctx context.Context, userID int64) (*DistributionAttribution, error) {
	s.getCalls++
	return s.getFn(ctx, userID)
}

func (s *distributionAttributionRepoStub) Create(ctx context.Context, input DistributionAttributionInput) (*DistributionAttribution, error) {
	s.createCalls++
	return s.createFn(ctx, input)
}

type distributionPromotionRepoStub struct {
	links    map[string]*DistributionPromotionLink
	getCalls []string
}

func (s *distributionPromotionRepoStub) GetByCode(ctx context.Context, code string) (*DistributionPromotionLink, error) {
	s.getCalls = append(s.getCalls, code)
	if link, ok := s.links[code]; ok {
		return link, nil
	}
	return nil, ErrDistributionPromotionLinkNotFound
}

func TestDistributionAttributionService_KeepsExistingAttribution(t *testing.T) {
	existingAttribution := &DistributionAttribution{
		UserID:           101,
		ChannelOrgID:     11,
		BoundSource:      "registration",
		BoundBy:          "system",
		BoundAt:          time.Unix(100, 0).UTC(),
		CreatedAt:        time.Unix(100, 0).UTC(),
		UpdatedAt:        time.Unix(100, 0).UTC(),
		PromotionLinkID:  nil,
		ReferrerMemberID: nil,
	}

	repo := &distributionAttributionRepoStub{
		getFn: func(ctx context.Context, userID int64) (*DistributionAttribution, error) {
			require.Equal(t, int64(101), userID)
			return existingAttribution, nil
		},
		createFn: func(ctx context.Context, input DistributionAttributionInput) (*DistributionAttribution, error) {
			t.Fatalf("Create should not be called when attribution already exists")
			return nil, nil
		},
	}

	svc := NewDistributionAttributionService(repo)
	out, err := svc.EnsureUserAttribution(context.Background(), DistributionAttributionInput{
		UserID:       101,
		ChannelOrgID: 11,
		BoundSource:  "oauth",
		BoundBy:      "admin",
	})
	require.NoError(t, err)
	require.Same(t, existingAttribution, out)
}

func TestDistributionAttributionService_CreatesMissingAttribution(t *testing.T) {
	var createdInput DistributionAttributionInput

	repo := &distributionAttributionRepoStub{
		getFn: func(ctx context.Context, userID int64) (*DistributionAttribution, error) {
			return nil, ErrDistributionAttributionNotFound
		},
		createFn: func(ctx context.Context, input DistributionAttributionInput) (*DistributionAttribution, error) {
			createdInput = input
			return &DistributionAttribution{
				UserID:       input.UserID,
				ChannelOrgID: input.ChannelOrgID,
				BoundSource:  input.BoundSource,
				BoundBy:      input.BoundBy,
				BoundAt:      input.BoundAt,
				CreatedAt:    input.BoundAt,
				UpdatedAt:    input.BoundAt,
			}, nil
		},
	}

	svc := NewDistributionAttributionService(repo)
	out, err := svc.EnsureUserAttribution(context.Background(), DistributionAttributionInput{
		UserID:       202,
		ChannelOrgID: 22,
		BoundSource:  "registration",
		BoundBy:      "system",
	})
	require.NoError(t, err)
	require.Equal(t, int64(202), out.UserID)
	require.Equal(t, int64(22), createdInput.ChannelOrgID)
	require.Equal(t, "registration", createdInput.BoundSource)
	require.Equal(t, "system", createdInput.BoundBy)
	require.False(t, createdInput.BoundAt.IsZero())
}

func TestDistributionAttributionService_ReturnsErrorForInvalidInput(t *testing.T) {
	svc := NewDistributionAttributionService(nil)
	_, err := svc.EnsureUserAttribution(context.Background(), DistributionAttributionInput{})
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrInvalidDistributionAttribution))
}

func TestDistributionAttributionService_KeepsFirstPromotionLink(t *testing.T) {
	var stored *DistributionAttribution
	repo := &distributionAttributionRepoStub{
		getFn: func(ctx context.Context, userID int64) (*DistributionAttribution, error) {
			if stored != nil {
				return stored, nil
			}
			return nil, ErrDistributionAttributionNotFound
		},
		createFn: func(ctx context.Context, input DistributionAttributionInput) (*DistributionAttribution, error) {
			stored = &DistributionAttribution{
				UserID:           input.UserID,
				ChannelOrgID:     input.ChannelOrgID,
				ReferrerMemberID: input.ReferrerMemberID,
				PromotionLinkID:  input.PromotionLinkID,
				BoundAt:          input.BoundAt,
				BoundSource:      input.BoundSource,
				BoundBy:          input.BoundBy,
				AuditID:          input.AuditID,
				CreatedAt:        input.BoundAt,
				UpdatedAt:        input.BoundAt,
			}
			return stored, nil
		},
	}
	promoRepo := &distributionPromotionRepoStub{
		links: map[string]*DistributionPromotionLink{
			"LINK-A": {
				ID:           11,
				ChannelOrgID: 101,
				MemberID:     1001,
				Code:         "LINK-A",
				Status:       "active",
			},
			"LINK-B": {
				ID:           22,
				ChannelOrgID: 202,
				MemberID:     2002,
				Code:         "LINK-B",
				Status:       "active",
			},
		},
	}

	svc := NewDistributionAttributionService(repo)
	svc.SetPromotionRepository(promoRepo)

	first, err := svc.EnsureUserAttributionFromPromotionCode(context.Background(), 7, "LINK-A", "registration", "system")
	require.NoError(t, err)
	require.NotNil(t, first)
	require.Equal(t, int64(101), first.ChannelOrgID)
	require.Equal(t, int64(1001), first.ReferrerMemberIDValue())
	require.Equal(t, int64(11), first.PromotionLinkIDValue())

	second, err := svc.EnsureUserAttributionFromPromotionCode(context.Background(), 7, "LINK-B", "oauth", "system")
	require.NoError(t, err)
	require.Same(t, first, second)
	require.Equal(t, int64(101), stored.ChannelOrgID)
	require.Equal(t, 1, repo.createCalls)
	require.Equal(t, []string{"LINK-A", "LINK-B"}, promoRepo.getCalls)
}

func (a *DistributionAttribution) ReferrerMemberIDValue() int64 {
	if a == nil || a.ReferrerMemberID == nil {
		return 0
	}
	return *a.ReferrerMemberID
}

func (a *DistributionAttribution) PromotionLinkIDValue() int64 {
	if a == nil || a.PromotionLinkID == nil {
		return 0
	}
	return *a.PromotionLinkID
}
