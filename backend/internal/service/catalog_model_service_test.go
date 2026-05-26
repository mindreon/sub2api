package service_test

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

// mockCatalogRepo implements service.CatalogModelRepository for unit tests.
type mockCatalogRepo struct {
	items []service.CatalogModel
	count int
}

func (m *mockCatalogRepo) List(_ context.Context, _ pagination.PaginationParams, _ service.CatalogModelListFilters) ([]service.CatalogModel, *pagination.PaginationResult, error) {
	return m.items, &pagination.PaginationResult{Total: int64(len(m.items))}, nil
}

func (m *mockCatalogRepo) GetByID(_ context.Context, id int64) (*service.CatalogModel, error) {
	for _, it := range m.items {
		if it.ID == id {
			cp := it
			return &cp, nil
		}
	}
	return nil, service.ErrCatalogModelNotFound
}

func (m *mockCatalogRepo) GetByModelID(_ context.Context, mid string) (*service.CatalogModel, error) {
	for _, it := range m.items {
		if it.ModelID == mid {
			cp := it
			return &cp, nil
		}
	}
	return nil, service.ErrCatalogModelNotFound
}

func (m *mockCatalogRepo) ListEnabled(_ context.Context) ([]service.CatalogModel, error) {
	var out []service.CatalogModel
	for _, it := range m.items {
		if it.IsEnabled {
			out = append(out, it)
		}
	}
	return out, nil
}

func (m *mockCatalogRepo) Update(_ context.Context, id int64, input service.UpdateCatalogModelInput) (*service.CatalogModel, error) {
	for i, it := range m.items {
		if it.ID == id {
			m.items[i].Name = input.Name
			cp := m.items[i]
			return &cp, nil
		}
	}
	return nil, service.ErrCatalogModelNotFound
}

func (m *mockCatalogRepo) Toggle(_ context.Context, id int64) (*service.CatalogModel, error) {
	for i, it := range m.items {
		if it.ID == id {
			m.items[i].IsEnabled = !it.IsEnabled
			cp := m.items[i]
			return &cp, nil
		}
	}
	return nil, service.ErrCatalogModelNotFound
}

func (m *mockCatalogRepo) CountAll(_ context.Context) (int, error) { return m.count, nil }

func (m *mockCatalogRepo) BulkUpsert(_ context.Context, models []service.CatalogModel) error {
	m.items = append(m.items, models...)
	return nil
}

func TestCatalogModelServiceSeedIfEmptySkipsWhenNotEmpty(t *testing.T) {
	repo := &mockCatalogRepo{count: 5}
	svc := service.NewCatalogModelService(repo)
	if err := svc.SeedIfEmpty(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repo.items) != 0 {
		t.Errorf("expected no items inserted when count>0, got %d", len(repo.items))
	}
}

func TestCatalogModelServiceSeedIfEmptyInsertsWhenEmpty(t *testing.T) {
	repo := &mockCatalogRepo{count: 0}
	svc := service.NewCatalogModelService(repo)
	if err := svc.SeedIfEmpty(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repo.items) == 0 {
		t.Error("expected items to be seeded when DB is empty")
	}
	// Verify that seeded models have expected fields populated
	first := repo.items[0]
	if first.ModelID == "" {
		t.Error("expected ModelID to be non-empty")
	}
	if first.Vendor == "" {
		t.Error("expected Vendor to be non-empty")
	}
	if !first.IsEnabled {
		t.Error("expected seeded models to be enabled by default")
	}
}

func TestCatalogModelServiceReseed(t *testing.T) {
	repo := &mockCatalogRepo{count: 5, items: []service.CatalogModel{}}
	svc := service.NewCatalogModelService(repo)
	n, err := svc.Reseed(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n == 0 {
		t.Error("expected Reseed to return non-zero count")
	}
	if len(repo.items) != n {
		t.Errorf("expected %d items inserted, got %d", n, len(repo.items))
	}
}

func TestCatalogModelServiceGetByIDNotFound(t *testing.T) {
	repo := &mockCatalogRepo{}
	svc := service.NewCatalogModelService(repo)
	_, err := svc.GetByID(context.Background(), 999)
	if err == nil {
		t.Error("expected error for missing ID")
	}
}

func TestCatalogModelServiceToggle(t *testing.T) {
	repo := &mockCatalogRepo{
		items: []service.CatalogModel{
			{ID: 1, ModelID: "test/model", IsEnabled: true},
		},
	}
	svc := service.NewCatalogModelService(repo)
	updated, err := svc.Toggle(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.IsEnabled {
		t.Error("expected IsEnabled to be flipped to false")
	}
}
