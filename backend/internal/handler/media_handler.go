package handler

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/media"
	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// MediaHandler 提供中转站风格的视频异步生成 API：提交任务、查询状态。
//
// 对外仅暴露 /v1/video/generations（prompt + metadata）。
// 鉴权与上游网关一致，走 API Key（而非 JWT 用户会话）。
type MediaHandler struct {
	taskService *media.TaskService
	tasks       media.TaskStore
	assets      media.AssetStore // 可选：无公共域名时按需为结果对象重新签发链接
}

// NewMediaHandler 构造多模态任务 handler。
func NewMediaHandler(taskService *media.TaskService, tasks media.TaskStore, assets media.AssetStore) *MediaHandler {
	return &MediaHandler{taskService: taskService, tasks: tasks, assets: assets}
}

// defaultMediaTaskTTL 是未显式指定过期时间时的预扣有效期（超时未完成则轮询释放）。
const defaultMediaTaskTTL = 30 * time.Minute

type submitCommonVideoGenerationRequest struct {
	TaskID   string         `json:"task_id"` // 客户端幂等键，留空自动生成
	Model    string         `json:"model" binding:"required"`
	Prompt   string         `json:"prompt" binding:"required"`
	Metadata map[string]any `json:"metadata"`
}

// SubmitCommonVideoGeneration 处理 POST /v1/video/generations：
// 中转站风格的 prompt + metadata 调用形态。
func (h *MediaHandler) SubmitCommonVideoGeneration(c *gin.Context) {
	apiKey, ok := middleware.GetAPIKeyFromContext(c)
	if !ok || apiKey == nil {
		writeCommonStyleMediaError(c, http.StatusUnauthorized, "invalid_api_key", "missing api key context")
		return
	}
	if !service.AllowsMediaGeneration(apiKey.Group) {
		writeCommonStyleMediaError(c, http.StatusForbidden, "permission_denied", "media generation is not enabled for this API key group")
		return
	}

	var req submitCommonVideoGenerationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeCommonStyleMediaError(c, http.StatusBadRequest, "invalid_request_error", "invalid request: "+err.Error())
		return
	}

	requestParams := commonVideoRequestParams(req)
	usage := commonVideoBillingUsage(requestParams)

	effectiveMultiplier := 1.0
	if apiKey.Group != nil && apiKey.Group.RateMultiplier > 0 {
		effectiveMultiplier = apiKey.Group.RateMultiplier
	}
	rateMultiplier := service.ResolveMediaRateMultiplier(apiKey, effectiveMultiplier)

	var groupID *int64
	if apiKey.GroupID != nil {
		groupID = apiKey.GroupID
	}

	task, err := h.taskService.Submit(c.Request.Context(), media.ReserveInput{
		TaskID:         req.TaskID,
		UserID:         apiKey.UserID,
		APIKeyID:       apiKey.ID,
		GroupID:        groupID,
		Model:          req.Model,
		MediaType:      "video",
		RateMultiplier: rateMultiplier,
		RequestParams:  requestParams,
		ExpiresAt:      time.Now().Add(defaultMediaTaskTTL),
		Usage:          usage,
	})
	if err != nil {
		writeCommonStyleTaskError(c, err)
		return
	}
	c.JSON(http.StatusAccepted, commonVideoSubmitResponse(task))
}

// GetCommonVideoGeneration 处理 GET /v1/video/generations/:task_id。
func (h *MediaHandler) GetCommonVideoGeneration(c *gin.Context) {
	apiKey, ok := middleware.GetAPIKeyFromContext(c)
	if !ok || apiKey == nil {
		writeCommonStyleMediaError(c, http.StatusUnauthorized, "invalid_api_key", "missing api key context")
		return
	}

	taskID := c.Param("task_id")
	task, err := h.tasks.GetByTaskID(c.Request.Context(), taskID)
	if err != nil {
		writeCommonStyleTaskError(c, err)
		return
	}
	if task.UserID != apiKey.UserID {
		writeCommonStyleMediaError(c, http.StatusNotFound, "task_not_found", "task not found")
		return
	}
	if h.assets != nil && task.ResultStorageKey != "" {
		if fresh, err := h.assets.PresignedURL(c.Request.Context(), task.ResultStorageKey); err == nil && fresh != "" {
			task.ResultURL = fresh
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    "success",
		"message": "",
		"data":    commonVideoTaskResponse(task),
	})
}

func writeCommonStyleTaskError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, media.ErrTaskNotFound):
		writeCommonStyleMediaError(c, http.StatusNotFound, "task_not_found", "task not found")
	case errors.Is(err, media.ErrInsufficientBalance):
		writeCommonStyleMediaError(c, http.StatusPaymentRequired, "insufficient_balance", "insufficient balance")
	case errors.Is(err, media.ErrModelNotPriceable), errors.Is(err, media.ErrProviderNotFound):
		writeCommonStyleMediaError(c, http.StatusBadRequest, "model_not_found", "model is not supported for media generation")
	case errors.Is(err, media.ErrCurrencyRateMissing):
		writeCommonStyleMediaError(c, http.StatusInternalServerError, "billing_error", "billing currency configuration error")
	case errors.Is(err, media.ErrUpstreamRequest):
		logger.FromContext(c.Request.Context()).Error("media upstream request failed", zap.Error(err))
		writeCommonStyleMediaError(c, http.StatusBadGateway, "upstream_error", "upstream request failed")
	default:
		logger.FromContext(c.Request.Context()).Error("media task request failed", zap.Error(err))
		writeCommonStyleMediaError(c, http.StatusInternalServerError, "internal_error", "internal error")
	}
}

func writeCommonStyleMediaError(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{
		"error": gin.H{
			"code":    code,
			"message": message,
			"type":    "new_api_error",
		},
	})
}

func commonVideoRequestParams(req submitCommonVideoGenerationRequest) map[string]any {
	params := cloneMetadata(req.Metadata)
	params["prompt"] = strings.TrimSpace(req.Prompt)
	if resolution, ok := stringParam(params["resolution"]); ok {
		params["resolution"] = strings.ToLower(resolution)
	}
	if ratio, ok := stringParam(params["ratio"]); ok {
		params["ratio"] = ratio
	}
	hasVideo, hasAudio := inferCommonVideoContentFlags(params["content"])
	if hasVideo {
		params["has_video_input"] = true
	}
	if hasAudio {
		params["has_audio"] = true
	}
	if generateAudio, ok := params["generate_audio"].(bool); ok && generateAudio {
		params["has_audio"] = true
	}
	return params
}

func commonVideoBillingUsage(params map[string]any) media.BillingUsage {
	usage := media.BillingUsage{}
	if resolution, ok := stringParam(params["resolution"]); ok {
		usage.Resolution = resolution
	}
	if aspectRatio, ok := stringParam(params["aspect_ratio"]); ok {
		usage.AspectRatio = aspectRatio
	}
	if ratio, ok := stringParam(params["ratio"]); ok && usage.AspectRatio == "" {
		usage.AspectRatio = ratio
	}
	if seconds, ok := floatParam(params["video_output_seconds"]); ok {
		usage.VideoOutputSeconds = seconds
	}
	if duration, ok := floatParam(params["duration"]); ok && usage.VideoOutputSeconds == 0 {
		usage.VideoOutputSeconds = duration
	}
	if hasVideo, ok := boolParam(params["has_video_input"]); ok {
		usage.HasVideoInput = hasVideo
	}
	if hasAudio, ok := boolParam(params["has_audio"]); ok {
		usage.HasAudio = hasAudio
	}
	if generateAudio, ok := boolParam(params["generate_audio"]); ok && generateAudio {
		usage.HasAudio = true
	}
	return usage
}

func cloneMetadata(in map[string]any) map[string]any {
	out := make(map[string]any, len(in)+1)
	for k, v := range in {
		out[k] = v
	}
	return out
}

func inferCommonVideoContentFlags(content any) (hasVideo bool, hasAudio bool) {
	items, ok := content.([]any)
	if !ok {
		return false, false
	}
	for _, item := range items {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if typ, ok := stringParam(m["type"]); ok {
			switch strings.ToLower(typ) {
			case "video_url":
				hasVideo = true
			case "audio_url":
				hasAudio = true
			}
		}
		if _, ok := m["video_url"]; ok {
			hasVideo = true
		}
		if _, ok := m["audio_url"]; ok {
			hasAudio = true
		}
	}
	return hasVideo, hasAudio
}

func commonVideoSubmitResponse(task *media.Task) gin.H {
	if task == nil {
		return gin.H{}
	}
	return gin.H{
		"id":         task.TaskID,
		"task_id":    task.TaskID,
		"object":     "video",
		"model":      task.Model,
		"status":     "queued",
		"progress":   0,
		"created_at": unixTime(task.CreatedAt),
	}
}

func commonVideoTaskResponse(task *media.Task) gin.H {
	status, progress := commonVideoStatusProgress(task.Status)
	out := gin.H{
		"id":          task.ID,
		"created_at":  unixTime(task.CreatedAt),
		"updated_at":  unixTime(task.UpdatedAt),
		"task_id":     task.TaskID,
		"user_id":     task.UserID,
		"model":       task.Model,
		"quota":       task.ReservedCost,
		"action":      "generate",
		"status":      status,
		"fail_reason": task.ErrorMessage,
		"submit_time": unixTime(task.CreatedAt),
		"finish_time": unixPtrTime(task.SettledAt),
		"progress":    progress,
		"properties": gin.H{
			"origin_model_name": task.Model,
		},
		"data": task.UpstreamUsage,
	}
	if task.ActualCost != nil {
		out["actual_cost"] = *task.ActualCost
	}
	if task.ResultURL != "" {
		out["result_url"] = task.ResultURL
	}
	if task.Status == media.TaskInProgress {
		out["start_time"] = unixTime(task.UpdatedAt)
	} else {
		out["start_time"] = unixTime(task.CreatedAt)
	}
	if out["data"] == nil {
		out["data"] = gin.H{}
	}
	return out
}

func commonVideoStatusProgress(status media.TaskStatus) (string, string) {
	switch status {
	case media.TaskCompleted:
		return "SUCCESS", "100%"
	case media.TaskFailed, media.TaskExpired, media.TaskCancelled:
		return "FAILED", "100%"
	case media.TaskInProgress:
		return "IN_PROGRESS", "50%"
	default:
		return "IN_PROGRESS", "0%"
	}
}

func unixTime(t time.Time) int64 {
	if t.IsZero() {
		return 0
	}
	return t.Unix()
}

func unixPtrTime(t *time.Time) int64 {
	if t == nil {
		return 0
	}
	return unixTime(*t)
}

func stringParam(v any) (string, bool) {
	s, ok := v.(string)
	if !ok {
		return "", false
	}
	s = strings.TrimSpace(s)
	return s, s != ""
}

func floatParam(v any) (float64, bool) {
	switch t := v.(type) {
	case float64:
		return t, true
	case float32:
		return float64(t), true
	case int:
		return float64(t), true
	case int64:
		return float64(t), true
	default:
		return 0, false
	}
}

func boolParam(v any) (bool, bool) {
	b, ok := v.(bool)
	return b, ok
}
