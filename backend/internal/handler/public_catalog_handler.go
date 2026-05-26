package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

// PublicCatalogHandler serves the public model catalog for gasboard and external consumers.
// Reads from the database — no authentication required.
type PublicCatalogHandler struct {
	catalogService *service.CatalogModelService
}

// NewPublicCatalogHandler creates a PublicCatalogHandler backed by the DB service.
func NewPublicCatalogHandler(catalogService *service.CatalogModelService) *PublicCatalogHandler {
	return &PublicCatalogHandler{catalogService: catalogService}
}

// ListModels handles GET /api/public/models
func (h *PublicCatalogHandler) ListModels(c *gin.Context) {
	models, err := h.catalogService.ListEnabled(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "failed to load models"})
		return
	}

	out := make([]dto.PublicModel, 0, len(models))
	for _, m := range models {
		out = append(out, publicModelFromService(m))
	}
	c.JSON(http.StatusOK, dto.PublicModelsListResponse{
		Success: true,
		Data:    out,
		Total:   len(out),
	})
}

// GetModel handles GET /api/public/models/*id
// Model IDs contain "/" (e.g. "openai/gpt-4o"), so a wildcard param is required.
func (h *PublicCatalogHandler) GetModel(c *gin.Context) {
	rawID := c.Param("id")
	modelID := strings.TrimPrefix(rawID, "/")

	m, err := h.catalogService.GetByModelID(c.Request.Context(), modelID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "model not found"})
		return
	}
	if !m.IsEnabled {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "model not found"})
		return
	}

	c.JSON(http.StatusOK, dto.PublicModelDetailResponse{Success: true, Data: publicModelFromService(*m)})
}

func publicModelFromService(m service.CatalogModel) dto.PublicModel {
	capabilities := make(map[string]bool)
	for _, f := range m.Features {
		capabilities[f] = true
	}
	pricing := map[string]string{}
	if m.InputPrice > 0 {
		pricing["input"] = fmt.Sprintf("%.4f", m.InputPrice)
	}
	if m.OutputPrice > 0 {
		pricing["output"] = fmt.Sprintf("%.4f", m.OutputPrice)
	}
	if m.CacheWritePrice != nil {
		pricing["cache_write"] = fmt.Sprintf("%.4f", *m.CacheWritePrice)
	}
	if m.CacheReadPrice != nil {
		pricing["cache_read"] = fmt.Sprintf("%.4f", *m.CacheReadPrice)
	}
	return dto.PublicModel{
		ID:                     m.ModelID,
		DisplayName:            m.Name,
		Icon:                   m.IconURL,
		Mode:                   m.Category,
		ContextWindow:          int(m.ContextWindow),
		MaxOutputTokens:        int(m.MaxOutputTokens),
		Pricing:                pricing,
		Capabilities:           capabilities,
		SupportedProviderTypes: []string{},
		ReleasedAt:             "",
	}
}
