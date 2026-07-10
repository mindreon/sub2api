package handler

import (
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/media"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/gin-gonic/gin"
)

// ListTasks 处理 GET /api/v1/media/tasks：当前登录用户的多模态任务列表（JWT）。
func (h *MediaHandler) ListTasks(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	page, pageSize := response.ParsePagination(c)
	userID := subject.UserID
	result, err := h.tasks.List(c.Request.Context(), media.TaskListQuery{
		UserID:    &userID,
		Status:    strings.TrimSpace(c.Query("status")),
		MediaType: strings.TrimSpace(c.Query("media_type")),
		Model:     strings.TrimSpace(c.Query("model")),
		Page:      page,
		PageSize:  pageSize,
		SortBy:    c.DefaultQuery("sort_by", "created_at"),
		SortOrder: c.DefaultQuery("sort_order", "desc"),
	})
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
