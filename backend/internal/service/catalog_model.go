package service

import (
	"context"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

var (
	ErrCatalogModelNotFound = infraerrors.NotFound("CATALOG_MODEL_NOT_FOUND", "catalog model not found")
)

// CatalogModel is the domain model for a public-facing model catalog entry.
type CatalogModel struct {
	ID               int64
	ModelID          string
	Name             string
	Vendor           string
	Category         string
	Description      string
	Tags             []string
	DocURL           string
	IconURL          string
	ContextWindow    int64
	MaxOutputTokens  int64
	InputModalities  []string
	OutputModalities []string
	Features         []string
	InputPrice       float64
	OutputPrice      float64
	CacheWritePrice  *float64
	CacheReadPrice   *float64
	Currency         string
	IsEnabled        bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// CatalogModelListFilters holds optional filter parameters for listing catalog models.
type CatalogModelListFilters struct {
	Search   string // matches model_id or name
	Vendor   string
	Category string
	Enabled  *bool // nil = all, true = enabled only, false = disabled only
}

// UpdateCatalogModelInput holds all editable fields for a catalog model update.
type UpdateCatalogModelInput struct {
	Name             string
	Vendor           string
	Category         string
	Description      string
	Tags             []string
	DocURL           string
	IconURL          string
	ContextWindow    int64
	MaxOutputTokens  int64
	InputModalities  []string
	OutputModalities []string
	Features         []string
	InputPrice       float64
	OutputPrice      float64
	CacheWritePrice  *float64
	CacheReadPrice   *float64
	Currency         string
}

// CatalogModelRepository defines persistence operations for catalog models.
type CatalogModelRepository interface {
	List(ctx context.Context, params pagination.PaginationParams, filters CatalogModelListFilters) ([]CatalogModel, *pagination.PaginationResult, error)
	GetByID(ctx context.Context, id int64) (*CatalogModel, error)
	GetByModelID(ctx context.Context, modelID string) (*CatalogModel, error)
	ListEnabled(ctx context.Context) ([]CatalogModel, error)
	Update(ctx context.Context, id int64, input UpdateCatalogModelInput) (*CatalogModel, error)
	Toggle(ctx context.Context, id int64) (*CatalogModel, error)
	CountAll(ctx context.Context) (int, error)
	BulkUpsert(ctx context.Context, models []CatalogModel) error
}
