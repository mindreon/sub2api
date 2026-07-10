package routes

import (
	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterMediaRoutes 注册多模态异步生成任务路由（中转站风格视频接口）。
//
// 对外仅暴露 /v1/video/generations，便于后续接入多方厂商时保持统一契约。
// 鉴权与 /v1/messages 等网关端点一致（API Key）。
func RegisterMediaRoutes(
	r *gin.Engine,
	h *handler.Handlers,
	apiKeyAuth middleware.APIKeyAuthMiddleware,
) {
	videos := r.Group("/v1/video/generations")
	videos.Use(gin.HandlerFunc(apiKeyAuth))
	{
		videos.POST("", h.Media.SubmitCommonVideoGeneration)
		videos.GET("/:task_id", h.Media.GetCommonVideoGeneration)
	}
}
