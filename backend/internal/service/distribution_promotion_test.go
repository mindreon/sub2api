package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

type distributionPromotionLinkRepoStub struct {
	listByChannelOrgIDFn func(ctx context.Context, channelOrgID int64, params pagination.PaginationParams) ([]DistributionPromotionLink, *pagination.PaginationResult, error)
	listAdminFn          func(ctx context.Context, filter DistributionAdminListFilter, params pagination.PaginationParams) ([]DistributionPromotionLink, *pagination.PaginationResult, error)
	createFn             func(ctx context.Context, input DistributionPromotionLinkInput) (*DistributionPromotionLink, error)
}

func (s *distributionPromotionLinkRepoStub) ListByChannelOrgID(ctx context.Context, channelOrgID int64, params pagination.PaginationParams) ([]DistributionPromotionLink, *pagination.PaginationResult, error) {
	return s.listByChannelOrgIDFn(ctx, channelOrgID, params)
}

func (s *distributionPromotionLinkRepoStub) ListAdmin(ctx context.Context, filter DistributionAdminListFilter, params pagination.PaginationParams) ([]DistributionPromotionLink, *pagination.PaginationResult, error) {
	return s.listAdminFn(ctx, filter, params)
}

func (s *distributionPromotionLinkRepoStub) Create(ctx context.Context, input DistributionPromotionLinkInput) (*DistributionPromotionLink, error) {
	return s.createFn(ctx, input)
}

type distributionPromotionLinkMemberRepoStub struct {
	listByUserIDFn func(ctx context.Context, userID int64) ([]DistributionMemberView, error)
	getByIDFn      func(ctx context.Context, memberID int64) (*DistributionMemberView, error)
}

func (s *distributionPromotionLinkMemberRepoStub) ListByUserID(ctx context.Context, userID int64) ([]DistributionMemberView, error) {
	if s.listByUserIDFn == nil {
		return nil, nil
	}
	return s.listByUserIDFn(ctx, userID)
}

func (s *distributionPromotionLinkMemberRepoStub) GetByID(ctx context.Context, memberID int64) (*DistributionMemberView, error) {
	return s.getByIDFn(ctx, memberID)
}

type distributionPromotionLinkAttributionRepoStub struct {
	getByUserIDFn func(ctx context.Context, userID int64) (*DistributionAttribution, error)
}

func (s *distributionPromotionLinkAttributionRepoStub) GetByUserID(ctx context.Context, userID int64) (*DistributionAttribution, error) {
	return s.getByUserIDFn(ctx, userID)
}

func TestDistributionPromotionServiceCreateLinkGeneratesCodeAndNormalizesFields(t *testing.T) {
	var created DistributionPromotionLinkInput
	svc := NewDistributionPromotionService(
		&distributionPromotionLinkRepoStub{
			createFn: func(ctx context.Context, input DistributionPromotionLinkInput) (*DistributionPromotionLink, error) {
				created = input
				now := time.Now().UTC()
				return &DistributionPromotionLink{
					ID:           11,
					ChannelOrgID: 88,
					MemberID:     input.MemberID,
					Code:         input.Code,
					TargetType:   input.TargetType,
					Status:       input.Status,
					CreatedAt:    now,
					UpdatedAt:    now,
				}, nil
			},
		},
		&distributionPromotionLinkMemberRepoStub{
			getByIDFn: func(ctx context.Context, memberID int64) (*DistributionMemberView, error) {
				return &DistributionMemberView{
					MemberID:     memberID,
					ChannelOrgID: 88,
					RoleType:     "agent",
					Status:       "active",
				}, nil
			},
			listByUserIDFn: func(ctx context.Context, userID int64) ([]DistributionMemberView, error) {
				return nil, nil
			},
		},
		&distributionPromotionLinkAttributionRepoStub{
			getByUserIDFn: func(ctx context.Context, userID int64) (*DistributionAttribution, error) {
				return &DistributionAttribution{UserID: userID, ChannelOrgID: 88}, nil
			},
		},
	)

	out, err := svc.CreateLink(context.Background(), DistributionPromotionLinkInput{
		MemberID:   11,
		TargetType: "registration",
		Status:     "active",
	})

	require.NoError(t, err)
	require.NotNil(t, out)
	require.Equal(t, int64(11), created.MemberID)
	require.Equal(t, "registration", created.TargetType)
	require.Equal(t, "active", created.Status)
	require.NotEmpty(t, created.Code)
	require.Len(t, created.Code, 16)
}

func TestDistributionPromotionServiceCreateLinkForUserRejectsForeignMember(t *testing.T) {
	svc := NewDistributionPromotionService(
		&distributionPromotionLinkRepoStub{
			createFn: func(ctx context.Context, input DistributionPromotionLinkInput) (*DistributionPromotionLink, error) {
				t.Fatalf("Create should not be called")
				return nil, nil
			},
		},
		&distributionPromotionLinkMemberRepoStub{
			listByUserIDFn: func(ctx context.Context, userID int64) ([]DistributionMemberView, error) {
				return []DistributionMemberView{
					{MemberID: 22, ChannelOrgID: 88, RoleType: "agent", Status: "active"},
				}, nil
			},
			getByIDFn: func(ctx context.Context, memberID int64) (*DistributionMemberView, error) {
				return &DistributionMemberView{MemberID: memberID, ChannelOrgID: 99, RoleType: "agent", Status: "active"}, nil
			},
		},
		&distributionPromotionLinkAttributionRepoStub{
			getByUserIDFn: func(ctx context.Context, userID int64) (*DistributionAttribution, error) {
				return &DistributionAttribution{UserID: userID, ChannelOrgID: 88}, nil
			},
		},
	)

	_, err := svc.CreateLinkForUser(context.Background(), 7, DistributionPromotionLinkInput{
		MemberID:   11,
		TargetType: "registration",
		Status:     "active",
	})

	require.Error(t, err)
	require.True(t, errors.Is(err, ErrInvalidDistributionPromotionLink))
}

func TestDistributionPromotionServiceCreateLinkForUserAllowsDescendantMember(t *testing.T) {
	var created DistributionPromotionLinkInput
	svc := NewDistributionPromotionService(
		&distributionPromotionLinkRepoStub{
			createFn: func(ctx context.Context, input DistributionPromotionLinkInput) (*DistributionPromotionLink, error) {
				created = input
				return &DistributionPromotionLink{ID: 1, ChannelOrgID: 88, MemberID: input.MemberID, Code: input.Code}, nil
			},
		},
		&distributionPromotionLinkMemberRepoStub{
			listByUserIDFn: func(ctx context.Context, userID int64) ([]DistributionMemberView, error) {
				return []DistributionMemberView{{MemberID: 20, UserID: userID, ChannelOrgID: 88, RoleType: "kol1", Status: "active"}}, nil
			},
			getByIDFn: func(ctx context.Context, memberID int64) (*DistributionMemberView, error) {
				switch memberID {
				case 21:
					return &DistributionMemberView{MemberID: 21, ChannelOrgID: 88, RoleType: "kol2", ParentMemberID: ptrInt64(20), Status: "active"}, nil
				case 20:
					return &DistributionMemberView{MemberID: 20, ChannelOrgID: 88, RoleType: "kol1", Status: "active"}, nil
				default:
					return nil, ErrDistributionMemberNotFound
				}
			},
		},
		&distributionPromotionLinkAttributionRepoStub{
			getByUserIDFn: func(ctx context.Context, userID int64) (*DistributionAttribution, error) {
				return &DistributionAttribution{UserID: userID, ChannelOrgID: 88}, nil
			},
		},
	)

	out, err := svc.CreateLinkForUser(context.Background(), 7, DistributionPromotionLinkInput{
		MemberID:   21,
		TargetType: "registration",
		Status:     "active",
	})
	require.NoError(t, err)
	require.NotNil(t, out)
	require.Equal(t, int64(21), created.MemberID)
}

func TestDistributionPromotionServiceCreateLinkForManagerAllowsAnyChannelMember(t *testing.T) {
	var created DistributionPromotionLinkInput
	svc := NewDistributionPromotionService(
		&distributionPromotionLinkRepoStub{
			createFn: func(ctx context.Context, input DistributionPromotionLinkInput) (*DistributionPromotionLink, error) {
				created = input
				return &DistributionPromotionLink{ID: 1, ChannelOrgID: 88, MemberID: input.MemberID, Code: input.Code}, nil
			},
		},
		&distributionPromotionLinkMemberRepoStub{
			listByUserIDFn: func(ctx context.Context, userID int64) ([]DistributionMemberView, error) {
				return []DistributionMemberView{
					{MemberID: 90, UserID: userID, ChannelOrgID: 88, RoleType: "manager", Status: "active"},
				}, nil
			},
			getByIDFn: func(ctx context.Context, memberID int64) (*DistributionMemberView, error) {
				return &DistributionMemberView{MemberID: memberID, ChannelOrgID: 88, RoleType: "kol1", ParentMemberID: ptrInt64(10), Status: "active"}, nil
			},
		},
		&distributionPromotionLinkAttributionRepoStub{
			getByUserIDFn: func(ctx context.Context, userID int64) (*DistributionAttribution, error) {
				return &DistributionAttribution{UserID: userID, ChannelOrgID: 88}, nil
			},
		},
	)

	out, err := svc.CreateLinkForUser(context.Background(), 7, DistributionPromotionLinkInput{
		MemberID:   30,
		TargetType: "registration",
		Status:     "active",
	})
	require.NoError(t, err)
	require.NotNil(t, out)
	require.Equal(t, int64(30), created.MemberID)
	require.Equal(t, int64(88), created.ChannelOrgID)
}

func TestDistributionPromotionServiceCreateLinkForUserRejectsSameChannelNonDescendant(t *testing.T) {
	svc := NewDistributionPromotionService(
		&distributionPromotionLinkRepoStub{
			createFn: func(ctx context.Context, input DistributionPromotionLinkInput) (*DistributionPromotionLink, error) {
				t.Fatalf("Create should not be called for non-descendant member")
				return nil, nil
			},
		},
		&distributionPromotionLinkMemberRepoStub{
			listByUserIDFn: func(ctx context.Context, userID int64) ([]DistributionMemberView, error) {
				return []DistributionMemberView{{MemberID: 20, UserID: userID, ChannelOrgID: 88, RoleType: "kol1", Status: "active"}}, nil
			},
			getByIDFn: func(ctx context.Context, memberID int64) (*DistributionMemberView, error) {
				switch memberID {
				case 30:
					return &DistributionMemberView{MemberID: 30, ChannelOrgID: 88, RoleType: "kol1", ParentMemberID: ptrInt64(10), Status: "active"}, nil
				case 10:
					return &DistributionMemberView{MemberID: 10, ChannelOrgID: 88, RoleType: "agent", Status: "active"}, nil
				default:
					return nil, ErrDistributionMemberNotFound
				}
			},
		},
		&distributionPromotionLinkAttributionRepoStub{
			getByUserIDFn: func(ctx context.Context, userID int64) (*DistributionAttribution, error) {
				return &DistributionAttribution{UserID: userID, ChannelOrgID: 88}, nil
			},
		},
	)

	_, err := svc.CreateLinkForUser(context.Background(), 7, DistributionPromotionLinkInput{
		MemberID:   30,
		TargetType: "registration",
		Status:     "active",
	})
	require.ErrorIs(t, err, ErrInvalidDistributionPromotionLink)
}

func TestDistributionPromotionServiceListLinksForUserUsesChannelScope(t *testing.T) {
	svc := NewDistributionPromotionService(
		&distributionPromotionLinkRepoStub{
			listByChannelOrgIDFn: func(ctx context.Context, channelOrgID int64, params pagination.PaginationParams) ([]DistributionPromotionLink, *pagination.PaginationResult, error) {
				require.Equal(t, int64(88), channelOrgID)
				return []DistributionPromotionLink{{ID: 1, ChannelOrgID: 88, MemberID: 11, Code: "LINK-1"}}, paginationResult(1, params), nil
			},
		},
		&distributionPromotionLinkMemberRepoStub{
			listByUserIDFn: func(ctx context.Context, userID int64) ([]DistributionMemberView, error) {
				return []DistributionMemberView{
					{MemberID: 99, UserID: userID, ChannelOrgID: 88, RoleType: "manager", Status: "active"},
				}, nil
			},
			getByIDFn: func(ctx context.Context, memberID int64) (*DistributionMemberView, error) {
				return nil, ErrDistributionMemberNotFound
			},
		},
		&distributionPromotionLinkAttributionRepoStub{
			getByUserIDFn: func(ctx context.Context, userID int64) (*DistributionAttribution, error) {
				return &DistributionAttribution{UserID: userID, ChannelOrgID: 88}, nil
			},
		},
	)

	out, page, err := svc.ListLinksForUser(context.Background(), 7, pagination.PaginationParams{Page: 1, PageSize: 20})
	require.NoError(t, err)
	require.Len(t, out, 1)
	require.Equal(t, int64(1), page.Total)
	require.Equal(t, "LINK-1", out[0].Code)
}

func TestDistributionPromotionServiceListLinksForUserFiltersByMemberTreeForNonManager(t *testing.T) {
	svc := NewDistributionPromotionService(
		&distributionPromotionLinkRepoStub{
			listByChannelOrgIDFn: func(ctx context.Context, channelOrgID int64, params pagination.PaginationParams) ([]DistributionPromotionLink, *pagination.PaginationResult, error) {
				require.Equal(t, int64(88), channelOrgID)
				return []DistributionPromotionLink{
					{ID: 1, ChannelOrgID: 88, MemberID: 20, Code: "OWN"},
					{ID: 2, ChannelOrgID: 88, MemberID: 21, Code: "CHILD"},
					{ID: 3, ChannelOrgID: 88, MemberID: 30, Code: "FOREIGN"},
				}, paginationResult(3, params), nil
			},
		},
		&distributionPromotionLinkMemberRepoStub{
			listByUserIDFn: func(ctx context.Context, userID int64) ([]DistributionMemberView, error) {
				return []DistributionMemberView{
					{MemberID: 20, UserID: userID, ChannelOrgID: 88, RoleType: "kol1", Status: "active"},
				}, nil
			},
			getByIDFn: func(ctx context.Context, memberID int64) (*DistributionMemberView, error) {
				switch memberID {
				case 20:
					return &DistributionMemberView{MemberID: 20, ChannelOrgID: 88, RoleType: "kol1", Status: "active"}, nil
				case 21:
					return &DistributionMemberView{MemberID: 21, ChannelOrgID: 88, RoleType: "kol2", ParentMemberID: ptrInt64(20), Status: "active"}, nil
				case 30:
					return &DistributionMemberView{MemberID: 30, ChannelOrgID: 88, RoleType: "kol1", ParentMemberID: ptrInt64(10), Status: "active"}, nil
				case 10:
					return &DistributionMemberView{MemberID: 10, ChannelOrgID: 88, RoleType: "agent", Status: "active"}, nil
				default:
					return nil, ErrDistributionMemberNotFound
				}
			},
		},
		&distributionPromotionLinkAttributionRepoStub{
			getByUserIDFn: func(ctx context.Context, userID int64) (*DistributionAttribution, error) {
				return &DistributionAttribution{UserID: userID, ChannelOrgID: 88}, nil
			},
		},
	)
	svc.SetOrganizationRepository(&distributionMemberOrgRepoStub{})

	out, page, err := svc.ListLinksForUser(context.Background(), 7, pagination.PaginationParams{Page: 1, PageSize: 20})
	require.NoError(t, err)
	require.Len(t, out, 2)
	require.Equal(t, int64(2), page.Total)
	require.Equal(t, "OWN", out[0].Code)
	require.Equal(t, "CHILD", out[1].Code)
}

func TestDistributionPromotionServiceListLinksForUserUsesMemberScopeWhenNoAttribution(t *testing.T) {
	svc := NewDistributionPromotionService(
		&distributionPromotionLinkRepoStub{
			listByChannelOrgIDFn: func(ctx context.Context, channelOrgID int64, params pagination.PaginationParams) ([]DistributionPromotionLink, *pagination.PaginationResult, error) {
				require.Equal(t, int64(88), channelOrgID)
				return []DistributionPromotionLink{{ID: 1, ChannelOrgID: 88, MemberID: 11, Code: "LINK-1"}}, paginationResult(1, params), nil
			},
		},
		&distributionPromotionLinkMemberRepoStub{
			listByUserIDFn: func(ctx context.Context, userID int64) ([]DistributionMemberView, error) {
				return []DistributionMemberView{{MemberID: 22, UserID: userID, ChannelOrgID: 88, RoleType: "agent", Status: "active"}}, nil
			},
			getByIDFn: func(ctx context.Context, memberID int64) (*DistributionMemberView, error) {
				return nil, ErrDistributionMemberNotFound
			},
		},
		&distributionPromotionLinkAttributionRepoStub{
			getByUserIDFn: func(ctx context.Context, userID int64) (*DistributionAttribution, error) {
				return nil, ErrDistributionAttributionNotFound
			},
		},
	)

	out, page, err := svc.ListLinksForUser(context.Background(), 7, pagination.PaginationParams{Page: 1, PageSize: 20})
	require.NoError(t, err)
	require.Empty(t, out)
	require.Equal(t, int64(0), page.Total)
}

func paginationResult(total int64, params pagination.PaginationParams) *pagination.PaginationResult {
	return &pagination.PaginationResult{
		Total:    total,
		Page:     params.Page,
		PageSize: params.PageSize,
		Pages:    1,
	}
}
