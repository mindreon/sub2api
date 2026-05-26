package dto

import "github.com/Wei-Shaw/sub2api/internal/service"

// AdminCatalogModel is the admin-facing DTO for a catalog model entry.
type AdminCatalogModel struct {
	ID               int64    `json:"id"`
	ModelID          string   `json:"model_id"`
	Name             string   `json:"name"`
	Vendor           string   `json:"vendor"`
	Category         string   `json:"category"`
	Description      string   `json:"description"`
	Tags             []string `json:"tags"`
	DocURL           string   `json:"doc_url"`
	IconURL          string   `json:"icon_url"`
	ContextWindow    int64    `json:"context_window"`
	MaxOutputTokens  int64    `json:"max_output_tokens"`
	InputModalities  []string `json:"input_modalities"`
	OutputModalities []string `json:"output_modalities"`
	Features         []string `json:"features"`
	InputPrice       float64  `json:"input_price"`
	OutputPrice      float64  `json:"output_price"`
	CacheWritePrice  *float64 `json:"cache_write_price"`
	CacheReadPrice   *float64 `json:"cache_read_price"`
	Currency         string   `json:"currency"`
	IsEnabled        bool     `json:"is_enabled"`
	CreatedAt        int64    `json:"created_at"`
	UpdatedAt        int64    `json:"updated_at"`
}

// AdminCatalogModelFromService converts a service domain model to the admin DTO.
func AdminCatalogModelFromService(m *service.CatalogModel) *AdminCatalogModel {
	if m == nil {
		return nil
	}
	return &AdminCatalogModel{
		ID:               m.ID,
		ModelID:          m.ModelID,
		Name:             m.Name,
		Vendor:           m.Vendor,
		Category:         m.Category,
		Description:      m.Description,
		Tags:             orEmptySlice(m.Tags),
		DocURL:           m.DocURL,
		IconURL:          m.IconURL,
		ContextWindow:    m.ContextWindow,
		MaxOutputTokens:  m.MaxOutputTokens,
		InputModalities:  orEmptySlice(m.InputModalities),
		OutputModalities: orEmptySlice(m.OutputModalities),
		Features:         orEmptySlice(m.Features),
		InputPrice:       m.InputPrice,
		OutputPrice:      m.OutputPrice,
		CacheWritePrice:  m.CacheWritePrice,
		CacheReadPrice:   m.CacheReadPrice,
		Currency:         m.Currency,
		IsEnabled:        m.IsEnabled,
		CreatedAt:        m.CreatedAt.Unix(),
		UpdatedAt:        m.UpdatedAt.Unix(),
	}
}

func orEmptySlice(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

// UpdateCatalogModelRequest is the JSON body for PUT /api/v1/admin/catalog/models/:id
type UpdateCatalogModelRequest struct {
	Name             string   `json:"name" binding:"required"`
	Vendor           string   `json:"vendor"`
	Category         string   `json:"category" binding:"omitempty,oneof=chat embedding image audio video"`
	Description      string   `json:"description"`
	Tags             []string `json:"tags"`
	DocURL           string   `json:"doc_url"`
	IconURL          string   `json:"icon_url"`
	ContextWindow    int64    `json:"context_window" binding:"min=0"`
	MaxOutputTokens  int64    `json:"max_output_tokens" binding:"min=0"`
	InputModalities  []string `json:"input_modalities"`
	OutputModalities []string `json:"output_modalities"`
	Features         []string `json:"features"`
	InputPrice       float64  `json:"input_price" binding:"min=0"`
	OutputPrice      float64  `json:"output_price" binding:"min=0"`
	CacheWritePrice  *float64 `json:"cache_write_price"`
	CacheReadPrice   *float64 `json:"cache_read_price"`
	Currency         string   `json:"currency"`
}
