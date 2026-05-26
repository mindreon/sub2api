package admin

import (
	"strconv"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// CatalogModelHandler handles admin model catalog management.
type CatalogModelHandler struct {
	catalogService *service.CatalogModelService
}

// NewCatalogModelHandler creates a new CatalogModelHandler.
func NewCatalogModelHandler(catalogService *service.CatalogModelService) *CatalogModelHandler {
	return &CatalogModelHandler{catalogService: catalogService}
}

// List handles GET /api/v1/admin/catalog/models
// Query params: vendor, category, enabled (true/false), q (search), page, page_size, sort_by, sort_order
func (h *CatalogModelHandler) List(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	params := pagination.PaginationParams{
		Page:      page,
		PageSize:  pageSize,
		SortBy:    c.DefaultQuery("sort_by", "model_id"),
		SortOrder: c.DefaultQuery("sort_order", "asc"),
	}

	filters := service.CatalogModelListFilters{
		Vendor:   strings.TrimSpace(c.Query("vendor")),
		Category: strings.TrimSpace(c.Query("category")),
		Search:   strings.TrimSpace(c.Query("q")),
	}
	if raw := c.Query("enabled"); raw == "true" {
		t := true
		filters.Enabled = &t
	} else if raw == "false" {
		f := false
		filters.Enabled = &f
	}

	items, result, err := h.catalogService.List(c.Request.Context(), params, filters)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	out := make([]dto.AdminCatalogModel, 0, len(items))
	for i := range items {
		if d := dto.AdminCatalogModelFromService(&items[i]); d != nil {
			out = append(out, *d)
		}
	}
	response.Paginated(c, out, result.Total, page, pageSize)
}

// Update handles PUT /api/v1/admin/catalog/models/:id
func (h *CatalogModelHandler) Update(c *gin.Context) {
	id, err := parseCatalogID(c)
	if err != nil {
		response.BadRequest(c, "invalid model ID")
		return
	}

	var req dto.UpdateCatalogModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	input := service.UpdateCatalogModelInput{
		Name:             req.Name,
		Vendor:           req.Vendor,
		Category:         req.Category,
		Description:      req.Description,
		Tags:             req.Tags,
		DocURL:           req.DocURL,
		IconURL:          req.IconURL,
		ContextWindow:    req.ContextWindow,
		MaxOutputTokens:  req.MaxOutputTokens,
		InputModalities:  req.InputModalities,
		OutputModalities: req.OutputModalities,
		Features:         req.Features,
		InputPrice:       req.InputPrice,
		OutputPrice:      req.OutputPrice,
		CacheWritePrice:  req.CacheWritePrice,
		CacheReadPrice:   req.CacheReadPrice,
		Currency:         req.Currency,
	}

	updated, err := h.catalogService.Update(c.Request.Context(), id, input)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, dto.AdminCatalogModelFromService(updated))
}

// Toggle handles PATCH /api/v1/admin/catalog/models/:id/toggle
func (h *CatalogModelHandler) Toggle(c *gin.Context) {
	id, err := parseCatalogID(c)
	if err != nil {
		response.BadRequest(c, "invalid model ID")
		return
	}
	updated, err := h.catalogService.Toggle(c.Request.Context(), id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, dto.AdminCatalogModelFromService(updated))
}

// Seed handles POST /api/v1/admin/catalog/models/seed
func (h *CatalogModelHandler) Seed(c *gin.Context) {
	n, err := h.catalogService.Reseed(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"seeded": n})
}

func parseCatalogID(c *gin.Context) (int64, error) {
	return strconv.ParseInt(c.Param("id"), 10, 64)
}
