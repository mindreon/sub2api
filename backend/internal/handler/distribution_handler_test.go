package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	servermiddleware "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type distributionHandlerScopeAttributionRepoStub struct {
	channelOrgID int64
}

func (s *distributionHandlerScopeAttributionRepoStub) GetByUserID(ctx context.Context, userID int64) (*service.DistributionAttribution, error) {
	return &service.DistributionAttribution{
		UserID:       userID,
		ChannelOrgID: s.channelOrgID,
		BoundAt:      time.Now().UTC(),
		BoundSource:  "registration",
		BoundBy:      "system",
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}, nil
}

func (s *distributionHandlerScopeAttributionRepoStub) ListByChannelOrgID(ctx context.Context, channelOrgID int64, params pagination.PaginationParams) ([]service.DistributionAttributionView, *pagination.PaginationResult, error) {
	return nil, &pagination.PaginationResult{Total: 0, Page: params.Page, PageSize: params.PageSize, Pages: 0}, nil
}

type distributionHandlerScopeMemberRepoStub struct {
	byUser map[int64][]service.DistributionMemberView
}

func (s *distributionHandlerScopeMemberRepoStub) ListByChannelOrgID(ctx context.Context, channelOrgID int64, params pagination.PaginationParams, roleType string) ([]service.DistributionMemberView, *pagination.PaginationResult, error) {
	return nil, &pagination.PaginationResult{Total: 0, Page: params.Page, PageSize: params.PageSize, Pages: 0}, nil
}

func (s *distributionHandlerScopeMemberRepoStub) ListByUserID(ctx context.Context, userID int64) ([]service.DistributionMemberView, error) {
	return s.byUser[userID], nil
}

type distributionHandlerScopeStatsRepoStub struct {
	summary *service.DistributionChannelSummary
}

func (s *distributionHandlerScopeStatsRepoStub) GetChannelSummary(ctx context.Context, channelOrgID int64) (*service.DistributionChannelSummary, error) {
	return s.summary, nil
}

type distributionHandlerEnvelope struct {
	Code    int                    `json:"code"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data"`
}

func TestDistributionHandlerGetOverview_HidesSummaryForNonManager(t *testing.T) {
	gin.SetMode(gin.TestMode)

	scopeSvc := service.NewDistributionScopeService(
		&distributionHandlerScopeAttributionRepoStub{channelOrgID: 88},
		&distributionHandlerScopeMemberRepoStub{
			byUser: map[int64][]service.DistributionMemberView{
				7: {{MemberID: 70, UserID: 7, ChannelOrgID: 88, RoleType: "agent", Status: "active"}},
			},
		},
		nil,
		nil,
		&distributionHandlerScopeStatsRepoStub{
			summary: &service.DistributionChannelSummary{
				Organization: service.DistributionOrganization{ID: 88, Type: "reseller", Name: "Channel A", Status: "active"},
			},
		},
	)

	h := NewDistributionHandler(scopeSvc, nil, nil, nil, nil)
	r := gin.New()
	r.GET("/overview", func(c *gin.Context) {
		c.Set(string(servermiddleware.ContextKeyUser), servermiddleware.AuthSubject{UserID: 7})
		h.GetOverview(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/overview", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var out distributionHandlerEnvelope
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
	require.Equal(t, 0, out.Code)
	require.Equal(t, false, out.Data["can_manage_channel"])
	require.Equal(t, float64(88), out.Data["channel_org_id"])
	require.Nil(t, out.Data["summary"])
}

func TestDistributionHandlerGetOverview_ReturnsEmptyWhenNoChannel(t *testing.T) {
	gin.SetMode(gin.TestMode)

	scopeSvc := service.NewDistributionScopeService(
		&distributionHandlerScopeAttributionRepoStub{},
		&distributionHandlerScopeMemberRepoStub{},
		nil,
		nil,
		nil,
	)

	h := NewDistributionHandler(scopeSvc, nil, nil, nil, nil)
	r := gin.New()
	r.GET("/overview", func(c *gin.Context) {
		c.Set(string(servermiddleware.ContextKeyUser), servermiddleware.AuthSubject{UserID: 1})
		h.GetOverview(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/overview", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var out distributionHandlerEnvelope
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
	require.Equal(t, 0, out.Code)
	require.Equal(t, float64(1), out.Data["user_id"])
	require.Equal(t, float64(0), out.Data["channel_org_id"])
	require.Equal(t, false, out.Data["can_manage_channel"])
	require.Nil(t, out.Data["summary"])
}

func TestDistributionHandlerGetOverview_ReturnsSummaryForManager(t *testing.T) {
	gin.SetMode(gin.TestMode)

	scopeSvc := service.NewDistributionScopeService(
		&distributionHandlerScopeAttributionRepoStub{channelOrgID: 88},
		&distributionHandlerScopeMemberRepoStub{
			byUser: map[int64][]service.DistributionMemberView{
				9: {{MemberID: 90, UserID: 9, ChannelOrgID: 88, RoleType: "manager", Status: "active"}},
			},
		},
		nil,
		nil,
		&distributionHandlerScopeStatsRepoStub{
			summary: &service.DistributionChannelSummary{
				Organization: service.DistributionOrganization{ID: 88, Type: "reseller", Name: "Channel A", Status: "active"},
			},
		},
	)

	h := NewDistributionHandler(scopeSvc, nil, nil, nil, nil)
	r := gin.New()
	r.GET("/overview", func(c *gin.Context) {
		c.Set(string(servermiddleware.ContextKeyUser), servermiddleware.AuthSubject{UserID: 9})
		h.GetOverview(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/overview", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var out distributionHandlerEnvelope
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
	require.Equal(t, 0, out.Code)
	require.Equal(t, true, out.Data["can_manage_channel"])
	require.Equal(t, float64(88), out.Data["channel_org_id"])
	require.NotNil(t, out.Data["summary"])

	rawSummary, ok := out.Data["summary"].(map[string]interface{})
	require.True(t, ok)
	rawOrg, ok := rawSummary["organization"].(map[string]interface{})
	require.True(t, ok)
	require.Equal(t, float64(88), rawOrg["id"])
}

