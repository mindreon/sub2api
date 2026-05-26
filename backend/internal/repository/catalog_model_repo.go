package repository

import (
	"context"
	"fmt"
	"strings"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/catalogmodel"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"

	entsql "entgo.io/ent/dialect/sql"
)

// CatalogModelRepo is exported so test files can reference it for compile-time interface assertion.
type CatalogModelRepo struct {
	client *dbent.Client
}

// NewCatalogModelRepository creates a new CatalogModelRepository backed by ent.
func NewCatalogModelRepository(client *dbent.Client) service.CatalogModelRepository {
	return &CatalogModelRepo{client: client}
}

func (r *CatalogModelRepo) List(
	ctx context.Context,
	params pagination.PaginationParams,
	filters service.CatalogModelListFilters,
) ([]service.CatalogModel, *pagination.PaginationResult, error) {
	q := r.client.CatalogModel.Query()

	if filters.Vendor != "" {
		q = q.Where(catalogmodel.VendorEQ(filters.Vendor))
	}
	if filters.Category != "" {
		q = q.Where(catalogmodel.CategoryEQ(filters.Category))
	}
	if filters.Enabled != nil {
		q = q.Where(catalogmodel.IsEnabledEQ(*filters.Enabled))
	}
	if filters.Search != "" {
		q = q.Where(catalogmodel.Or(
			catalogmodel.ModelIDContainsFold(filters.Search),
			catalogmodel.NameContainsFold(filters.Search),
		))
	}

	total, err := q.Count(ctx)
	if err != nil {
		return nil, nil, err
	}

	itemsQuery := q.
		Offset(params.Offset()).
		Limit(params.Limit())
	for _, order := range catalogModelListOrders(params) {
		itemsQuery = itemsQuery.Order(order)
	}

	items, err := itemsQuery.All(ctx)
	if err != nil {
		return nil, nil, err
	}

	return catalogModelEntitiesToService(items), paginationResultFromTotal(int64(total), params), nil
}

func (r *CatalogModelRepo) GetByID(ctx context.Context, id int64) (*service.CatalogModel, error) {
	m, err := r.client.CatalogModel.Get(ctx, id)
	if err != nil {
		return nil, translatePersistenceError(err, service.ErrCatalogModelNotFound, nil)
	}
	return catalogModelEntityToService(m), nil
}

func (r *CatalogModelRepo) GetByModelID(ctx context.Context, modelID string) (*service.CatalogModel, error) {
	m, err := r.client.CatalogModel.Query().
		Where(catalogmodel.ModelIDEQ(modelID)).
		Only(ctx)
	if err != nil {
		return nil, translatePersistenceError(err, service.ErrCatalogModelNotFound, nil)
	}
	return catalogModelEntityToService(m), nil
}

func (r *CatalogModelRepo) ListEnabled(ctx context.Context) ([]service.CatalogModel, error) {
	items, err := r.client.CatalogModel.Query().
		Where(catalogmodel.IsEnabledEQ(true)).
		Order(dbent.Asc(catalogmodel.FieldModelID)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return catalogModelEntitiesToService(items), nil
}

func (r *CatalogModelRepo) Update(ctx context.Context, id int64, input service.UpdateCatalogModelInput) (*service.CatalogModel, error) {
	builder := r.client.CatalogModel.UpdateOneID(id).
		SetName(input.Name).
		SetVendor(input.Vendor).
		SetCategory(input.Category).
		SetDescription(input.Description).
		SetTags(stringsOrEmpty(input.Tags)).
		SetDocURL(input.DocURL).
		SetIconURL(input.IconURL).
		SetContextWindow(input.ContextWindow).
		SetMaxOutputTokens(input.MaxOutputTokens).
		SetInputModalities(stringsOrEmpty(input.InputModalities)).
		SetOutputModalities(stringsOrEmpty(input.OutputModalities)).
		SetFeatures(stringsOrEmpty(input.Features)).
		SetInputPrice(input.InputPrice).
		SetOutputPrice(input.OutputPrice).
		SetCurrency(input.Currency)

	if input.CacheWritePrice != nil {
		builder = builder.SetCacheWritePrice(*input.CacheWritePrice)
	} else {
		builder = builder.ClearCacheWritePrice()
	}
	if input.CacheReadPrice != nil {
		builder = builder.SetCacheReadPrice(*input.CacheReadPrice)
	} else {
		builder = builder.ClearCacheReadPrice()
	}

	updated, err := builder.Save(ctx)
	if err != nil {
		return nil, translatePersistenceError(err, service.ErrCatalogModelNotFound, nil)
	}
	return catalogModelEntityToService(updated), nil
}

func (r *CatalogModelRepo) Toggle(ctx context.Context, id int64) (*service.CatalogModel, error) {
	m, err := r.client.CatalogModel.Get(ctx, id)
	if err != nil {
		return nil, translatePersistenceError(err, service.ErrCatalogModelNotFound, nil)
	}
	updated, err := m.Update().SetIsEnabled(!m.IsEnabled).Save(ctx)
	if err != nil {
		return nil, err
	}
	return catalogModelEntityToService(updated), nil
}

func (r *CatalogModelRepo) CountAll(ctx context.Context) (int, error) {
	return r.client.CatalogModel.Query().Count(ctx)
}

func (r *CatalogModelRepo) BulkUpsert(ctx context.Context, models []service.CatalogModel) error {
	for _, m := range models {
		tags := stringsOrEmpty(m.Tags)
		inputMod := stringsOrEmpty(m.InputModalities)
		outputMod := stringsOrEmpty(m.OutputModalities)
		features := stringsOrEmpty(m.Features)

		builder := r.client.CatalogModel.Create().
			SetModelID(m.ModelID).
			SetName(m.Name).
			SetVendor(m.Vendor).
			SetCategory(m.Category).
			SetDescription(m.Description).
			SetTags(tags).
			SetDocURL(m.DocURL).
			SetIconURL(m.IconURL).
			SetContextWindow(m.ContextWindow).
			SetMaxOutputTokens(m.MaxOutputTokens).
			SetInputModalities(inputMod).
			SetOutputModalities(outputMod).
			SetFeatures(features).
			SetInputPrice(m.InputPrice).
			SetOutputPrice(m.OutputPrice).
			SetCurrency(m.Currency).
			SetIsEnabled(m.IsEnabled)

		if m.CacheWritePrice != nil {
			builder = builder.SetCacheWritePrice(*m.CacheWritePrice)
		}
		if m.CacheReadPrice != nil {
			builder = builder.SetCacheReadPrice(*m.CacheReadPrice)
		}

		err := builder.
			OnConflictColumns(catalogmodel.FieldModelID).
			UpdateNewValues().
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("upsert %s: %w", m.ModelID, err)
		}
	}
	return nil
}

func catalogModelListOrders(params pagination.PaginationParams) []func(*entsql.Selector) {
	field, sortOrder := catalogModelSortOrder(params)

	if sortOrder == pagination.SortOrderAsc {
		if field == catalogmodel.FieldModelID {
			return []func(*entsql.Selector){
				dbent.Asc(field),
			}
		}
		return []func(*entsql.Selector){
			dbent.Asc(field),
			dbent.Asc(catalogmodel.FieldModelID),
		}
	}

	if field == catalogmodel.FieldModelID {
		return []func(*entsql.Selector){
			dbent.Desc(field),
		}
	}
	return []func(*entsql.Selector){
		dbent.Desc(field),
		dbent.Asc(catalogmodel.FieldModelID),
	}
}

func catalogModelSortOrder(params pagination.PaginationParams) (string, string) {
	sortBy := strings.ToLower(strings.TrimSpace(params.SortBy))
	order := params.NormalizedSortOrder(pagination.SortOrderAsc)
	switch sortBy {
	case "name":
		return catalogmodel.FieldName, order
	case "vendor":
		return catalogmodel.FieldVendor, order
	case "category":
		return catalogmodel.FieldCategory, order
	case "input_price":
		return catalogmodel.FieldInputPrice, order
	case "output_price":
		return catalogmodel.FieldOutputPrice, order
	default:
		return catalogmodel.FieldModelID, pagination.SortOrderAsc
	}
}

func stringsOrEmpty(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

func catalogModelEntityToService(m *dbent.CatalogModel) *service.CatalogModel {
	if m == nil {
		return nil
	}
	return &service.CatalogModel{
		ID:               m.ID,
		ModelID:          m.ModelID,
		Name:             m.Name,
		Vendor:           m.Vendor,
		Category:         m.Category,
		Description:      m.Description,
		Tags:             m.Tags,
		DocURL:           m.DocURL,
		IconURL:          m.IconURL,
		ContextWindow:    m.ContextWindow,
		MaxOutputTokens:  m.MaxOutputTokens,
		InputModalities:  m.InputModalities,
		OutputModalities: m.OutputModalities,
		Features:         m.Features,
		InputPrice:       m.InputPrice,
		OutputPrice:      m.OutputPrice,
		CacheWritePrice:  m.CacheWritePrice,
		CacheReadPrice:   m.CacheReadPrice,
		Currency:         m.Currency,
		IsEnabled:        m.IsEnabled,
		CreatedAt:        m.CreatedAt,
		UpdatedAt:        m.UpdatedAt,
	}
}

func catalogModelEntitiesToService(items []*dbent.CatalogModel) []service.CatalogModel {
	out := make([]service.CatalogModel, 0, len(items))
	for i := range items {
		if s := catalogModelEntityToService(items[i]); s != nil {
			out = append(out, *s)
		}
	}
	return out
}
