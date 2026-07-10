package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func setupAPIKeyHandler(adminSvc service.AdminService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	h := NewAdminAPIKeyHandler(adminSvc, nil)
	router.PUT("/api/v1/admin/api-keys/:id", h.UpdateGroup)
	return router
}

func setupAPIKeyHandlerWithAPIKeyService(adminSvc service.AdminService, apiKeySvc *service.APIKeyService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	h := NewAdminAPIKeyHandler(adminSvc, apiKeySvc)
	router.POST("/api/v1/admin/users/:id/api-keys", h.CreateForUser)
	router.PUT("/api/v1/admin/api-keys/:id", h.UpdateGroup)
	return router
}

func TestAdminAPIKeyHandler_CreateForUser(t *testing.T) {
	userID := int64(42)
	groupID := int64(2)
	apiKeyRepo := &adminCreateAPIKeyRepo{}
	userRepo := &adminCreateAPIKeyUserRepo{
		users: map[int64]*service.User{
			userID: {
				ID:            userID,
				Email:         "target@example.com",
				Status:        service.StatusActive,
				AllowedGroups: []int64{groupID},
			},
		},
	}
	groupRepo := &adminCreateAPIKeyGroupRepo{
		groups: map[int64]*service.Group{
			groupID: {
				ID:        groupID,
				Name:      "target-group",
				Status:    service.StatusActive,
				Platform:  service.PlatformAnthropic,
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			},
		},
	}
	apiKeySvc := service.NewAPIKeyService(apiKeyRepo, userRepo, groupRepo, &adminCreateAPIKeySubscriptionRepo{}, nil, nil, nil)
	router := setupAPIKeyHandlerWithAPIKeyService(newStubAdminService(), apiKeySvc)

	body := map[string]any{
		"name":            "admin-created",
		"group_id":        groupID,
		"custom_key":      "admin_custom_key_123",
		"ip_whitelist":    []string{"192.0.2.1"},
		"ip_blacklist":    []string{"198.51.100.0/24"},
		"quota":           12.5,
		"expires_in_days": 7,
		"rate_limit_5h":   1.25,
		"rate_limit_1d":   2.5,
		"rate_limit_7d":   3.75,
	}
	raw, err := json.Marshal(body)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/users/42/api-keys", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, apiKeyRepo.created)
	require.Equal(t, userID, apiKeyRepo.created.UserID)
	require.Equal(t, "admin-created", apiKeyRepo.created.Name)
	require.NotNil(t, apiKeyRepo.created.GroupID)
	require.Equal(t, groupID, *apiKeyRepo.created.GroupID)
	require.Equal(t, "admin_custom_key_123", apiKeyRepo.created.Key)
	require.Equal(t, []string{"192.0.2.1"}, apiKeyRepo.created.IPWhitelist)
	require.Equal(t, []string{"198.51.100.0/24"}, apiKeyRepo.created.IPBlacklist)
	require.Equal(t, 12.5, apiKeyRepo.created.Quota)
	require.Equal(t, 1.25, apiKeyRepo.created.RateLimit5h)
	require.Equal(t, 2.5, apiKeyRepo.created.RateLimit1d)
	require.Equal(t, 3.75, apiKeyRepo.created.RateLimit7d)
	require.NotNil(t, apiKeyRepo.created.ExpiresAt)

	var resp struct {
		Code int `json:"code"`
		Data struct {
			ID     int64  `json:"id"`
			UserID int64  `json:"user_id"`
			Key    string `json:"key"`
			Name   string `json:"name"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, 0, resp.Code)
	require.Equal(t, userID, resp.Data.UserID)
	require.Equal(t, "admin_custom_key_123", resp.Data.Key)
	require.Equal(t, "admin-created", resp.Data.Name)
}

func TestAdminAPIKeyHandler_CreateForUser_InvalidUserID(t *testing.T) {
	apiKeySvc := service.NewAPIKeyService(&adminCreateAPIKeyRepo{}, &adminCreateAPIKeyUserRepo{}, nil, nil, nil, nil, nil)
	router := setupAPIKeyHandlerWithAPIKeyService(newStubAdminService(), apiKeySvc)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/users/not-a-number/api-keys", bytes.NewBufferString(`{"name":"x"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Contains(t, rec.Body.String(), "Invalid user ID")
}

func TestAdminAPIKeyHandler_CreateForUser_InvalidJSON(t *testing.T) {
	apiKeySvc := service.NewAPIKeyService(&adminCreateAPIKeyRepo{}, &adminCreateAPIKeyUserRepo{}, nil, nil, nil, nil, nil)
	router := setupAPIKeyHandlerWithAPIKeyService(newStubAdminService(), apiKeySvc)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/users/42/api-keys", bytes.NewBufferString(`{bad json`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Contains(t, rec.Body.String(), "Invalid request")
}

func TestAdminAPIKeyHandler_UpdateGroup_InvalidID(t *testing.T) {
	router := setupAPIKeyHandler(newStubAdminService())
	body := `{"group_id": 2}`

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/api-keys/abc", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Contains(t, rec.Body.String(), "Invalid API key ID")
}

func TestAdminAPIKeyHandler_UpdateGroup_InvalidJSON(t *testing.T) {
	router := setupAPIKeyHandler(newStubAdminService())

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/api-keys/10", bytes.NewBufferString(`{bad json`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Contains(t, rec.Body.String(), "Invalid request")
}

func TestAdminAPIKeyHandler_UpdateGroup_KeyNotFound(t *testing.T) {
	router := setupAPIKeyHandler(newStubAdminService())
	body := `{"group_id": 2}`

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/api-keys/999", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	// ErrAPIKeyNotFound maps to 404
	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestAdminAPIKeyHandler_UpdateGroup_BindGroup(t *testing.T) {
	router := setupAPIKeyHandler(newStubAdminService())
	body := `{"group_id": 2}`

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/api-keys/10", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp struct {
		Code int             `json:"code"`
		Data json.RawMessage `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, 0, resp.Code)

	var data struct {
		APIKey struct {
			ID      int64  `json:"id"`
			GroupID *int64 `json:"group_id"`
		} `json:"api_key"`
		AutoGrantedGroupAccess bool `json:"auto_granted_group_access"`
	}
	require.NoError(t, json.Unmarshal(resp.Data, &data))
	require.Equal(t, int64(10), data.APIKey.ID)
	require.NotNil(t, data.APIKey.GroupID)
	require.Equal(t, int64(2), *data.APIKey.GroupID)
}

func TestAdminAPIKeyHandler_UpdateGroup_Unbind(t *testing.T) {
	svc := newStubAdminService()
	gid := int64(2)
	svc.apiKeys[0].GroupID = &gid
	router := setupAPIKeyHandler(svc)
	body := `{"group_id": 0}`

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/api-keys/10", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp struct {
		Data struct {
			APIKey struct {
				GroupID *int64 `json:"group_id"`
			} `json:"api_key"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Nil(t, resp.Data.APIKey.GroupID)
}

func TestAdminAPIKeyHandler_ResetRateLimitUsage(t *testing.T) {
	svc := newStubAdminService()
	now := time.Now()
	svc.apiKeys[0].Usage5h = 1.2
	svc.apiKeys[0].Usage1d = 3.4
	svc.apiKeys[0].Usage7d = 5.6
	svc.apiKeys[0].Window5hStart = &now
	svc.apiKeys[0].Window1dStart = &now
	svc.apiKeys[0].Window7dStart = &now
	router := setupAPIKeyHandler(svc)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/api-keys/10", bytes.NewBufferString(`{"reset_rate_limit_usage":true}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp struct {
		Data struct {
			APIKey struct {
				Usage5h       float64    `json:"usage_5h"`
				Usage1d       float64    `json:"usage_1d"`
				Usage7d       float64    `json:"usage_7d"`
				Window5hStart *time.Time `json:"window_5h_start"`
				Window1dStart *time.Time `json:"window_1d_start"`
				Window7dStart *time.Time `json:"window_7d_start"`
			} `json:"api_key"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Zero(t, resp.Data.APIKey.Usage5h)
	require.Zero(t, resp.Data.APIKey.Usage1d)
	require.Zero(t, resp.Data.APIKey.Usage7d)
	require.Nil(t, resp.Data.APIKey.Window5hStart)
	require.Nil(t, resp.Data.APIKey.Window1dStart)
	require.Nil(t, resp.Data.APIKey.Window7dStart)
}

func TestAdminAPIKeyHandler_UpdateGroup_ServiceError(t *testing.T) {
	svc := &failingUpdateGroupService{
		stubAdminService: newStubAdminService(),
		err:              errors.New("internal failure"),
	}
	router := setupAPIKeyHandler(svc)
	body := `{"group_id": 2}`

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/api-keys/10", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
}

// H2: empty body → group_id is nil → no-op, returns original key
func TestAdminAPIKeyHandler_UpdateGroup_EmptyBody_NoChange(t *testing.T) {
	router := setupAPIKeyHandler(newStubAdminService())

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/api-keys/10", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp struct {
		Code int `json:"code"`
		Data struct {
			APIKey struct {
				ID int64 `json:"id"`
			} `json:"api_key"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, 0, resp.Code)
	require.Equal(t, int64(10), resp.Data.APIKey.ID)
}

// M2: service returns GROUP_NOT_ACTIVE → handler maps to 400
func TestAdminAPIKeyHandler_UpdateGroup_GroupNotActive(t *testing.T) {
	svc := &failingUpdateGroupService{
		stubAdminService: newStubAdminService(),
		err:              infraerrors.BadRequest("GROUP_NOT_ACTIVE", "target group is not active"),
	}
	router := setupAPIKeyHandler(svc)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/api-keys/10", bytes.NewBufferString(`{"group_id": 5}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Contains(t, rec.Body.String(), "GROUP_NOT_ACTIVE")
}

// M2: service returns INVALID_GROUP_ID → handler maps to 400
func TestAdminAPIKeyHandler_UpdateGroup_NegativeGroupID(t *testing.T) {
	svc := &failingUpdateGroupService{
		stubAdminService: newStubAdminService(),
		err:              infraerrors.BadRequest("INVALID_GROUP_ID", "group_id must be non-negative"),
	}
	router := setupAPIKeyHandler(svc)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/api-keys/10", bytes.NewBufferString(`{"group_id": -5}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Contains(t, rec.Body.String(), "INVALID_GROUP_ID")
}

// failingUpdateGroupService overrides AdminUpdateAPIKeyGroupID to return an error.
type failingUpdateGroupService struct {
	*stubAdminService
	err error
}

func (f *failingUpdateGroupService) AdminUpdateAPIKeyGroupID(_ context.Context, _ int64, _ *int64) (*service.AdminUpdateAPIKeyGroupIDResult, error) {
	return nil, f.err
}

type adminCreateAPIKeyRepo struct {
	created *service.APIKey
}

func (r *adminCreateAPIKeyRepo) Create(_ context.Context, key *service.APIKey) error {
	cp := *key
	cp.ID = 101
	now := time.Now().UTC()
	cp.CreatedAt = now
	cp.UpdatedAt = now
	r.created = &cp
	*key = cp
	return nil
}

func (r *adminCreateAPIKeyRepo) GetByID(context.Context, int64) (*service.APIKey, error) {
	return nil, service.ErrAPIKeyNotFound
}

func (r *adminCreateAPIKeyRepo) GetKeyAndOwnerID(context.Context, int64) (string, int64, error) {
	return "", 0, service.ErrAPIKeyNotFound
}

func (r *adminCreateAPIKeyRepo) GetByKey(context.Context, string) (*service.APIKey, error) {
	return nil, service.ErrAPIKeyNotFound
}

func (r *adminCreateAPIKeyRepo) GetByKeyForAuth(context.Context, string) (*service.APIKey, error) {
	return nil, service.ErrAPIKeyNotFound
}

func (r *adminCreateAPIKeyRepo) Update(context.Context, *service.APIKey) error { return nil }
func (r *adminCreateAPIKeyRepo) Delete(context.Context, int64) error           { return nil }
func (r *adminCreateAPIKeyRepo) DeleteWithAudit(context.Context, int64) error  { return nil }

func (r *adminCreateAPIKeyRepo) ListByUserID(context.Context, int64, pagination.PaginationParams, service.APIKeyListFilters) ([]service.APIKey, *pagination.PaginationResult, error) {
	return nil, nil, nil
}

func (r *adminCreateAPIKeyRepo) VerifyOwnership(context.Context, int64, []int64) ([]int64, error) {
	return nil, nil
}

func (r *adminCreateAPIKeyRepo) CountByUserID(context.Context, int64) (int64, error) { return 0, nil }
func (r *adminCreateAPIKeyRepo) ExistsByKey(context.Context, string) (bool, error)   { return false, nil }

func (r *adminCreateAPIKeyRepo) ListByGroupID(context.Context, int64, pagination.PaginationParams) ([]service.APIKey, *pagination.PaginationResult, error) {
	return nil, nil, nil
}

func (r *adminCreateAPIKeyRepo) SearchAPIKeys(context.Context, int64, string, int) ([]service.APIKey, error) {
	return nil, nil
}

func (r *adminCreateAPIKeyRepo) ClearGroupIDByGroupID(context.Context, int64) (int64, error) {
	return 0, nil
}

func (r *adminCreateAPIKeyRepo) UpdateGroupIDByUserAndGroup(context.Context, int64, int64, int64) (int64, error) {
	return 0, nil
}

func (r *adminCreateAPIKeyRepo) CountByGroupID(context.Context, int64) (int64, error) {
	return 0, nil
}

func (r *adminCreateAPIKeyRepo) ListKeysByUserID(context.Context, int64) ([]string, error) {
	return nil, nil
}

func (r *adminCreateAPIKeyRepo) ListKeysByGroupID(context.Context, int64) ([]string, error) {
	return nil, nil
}

func (r *adminCreateAPIKeyRepo) IncrementQuotaUsed(context.Context, int64, float64) (float64, error) {
	return 0, nil
}

func (r *adminCreateAPIKeyRepo) UpdateLastUsed(context.Context, int64, time.Time) error {
	return nil
}

func (r *adminCreateAPIKeyRepo) IncrementRateLimitUsage(context.Context, int64, float64) error {
	return nil
}

func (r *adminCreateAPIKeyRepo) ResetRateLimitWindows(context.Context, int64) error {
	return nil
}

func (r *adminCreateAPIKeyRepo) GetRateLimitData(context.Context, int64) (*service.APIKeyRateLimitData, error) {
	return nil, service.ErrAPIKeyNotFound
}

type adminCreateAPIKeyUserRepo struct {
	users map[int64]*service.User
}

func (r *adminCreateAPIKeyUserRepo) Create(context.Context, *service.User) error { return nil }

func (r *adminCreateAPIKeyUserRepo) GetByID(_ context.Context, id int64) (*service.User, error) {
	if r.users != nil {
		if user, ok := r.users[id]; ok {
			cp := *user
			return &cp, nil
		}
	}
	return nil, service.ErrUserNotFound
}

func (r *adminCreateAPIKeyUserRepo) GetByIDIncludeDeleted(ctx context.Context, id int64) (*service.User, error) {
	return r.GetByID(ctx, id)
}

func (r *adminCreateAPIKeyUserRepo) GetByEmail(context.Context, string) (*service.User, error) {
	return nil, service.ErrUserNotFound
}

func (r *adminCreateAPIKeyUserRepo) GetFirstAdmin(context.Context) (*service.User, error) {
	return nil, service.ErrUserNotFound
}

func (r *adminCreateAPIKeyUserRepo) Update(context.Context, *service.User) error { return nil }
func (r *adminCreateAPIKeyUserRepo) Delete(context.Context, int64) error         { return nil }

func (r *adminCreateAPIKeyUserRepo) GetUserAvatar(context.Context, int64) (*service.UserAvatar, error) {
	return nil, nil
}

func (r *adminCreateAPIKeyUserRepo) UpsertUserAvatar(context.Context, int64, service.UpsertUserAvatarInput) (*service.UserAvatar, error) {
	return nil, nil
}

func (r *adminCreateAPIKeyUserRepo) DeleteUserAvatar(context.Context, int64) error { return nil }

func (r *adminCreateAPIKeyUserRepo) List(context.Context, pagination.PaginationParams) ([]service.User, *pagination.PaginationResult, error) {
	return nil, nil, nil
}

func (r *adminCreateAPIKeyUserRepo) ListWithFilters(context.Context, pagination.PaginationParams, service.UserListFilters) ([]service.User, *pagination.PaginationResult, error) {
	return nil, nil, nil
}

func (r *adminCreateAPIKeyUserRepo) GetLatestUsedAtByUserIDs(context.Context, []int64) (map[int64]*time.Time, error) {
	return nil, nil
}

func (r *adminCreateAPIKeyUserRepo) GetLatestUsedAtByUserID(context.Context, int64) (*time.Time, error) {
	return nil, nil
}

func (r *adminCreateAPIKeyUserRepo) UpdateUserLastActiveAt(context.Context, int64, time.Time) error {
	return nil
}

func (r *adminCreateAPIKeyUserRepo) UpdateBalance(context.Context, int64, float64) error {
	return nil
}

func (r *adminCreateAPIKeyUserRepo) DeductBalance(context.Context, int64, float64) error {
	return nil
}

func (r *adminCreateAPIKeyUserRepo) UpdateConcurrency(context.Context, int64, int) error {
	return nil
}

func (r *adminCreateAPIKeyUserRepo) BatchSetConcurrency(context.Context, []int64, int) (int, error) {
	return 0, nil
}

func (r *adminCreateAPIKeyUserRepo) BatchAddConcurrency(context.Context, []int64, int) (int, error) {
	return 0, nil
}

func (r *adminCreateAPIKeyUserRepo) ExistsByEmail(context.Context, string) (bool, error) {
	return false, nil
}

func (r *adminCreateAPIKeyUserRepo) RemoveGroupFromAllowedGroups(context.Context, int64) (int64, error) {
	return 0, nil
}

func (r *adminCreateAPIKeyUserRepo) AddGroupToAllowedGroups(context.Context, int64, int64) error {
	return nil
}

func (r *adminCreateAPIKeyUserRepo) RemoveGroupFromUserAllowedGroups(context.Context, int64, int64) error {
	return nil
}

func (r *adminCreateAPIKeyUserRepo) ListUserAuthIdentities(context.Context, int64) ([]service.UserAuthIdentityRecord, error) {
	return nil, nil
}

func (r *adminCreateAPIKeyUserRepo) UnbindUserAuthProvider(context.Context, int64, string) error {
	return nil
}

func (r *adminCreateAPIKeyUserRepo) UpdateTotpSecret(context.Context, int64, *string) error {
	return nil
}

func (r *adminCreateAPIKeyUserRepo) EnableTotp(context.Context, int64) error  { return nil }
func (r *adminCreateAPIKeyUserRepo) DisableTotp(context.Context, int64) error { return nil }

type adminCreateAPIKeyGroupRepo struct {
	groups map[int64]*service.Group
}

func (r *adminCreateAPIKeyGroupRepo) Create(context.Context, *service.Group) error { return nil }

func (r *adminCreateAPIKeyGroupRepo) GetByID(_ context.Context, id int64) (*service.Group, error) {
	if r.groups != nil {
		if group, ok := r.groups[id]; ok {
			cp := *group
			return &cp, nil
		}
	}
	return nil, service.ErrGroupNotFound
}

func (r *adminCreateAPIKeyGroupRepo) GetByIDLite(ctx context.Context, id int64) (*service.Group, error) {
	return r.GetByID(ctx, id)
}

func (r *adminCreateAPIKeyGroupRepo) Update(context.Context, *service.Group) error { return nil }
func (r *adminCreateAPIKeyGroupRepo) Delete(context.Context, int64) error          { return nil }

func (r *adminCreateAPIKeyGroupRepo) DeleteCascade(context.Context, int64) ([]int64, error) {
	return nil, nil
}

func (r *adminCreateAPIKeyGroupRepo) List(context.Context, pagination.PaginationParams) ([]service.Group, *pagination.PaginationResult, error) {
	return nil, nil, nil
}

func (r *adminCreateAPIKeyGroupRepo) ListWithFilters(context.Context, pagination.PaginationParams, string, string, string, *bool) ([]service.Group, *pagination.PaginationResult, error) {
	return nil, nil, nil
}

func (r *adminCreateAPIKeyGroupRepo) ListActive(context.Context) ([]service.Group, error) {
	out := make([]service.Group, 0, len(r.groups))
	for _, group := range r.groups {
		out = append(out, *group)
	}
	return out, nil
}

func (r *adminCreateAPIKeyGroupRepo) ListActiveByPlatform(context.Context, string) ([]service.Group, error) {
	return nil, nil
}

func (r *adminCreateAPIKeyGroupRepo) ExistsByName(context.Context, string) (bool, error) {
	return false, nil
}

func (r *adminCreateAPIKeyGroupRepo) GetAccountCount(context.Context, int64) (int64, int64, error) {
	return 0, 0, nil
}

func (r *adminCreateAPIKeyGroupRepo) DeleteAccountGroupsByGroupID(context.Context, int64) (int64, error) {
	return 0, nil
}

func (r *adminCreateAPIKeyGroupRepo) GetAccountIDsByGroupIDs(context.Context, []int64) ([]int64, error) {
	return nil, nil
}

func (r *adminCreateAPIKeyGroupRepo) BindAccountsToGroup(context.Context, int64, []int64) error {
	return nil
}

func (r *adminCreateAPIKeyGroupRepo) UpdateSortOrders(context.Context, []service.GroupSortOrderUpdate) error {
	return nil
}

type adminCreateAPIKeySubscriptionRepo struct{}

func (r *adminCreateAPIKeySubscriptionRepo) Create(context.Context, *service.UserSubscription) error {
	return nil
}

func (r *adminCreateAPIKeySubscriptionRepo) GetByID(context.Context, int64) (*service.UserSubscription, error) {
	return nil, service.ErrSubscriptionNotFound
}

func (r *adminCreateAPIKeySubscriptionRepo) GetByUserIDAndGroupID(context.Context, int64, int64) (*service.UserSubscription, error) {
	return nil, service.ErrSubscriptionNotFound
}

func (r *adminCreateAPIKeySubscriptionRepo) GetActiveByUserIDAndGroupID(context.Context, int64, int64) (*service.UserSubscription, error) {
	return nil, service.ErrSubscriptionNotFound
}

func (r *adminCreateAPIKeySubscriptionRepo) Update(context.Context, *service.UserSubscription) error {
	return nil
}

func (r *adminCreateAPIKeySubscriptionRepo) Delete(context.Context, int64) error { return nil }

func (r *adminCreateAPIKeySubscriptionRepo) ListByUserID(context.Context, int64) ([]service.UserSubscription, error) {
	return nil, nil
}

func (r *adminCreateAPIKeySubscriptionRepo) ListActiveByUserID(context.Context, int64) ([]service.UserSubscription, error) {
	return nil, nil
}

func (r *adminCreateAPIKeySubscriptionRepo) ListByGroupID(context.Context, int64, pagination.PaginationParams) ([]service.UserSubscription, *pagination.PaginationResult, error) {
	return nil, nil, nil
}

func (r *adminCreateAPIKeySubscriptionRepo) List(context.Context, pagination.PaginationParams, *int64, *int64, string, string, string, string) ([]service.UserSubscription, *pagination.PaginationResult, error) {
	return nil, nil, nil
}

func (r *adminCreateAPIKeySubscriptionRepo) ExistsByUserIDAndGroupID(context.Context, int64, int64) (bool, error) {
	return false, nil
}

func (r *adminCreateAPIKeySubscriptionRepo) ExtendExpiry(context.Context, int64, time.Time) error {
	return nil
}

func (r *adminCreateAPIKeySubscriptionRepo) UpdateStatus(context.Context, int64, string) error {
	return nil
}

func (r *adminCreateAPIKeySubscriptionRepo) UpdateNotes(context.Context, int64, string) error {
	return nil
}

func (r *adminCreateAPIKeySubscriptionRepo) ActivateWindows(context.Context, int64, time.Time) error {
	return nil
}

func (r *adminCreateAPIKeySubscriptionRepo) ResetDailyUsage(context.Context, int64, time.Time) error {
	return nil
}

func (r *adminCreateAPIKeySubscriptionRepo) ResetWeeklyUsage(context.Context, int64, time.Time) error {
	return nil
}

func (r *adminCreateAPIKeySubscriptionRepo) ResetMonthlyUsage(context.Context, int64, time.Time) error {
	return nil
}

func (r *adminCreateAPIKeySubscriptionRepo) IncrementUsage(context.Context, int64, float64) error {
	return nil
}

func (r *adminCreateAPIKeySubscriptionRepo) BatchUpdateExpiredStatus(context.Context) (int64, error) {
	return 0, nil
}
