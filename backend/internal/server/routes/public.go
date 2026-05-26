package routes

import (
	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/gin-gonic/gin"
)

// RegisterPublicRoutes registers unauthenticated endpoints consumed by gasboard and external clients.
// Route namespace: /api/public — distinct from /api/v1 and /v1.
func RegisterPublicRoutes(r *gin.Engine, h *handler.Handlers) {
	public := r.Group("/api/public")
	{
		public.GET("/models", h.PublicCatalog.ListModels)
		public.GET("/models/*id", h.PublicCatalog.GetModel)
	}
}
