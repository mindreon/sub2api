package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

type distributionMemberCreateRepoStub struct {
	membersByUser map[int64][]DistributionMemberView
	membersByID   map[int64]*DistributionMemberView
	createdInput  *DistributionMemberInput
	nextID        int64
}

func (s *distributionMemberCreateRepoStub) ListByUserID(ctx context.Context, userID int64) ([]DistributionMemberView, error) {
	return s.membersByUser[userID], nil
}

func (s *distributionMemberCreateRepoStub) GetByID(ctx context.Context, memberID int64) (*DistributionMemberView, error) {
	member := s.membersByID[memberID]
	if member == nil {
		return nil, ErrDistributionMemberNotFound
	}
	return member, nil
}

func (s *distributionMemberCreateRepoStub) CountByChannelOrgIDAndRole(ctx context.Context, channelOrgID int64, roleType string) (int64, error) {
	var total int64
	for _, members := range s.membersByUser {
		for _, member := range members {
			if member.ChannelOrgID == channelOrgID && member.RoleType == roleType {
				total++
			}
		}
	}
	for _, member := range s.membersByID {
		if member != nil && member.ChannelOrgID == channelOrgID && member.RoleType == roleType {
			total++
		}
	}
	return total, nil
}

func (s *distributionMemberCreateRepoStub) Create(ctx context.Context, input DistributionMemberInput) (*DistributionMemberView, error) {
	s.createdInput = &input
	now := time.Now().UTC()
	if s.nextID == 0 {
		s.nextID = 100
	}
	return &DistributionMemberView{
		MemberID:       s.nextID,
		UserID:         input.UserID,
		ChannelOrgID:   input.ChannelOrgID,
		RoleType:       input.RoleType,
		ParentMemberID: input.ParentMemberID,
		LevelCode:      input.LevelCode,
		CommissionRate: input.CommissionRate,
		Status:         input.Status,
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}

type distributionMemberOrgRepoStub struct {
	orgs map[int64]*DistributionOrganization
}

func (s *distributionMemberOrgRepoStub) GetByID(ctx context.Context, id int64) (*DistributionOrganization, error) {
	if org, ok := s.orgs[id]; ok {
		return org, nil
	}
	return nil, ErrInvalidDistributionOrganization
}

func (s *distributionMemberOrgRepoStub) GetByOwnerUserID(ctx context.Context, userID int64) (*DistributionOrganization, error) {
	for _, org := range s.orgs {
		if org != nil && org.OwnerUserID != nil && *org.OwnerUserID == userID {
			return org, nil
		}
	}
	return nil, nil
}

func TestDistributionMemberServiceCreateMemberRejectsCrossChannelUser(t *testing.T) {
	repo := &distributionMemberCreateRepoStub{
		membersByUser: map[int64][]DistributionMemberView{
			7: {{MemberID: 10, UserID: 7, ChannelOrgID: 99, RoleType: "agent", Status: "active"}},
		},
	}
	svc := NewDistributionMemberService(repo)

	_, err := svc.CreateMember(context.Background(), DistributionMemberInput{
		ChannelOrgID:   100,
		UserID:         7,
		RoleType:       "kol1",
		CommissionRate: 0.10,
		ParentMemberID: ptrInt64(20),
	})

	require.ErrorIs(t, err, ErrDistributionMemberChannelConflict)
	require.Nil(t, repo.createdInput)
}

func TestDistributionMemberServiceCreateMemberRejectsForeignParent(t *testing.T) {
	repo := &distributionMemberCreateRepoStub{
		membersByUser: map[int64][]DistributionMemberView{},
		membersByID: map[int64]*DistributionMemberView{
			20: {MemberID: 20, UserID: 2, ChannelOrgID: 99, RoleType: "agent", Status: "active"},
		},
	}
	svc := NewDistributionMemberService(repo)

	_, err := svc.CreateMember(context.Background(), DistributionMemberInput{
		ChannelOrgID:   100,
		UserID:         7,
		RoleType:       "kol1",
		CommissionRate: 0.10,
		ParentMemberID: ptrInt64(20),
	})

	require.ErrorIs(t, err, ErrDistributionMemberParentForbidden)
	require.Nil(t, repo.createdInput)
}

func TestDistributionMemberServiceCreateMemberRejectsAgentUnderAgent(t *testing.T) {
	repo := &distributionMemberCreateRepoStub{
		membersByUser: map[int64][]DistributionMemberView{},
		membersByID: map[int64]*DistributionMemberView{
			20: {MemberID: 20, UserID: 2, ChannelOrgID: 100, RoleType: "agent", Status: "active"},
		},
	}
	svc := NewDistributionMemberService(repo)

	_, err := svc.CreateMember(context.Background(), DistributionMemberInput{
		ChannelOrgID:   100,
		UserID:         7,
		RoleType:       "agent",
		CommissionRate: 0.10,
		ParentMemberID: ptrInt64(20),
	})

	require.ErrorIs(t, err, ErrDistributionMemberParentForbidden)
	require.Nil(t, repo.createdInput)
}

func TestDistributionMemberServiceCreateMemberAcceptsKol1UnderAgent(t *testing.T) {
	repo := &distributionMemberCreateRepoStub{
		membersByUser: map[int64][]DistributionMemberView{},
		membersByID: map[int64]*DistributionMemberView{
			20: {MemberID: 20, UserID: 2, ChannelOrgID: 100, RoleType: "agent", LevelCode: "agent", Status: "active"},
		},
	}
	svc := NewDistributionMemberService(repo)

	out, err := svc.CreateMember(context.Background(), DistributionMemberInput{
		ChannelOrgID:   100,
		UserID:         7,
		RoleType:       "kol1",
		CommissionRate: 0.10,
		ParentMemberID: ptrInt64(20),
	})

	require.NoError(t, err)
	require.Equal(t, int64(100), out.MemberID)
	require.NotNil(t, repo.createdInput)
	require.Equal(t, "agent/kol1", repo.createdInput.LevelCode)
	require.Equal(t, "active", repo.createdInput.Status)
}

func TestDistributionMemberServiceCreateMemberWrapsRepositoryConflict(t *testing.T) {
	repo := &distributionMemberCreateRepoStub{
		membersByUser: map[int64][]DistributionMemberView{},
		membersByID: map[int64]*DistributionMemberView{
			20: {MemberID: 20, UserID: 2, ChannelOrgID: 100, RoleType: "agent", Status: "active"},
		},
	}
	repoErr := errors.New("duplicate key")
	svc := NewDistributionMemberService(distributionMemberCreateRepoFunc{
		listByUserID: repo.ListByUserID,
		getByID:      repo.GetByID,
		create: func(ctx context.Context, input DistributionMemberInput) (*DistributionMemberView, error) {
			return nil, repoErr
		},
	})

	_, err := svc.CreateMember(context.Background(), DistributionMemberInput{
		ChannelOrgID:   100,
		UserID:         7,
		RoleType:       "kol1",
		CommissionRate: 0.10,
		ParentMemberID: ptrInt64(20),
	})

	require.ErrorIs(t, err, repoErr)
}

func TestDistributionMemberServiceCreateMemberUsesChannelLevelRate(t *testing.T) {
	repo := &distributionMemberCreateRepoStub{membersByUser: map[int64][]DistributionMemberView{}}
	svc := NewDistributionMemberService(repo)
	svc.SetOrganizationRepository(&distributionMemberOrgRepoStub{
		orgs: map[int64]*DistributionOrganization{
			100: {
				ID:   100,
				Type: "reseller",
				Name: "Channel A",
				Config: map[string]any{
					"distribution_levels": []map[string]any{
						{
							"code":            "vip",
							"name":            "VIP",
							"commission_rate": 18,
							"active":          true,
						},
					},
				},
			},
		},
	})
	svc.SetSettingService(NewSettingService(&distributionLevelSettingRepoStub{
		values: map[string]string{
			SettingKeyDistributionGlobalLevels: `[{"code":"VIP","name":"VIP","commission_rate":12,"active":true}]`,
		},
	}, &config.Config{}))

	out, err := svc.CreateMember(context.Background(), DistributionMemberInput{
		ChannelOrgID:   100,
		UserID:         7,
		RoleType:       "agent",
		LevelCode:      "vip",
		CommissionRate: 0,
	})

	require.NoError(t, err)
	require.Equal(t, int64(100), out.MemberID)
	require.NotNil(t, repo.createdInput)
	require.Equal(t, "vip", repo.createdInput.LevelCode)
	require.Equal(t, 0.18, repo.createdInput.CommissionRate)
}

func TestDistributionMemberServiceCreateMemberFallsBackToGlobalLevelRate(t *testing.T) {
	repo := &distributionMemberCreateRepoStub{membersByUser: map[int64][]DistributionMemberView{}}
	svc := NewDistributionMemberService(repo)
	svc.SetSettingService(NewSettingService(&distributionLevelSettingRepoStub{
		values: map[string]string{
			SettingKeyDistributionGlobalLevels: `[{"code":"gold","name":"Gold","commission_rate":9.5,"active":true}]`,
		},
	}, &config.Config{}))

	_, err := svc.CreateMember(context.Background(), DistributionMemberInput{
		ChannelOrgID:   100,
		UserID:         7,
		RoleType:       "agent",
		LevelCode:      "gold",
		CommissionRate: 0,
	})

	require.NoError(t, err)
	require.NotNil(t, repo.createdInput)
	require.Equal(t, 0.095, repo.createdInput.CommissionRate)
}

func TestDistributionMemberServiceCreateMemberRejectsWhenRoleLimitExceeded(t *testing.T) {
	repo := &distributionMemberCreateRepoStub{
		membersByUser: map[int64][]DistributionMemberView{
			9: {
				{MemberID: 20, UserID: 9, ChannelOrgID: 100, RoleType: "agent", Status: "active"},
			},
		},
	}
	svc := NewDistributionMemberService(repo)
	svc.SetOrganizationRepository(&distributionMemberOrgRepoStub{
		orgs: map[int64]*DistributionOrganization{
			100: {
				ID:     100,
				Type:   "reseller",
				Name:   "Channel A",
				Status: "active",
				Config: map[string]any{"max_agent_count": 1},
			},
		},
	})

	_, err := svc.CreateMember(context.Background(), DistributionMemberInput{
		ChannelOrgID:   100,
		UserID:         7,
		RoleType:       "agent",
		CommissionRate: 0.1,
	})

	require.ErrorIs(t, err, ErrDistributionMemberLimitExceeded)
	require.Nil(t, repo.createdInput)
}

func TestDistributionMemberServiceCreateMemberForUserAllowsOwnerToCreateAgent(t *testing.T) {
	repo := &distributionMemberCreateRepoStub{membersByUser: map[int64][]DistributionMemberView{}}
	svc := NewDistributionMemberService(repo)
	svc.SetOrganizationRepository(&distributionMemberOrgRepoStub{
		orgs: map[int64]*DistributionOrganization{
			100: {ID: 100, Type: "reseller", Name: "Channel A", OwnerUserID: ptrInt64(9), Status: "active"},
		},
	})

	out, err := svc.CreateMemberForUser(context.Background(), 9, DistributionMemberInput{
		UserID:         7,
		RoleType:       "agent",
		CommissionRate: 0.12,
	})

	require.NoError(t, err)
	require.Equal(t, int64(100), out.ChannelOrgID)
	require.NotNil(t, repo.createdInput)
	require.Equal(t, int64(100), repo.createdInput.ChannelOrgID)
	require.Nil(t, repo.createdInput.ParentMemberID)
}

func TestDistributionMemberServiceCreateMemberForUserAllowsAgentToCreateKol1UnderSelf(t *testing.T) {
	repo := &distributionMemberCreateRepoStub{
		membersByUser: map[int64][]DistributionMemberView{
			9: {{MemberID: 20, UserID: 9, ChannelOrgID: 100, RoleType: "agent", LevelCode: "agent", Status: "active"}},
		},
		membersByID: map[int64]*DistributionMemberView{
			20: {MemberID: 20, UserID: 9, ChannelOrgID: 100, RoleType: "agent", LevelCode: "agent", Status: "active"},
		},
	}
	svc := NewDistributionMemberService(repo)

	out, err := svc.CreateMemberForUser(context.Background(), 9, DistributionMemberInput{
		UserID:         7,
		RoleType:       "kol1",
		ParentMemberID: ptrInt64(20),
		CommissionRate: 0.12,
	})

	require.NoError(t, err)
	require.Equal(t, int64(100), out.ChannelOrgID)
	require.NotNil(t, repo.createdInput)
	require.Equal(t, "agent/kol1", repo.createdInput.LevelCode)
}

func TestDistributionMemberServiceCreateMemberForUserRejectsCreatingAgentWithoutOwnerOrManager(t *testing.T) {
	repo := &distributionMemberCreateRepoStub{
		membersByUser: map[int64][]DistributionMemberView{
			9: {{MemberID: 20, UserID: 9, ChannelOrgID: 100, RoleType: "agent", Status: "active"}},
		},
	}
	svc := NewDistributionMemberService(repo)

	_, err := svc.CreateMemberForUser(context.Background(), 9, DistributionMemberInput{
		UserID:         7,
		RoleType:       "agent",
		CommissionRate: 0.12,
	})

	require.ErrorIs(t, err, ErrDistributionMemberPermissionDenied)
	require.Nil(t, repo.createdInput)
}

func TestDistributionMemberServiceCreateMemberForUserAllowsManagerToCreateKolUnderChannelParent(t *testing.T) {
	repo := &distributionMemberCreateRepoStub{
		membersByUser: map[int64][]DistributionMemberView{
			9: {
				{MemberID: 90, UserID: 9, ChannelOrgID: 100, RoleType: "manager", Status: "active"},
			},
		},
		membersByID: map[int64]*DistributionMemberView{
			20: {MemberID: 20, UserID: 2, ChannelOrgID: 100, RoleType: "agent", LevelCode: "agent", Status: "active"},
		},
	}
	svc := NewDistributionMemberService(repo)

	out, err := svc.CreateMemberForUser(context.Background(), 9, DistributionMemberInput{
		UserID:         7,
		RoleType:       "kol1",
		ParentMemberID: ptrInt64(20),
		CommissionRate: 0.12,
	})

	require.NoError(t, err)
	require.Equal(t, int64(100), out.ChannelOrgID)
	require.NotNil(t, repo.createdInput)
	require.Equal(t, "agent/kol1", repo.createdInput.LevelCode)
}

type distributionMemberCreateRepoFunc struct {
	listByUserID func(ctx context.Context, userID int64) ([]DistributionMemberView, error)
	getByID      func(ctx context.Context, memberID int64) (*DistributionMemberView, error)
	countByRole  func(ctx context.Context, channelOrgID int64, roleType string) (int64, error)
	create       func(ctx context.Context, input DistributionMemberInput) (*DistributionMemberView, error)
}

func (f distributionMemberCreateRepoFunc) ListByUserID(ctx context.Context, userID int64) ([]DistributionMemberView, error) {
	return f.listByUserID(ctx, userID)
}

func (f distributionMemberCreateRepoFunc) GetByID(ctx context.Context, memberID int64) (*DistributionMemberView, error) {
	return f.getByID(ctx, memberID)
}

func (f distributionMemberCreateRepoFunc) CountByChannelOrgIDAndRole(ctx context.Context, channelOrgID int64, roleType string) (int64, error) {
	return f.countByRole(ctx, channelOrgID, roleType)
}

func (f distributionMemberCreateRepoFunc) Create(ctx context.Context, input DistributionMemberInput) (*DistributionMemberView, error) {
	return f.create(ctx, input)
}

func ptrInt64(v int64) *int64 {
	return &v
}
