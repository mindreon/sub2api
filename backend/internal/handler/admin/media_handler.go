package admin

import (
	"errors"
	"strconv"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/media"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/gin-gonic/gin"
)

// MediaHandler 管理端多模态计费：任务列表与汇率配置。
type MediaHandler struct {
	tasks   media.TaskStore
	configs *media.ConfigStore
}

// NewMediaHandler 构造管理端多模态 handler。
func NewMediaHandler(tasks media.TaskStore, configs *media.ConfigStore) *MediaHandler {
	return &MediaHandler{tasks: tasks, configs: configs}
}

// GetSettings GET /api/v1/admin/media/settings
func (h *MediaHandler) GetSettings(c *gin.Context) {
	if h.configs == nil {
		response.InternalError(c, "media config store not configured")
		return
	}
	cfg, err := h.configs.Load(c.Request.Context())
	if err != nil {
		response.InternalError(c, "failed to load media billing settings")
		return
	}
	response.Success(c, cfg.PublicView())
}

// UpdateSettings PUT /api/v1/admin/media/settings
func (h *MediaHandler) UpdateSettings(c *gin.Context) {
	if h.configs == nil {
		response.InternalError(c, "media config store not configured")
		return
	}
	var req media.BillingConfigUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}
	current, err := h.configs.Load(c.Request.Context())
	if err != nil {
		response.InternalError(c, "failed to load current settings")
		return
	}
	updated := current.ApplyUpdate(req)
	if req.CNYToUSDRate != nil && updated.CNYToUSDRate <= 0 {
		response.BadRequest(c, "cny_to_usd_rate must be positive")
		return
	}
	if err := h.configs.Save(c.Request.Context(), updated); err != nil {
		if errors.Is(err, media.ErrInvalidPricingOverride) || errors.Is(err, media.ErrCurrencyRateMissing) {
			response.BadRequest(c, err.Error())
			return
		}
		response.InternalError(c, "failed to save media billing settings")
		return
	}
	response.Success(c, updated.PublicView())
}

// ListTasks GET /api/v1/admin/media/tasks
func (h *MediaHandler) ListTasks(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	q := media.TaskListQuery{
		Status:    strings.TrimSpace(c.Query("status")),
		MediaType: strings.TrimSpace(c.Query("media_type")),
		Model:     strings.TrimSpace(c.Query("model")),
		Page:      page,
		PageSize:  pageSize,
		SortBy:    c.DefaultQuery("sort_by", "created_at"),
		SortOrder: c.DefaultQuery("sort_order", "desc"),
	}
	if userIDStr := strings.TrimSpace(c.Query("user_id")); userIDStr != "" {
		userID, err := strconv.ParseInt(userIDStr, 10, 64)
		if err != nil || userID <= 0 {
			response.BadRequest(c, "invalid user_id")
			return
		}
		q.UserID = &userID
	}
	result, err := h.tasks.List(c.Request.Context(), q)
	if err != nil {
		response.InternalError(c, "failed to list media tasks")
		return
	}
	out := make([]media.TaskView, 0, len(result.Tasks))
	for _, t := range result.Tasks {
		out = append(out, media.ToTaskView(t))
	}
	response.Paginated(c, out, int64(result.Total), page, pageSize)
}
