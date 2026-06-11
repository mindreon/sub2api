package admin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type adminUsageByRequestRepo struct {
	service.UsageLogRepository
	requestID string
	log       *service.UsageLog
}

func (r *adminUsageByRequestRepo) GetByRequestID(_ context.Context, requestID string, apiKeyID int64) (*service.UsageLog, error) {
	if r.log != nil && r.requestID == requestID && (apiKeyID == 0 || r.log.APIKeyID == apiKeyID) {
		return r.log, nil
	}
	return nil, service.ErrUsageLogNotFound
}

func TestUsageHandlerGetByRequestIDWithClientRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &adminUsageByRequestRepo{
		requestID: "client:abc-123",
		log: &service.UsageLog{
			ID:           99,
			APIKeyID:     100,
			RequestID:    "client:abc-123",
			Model:        "gpt-4o-mini",
			InputTokens:  10,
			OutputTokens: 20,
			ActualCost:   0.001,
			CreatedAt:    time.Now(),
		},
	}
	handler := NewUsageHandler(service.NewUsageService(repo, nil, nil, nil), nil, nil, nil)

	router := gin.New()
	router.GET("/admin/usage/by-request-id", handler.GetByRequestID)

	req := httptest.NewRequest(http.MethodGet, "/admin/usage/by-request-id?client_request_id=abc-123", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	for _, part := range []string{`"request_id":"client:abc-123"`, `"actual_cost":0.001`, `"input_tokens":10`} {
		if !strings.Contains(body, part) {
			t.Fatalf("body missing %q: %s", part, body)
		}
	}
}

func TestUsageHandlerGetByRequestIDWithBareRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &adminUsageByRequestRepo{
		requestID: "client:550e8400-e29b-41d4-a716-446655440000",
		log: &service.UsageLog{
			ID:        42,
			APIKeyID:  1,
			RequestID: "client:550e8400-e29b-41d4-a716-446655440000",
			CreatedAt: time.Now(),
		},
	}
	handler := NewUsageHandler(service.NewUsageService(repo, nil, nil, nil), nil, nil, nil)

	router := gin.New()
	router.GET("/admin/usage/by-request-id", handler.GetByRequestID)

	req := httptest.NewRequest(http.MethodGet, "/admin/usage/by-request-id?request_id=550e8400-e29b-41d4-a716-446655440000", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"request_id":"client:550e8400-e29b-41d4-a716-446655440000"`) {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}

func TestUsageHandlerGetByRequestIDMissingLookupKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewUsageHandler(service.NewUsageService(&adminUsageByRequestRepo{}, nil, nil, nil), nil, nil, nil)

	router := gin.New()
	router.GET("/admin/usage/by-request-id", handler.GetByRequestID)

	req := httptest.NewRequest(http.MethodGet, "/admin/usage/by-request-id", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}
