package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/media"
	servermiddleware "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestWriteCommonStyleTaskErrorMapsUpstreamFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/video/generations", nil)

	writeCommonStyleTaskError(c, fmt.Errorf("%w: upstream http 404", media.ErrUpstreamRequest))

	require.Equal(t, http.StatusBadGateway, rec.Code)
	var resp map[string]map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, "upstream_error", resp["error"]["code"])
	require.Equal(t, "upstream request failed", resp["error"]["message"])
}

type mediaHandlerTestProvider struct {
	lastTask   *media.Task
	upstreamID string
}

func (p *mediaHandlerTestProvider) Submit(ctx context.Context, task *media.Task) (string, error) {
	_ = ctx
	cp := *task
	p.lastTask = &cp
	if p.upstreamID != "" {
		return p.upstreamID, nil
	}
	return "upstream-task-1", nil
}

func (p *mediaHandlerTestProvider) QueryStatus(ctx context.Context, task *media.Task) (*media.ProviderStatus, error) {
	_ = ctx
	_ = task
	return &media.ProviderStatus{State: media.ProviderInProgress}, nil
}

type mediaHandlerTestFactory struct {
	provider media.Provider
}

func (f mediaHandlerTestFactory) NewProvider(sel media.AccountSelection, model string) (media.Provider, error) {
	_ = sel
	_ = model
	return f.provider, nil
}

type mediaHandlerTestLoader struct{}

func (mediaHandlerTestLoader) Load(ctx context.Context, accountID int64) (media.AccountSelection, error) {
	_ = ctx
	return media.AccountSelection{AccountID: accountID, Platform: service.PlatformVolcengine, APIKey: "test-key"}, nil
}

func newCommonStyleMediaHandlerTest(t *testing.T, provider media.Provider) (*MediaHandler, *media.MemoryStore) {
	t.Helper()
	mem := media.NewMemoryStore()
	holds := mem.MemoryHoldStore()
	conv := media.NewStaticCurrencyConverter(media.CurrencyUSD, map[media.Currency]float64{media.CurrencyCNY: 0.14})
	quoter := media.NewQuoter(media.SeedanceRuleProvider(), conv, nil, media.CurrencyUSD)
	balance := media.BalanceReaderFunc(func(ctx context.Context, userID int64) (float64, error) {
		_ = ctx
		_ = userID
		return 100, nil
	})
	ledger := media.NewLedger(
		quoter,
		media.ChargerFunc(func(ctx context.Context, req media.ChargeRequest) (*media.ChargeResult, error) {
			_ = ctx
			_ = req
			return &media.ChargeResult{Applied: true}, nil
		}),
		mem,
		holds,
		media.NewHoldAwareBalance(balance, holds),
	)
	taskService := media.NewTaskService(ledger, nil, mediaHandlerTestLoader{}, mediaHandlerTestFactory{provider: provider}, nil)
	return NewMediaHandler(taskService, mem, nil), mem
}

func TestMediaHandlerSubmitCommonVideoGenerationAcceptsMetadataPayload(t *testing.T) {
	gin.SetMode(gin.TestMode)
	provider := &mediaHandlerTestProvider{upstreamID: "up-common-1"}
	handler, mem := newCommonStyleMediaHandlerTest(t, provider)
	body := `{
		"model":"dreamina-seedance-2-0-260128",
		"prompt":"小蝌蚪找妈妈",
		"metadata":{
			"resolution":"4K",
			"ratio":"1:1",
			"duration":15,
			"content":[
				{"type":"video_url","role":"reference_video","video_url":{"url":"https://example.com/ref.mp4"}},
				{"type":"audio_url","role":"reference_audio","audio_url":{"url":"https://example.com/ref.mp3"}}
			]
		}
	}`

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/video/generations", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set(string(servermiddleware.ContextKeyAPIKey), &service.APIKey{
		ID:     11,
		UserID: 22,
		Group:  &service.Group{Platform: service.PlatformVolcengine, AllowMediaGeneration: true, RateMultiplier: 1},
	})

	handler.SubmitCommonVideoGeneration(c)

	require.Equal(t, http.StatusAccepted, rec.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, "video", resp["object"])
	require.Equal(t, "dreamina-seedance-2-0-260128", resp["model"])
	require.Equal(t, "queued", resp["status"])
	require.NotEmpty(t, resp["id"])
	require.Equal(t, resp["id"], resp["task_id"])
	require.InDelta(t, float64(0), resp["progress"], 0)
	require.NotZero(t, resp["created_at"])

	taskID := resp["task_id"].(string)
	stored, err := mem.GetByTaskID(context.Background(), taskID)
	require.NoError(t, err)
	require.Equal(t, "小蝌蚪找妈妈", stored.RequestParams["prompt"])
	require.Equal(t, "4k", stored.RequestParams["resolution"])
	require.Equal(t, "1:1", stored.RequestParams["ratio"])
	require.InDelta(t, float64(15), stored.RequestParams["duration"], 0)
	require.True(t, stored.RequestParams["has_video_input"].(bool))
	require.True(t, stored.RequestParams["has_audio"].(bool))
	require.NotNil(t, provider.lastTask)
	require.Equal(t, stored.RequestParams, provider.lastTask.RequestParams)
}

func TestMediaHandlerGetCommonVideoGenerationMapsCompletedTask(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, mem := newCommonStyleMediaHandlerTest(t, &mediaHandlerTestProvider{})
	now := time.Now()
	actualCost := 0.1
	require.NoError(t, mem.Create(context.Background(), &media.Task{
		TaskID:        "task_done",
		UserID:        22,
		Model:         "dreamina-seedance-2-0-260128",
		MediaType:     "video",
		Status:        media.TaskCompleted,
		ReservedCost:  0.2,
		ActualCost:    &actualCost,
		ResultURL:     "https://example.com/video.mp4",
		UpstreamUsage: map[string]any{"status": "succeeded", "id": "up-1"},
		CreatedAt:     now,
		UpdatedAt:     now,
		SettledAt:     &now,
	}))

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/v1/video/generations/task_done", nil)
	c.Params = gin.Params{{Key: "task_id", Value: "task_done"}}
	c.Set(string(servermiddleware.ContextKeyAPIKey), &service.APIKey{ID: 11, UserID: 22})

	handler.GetCommonVideoGeneration(c)

	require.Equal(t, http.StatusOK, rec.Code)
	var resp struct {
		Code    string         `json:"code"`
		Message string         `json:"message"`
		Data    map[string]any `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, "success", resp.Code)
	require.Equal(t, "", resp.Message)
	require.Equal(t, "task_done", resp.Data["task_id"])
	require.Equal(t, "dreamina-seedance-2-0-260128", resp.Data["model"])
	require.Equal(t, "SUCCESS", resp.Data["status"])
	require.Equal(t, "100%", resp.Data["progress"])
	require.Equal(t, "https://example.com/video.mp4", resp.Data["result_url"])
	require.InDelta(t, 0.2, resp.Data["quota"], 0.0001)
	require.InDelta(t, 0.1, resp.Data["actual_cost"], 0.0001)
	require.Equal(t, "succeeded", resp.Data["data"].(map[string]any)["status"])
}
