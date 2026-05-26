package repository

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestDistributionOrganizationRepositoryCreate_PersistsOrganizationInput(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &distributionOrganizationRepository{db: db}

	ownerID := int64(7)
	createdAt := time.Date(2026, 5, 23, 12, 0, 0, 0, time.UTC)
	input := service.DistributionOrganizationInput{
		Type:        "reseller",
		Name:        "Independent Agent",
		OwnerUserID: &ownerID,
		Status:      "active",
		Config:      map[string]any{"freeze_days": float64(7)},
		BrandConfig: map[string]any{"brand": "agent"},
	}

	mock.ExpectQuery("INSERT INTO channel_organizations").
		WithArgs(
			input.Type,
			input.Name,
			ownerID,
			input.Status,
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"type",
			"name",
			"owner_user_id",
			"status",
			"config",
			"brand_config",
			"created_at",
			"updated_at",
		}).AddRow(
			int64(88),
			input.Type,
			input.Name,
			ownerID,
			input.Status,
			[]byte(`{"freeze_days":7}`),
			[]byte(`{"brand":"agent"}`),
			createdAt,
			createdAt,
		))

	out, err := repo.Create(context.Background(), input)
	require.NoError(t, err)
	require.Equal(t, int64(88), out.ID)
	require.Equal(t, "reseller", out.Type)
	require.Equal(t, float64(7), out.Config["freeze_days"])
	require.Equal(t, "agent", out.BrandConfig["brand"])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDistributionOrganizationRepositoryList_ReturnsPaginatedOrganizations(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &distributionOrganizationRepository{db: db}

	ownerID := int64(7)
	createdAt := time.Date(2026, 5, 23, 12, 0, 0, 0, time.UTC)
	mock.ExpectQuery("SELECT COUNT\\(\\*\\).*FROM channel_organizations").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))
	mock.ExpectQuery("FROM channel_organizations").
		WithArgs(20, 20).
		WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"type",
			"name",
			"owner_user_id",
			"status",
			"config",
			"brand_config",
			"created_at",
			"updated_at",
		}).AddRow(
			int64(88),
			"reseller",
			"Independent Agent",
			ownerID,
			"active",
			[]byte(`{}`),
			[]byte(`{}`),
			createdAt,
			createdAt,
		))

	out, page, err := repo.List(context.Background(), pagination.PaginationParams{Page: 2, PageSize: 20})
	require.NoError(t, err)
	require.Len(t, out, 1)
	require.Equal(t, int64(1), page.Total)
	require.Equal(t, int64(88), out[0].ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDistributionOrganizationRepositoryGetByID_ReturnsOrganization(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &distributionOrganizationRepository{db: db}

	createdAt := time.Date(2026, 5, 23, 12, 0, 0, 0, time.UTC)
	mock.ExpectQuery("FROM channel_organizations").
		WithArgs(int64(88)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"type",
			"name",
			"owner_user_id",
			"status",
			"config",
			"brand_config",
			"created_at",
			"updated_at",
		}).AddRow(
			int64(88),
			"reseller",
			"Independent Agent",
			int64(7),
			"active",
			[]byte(`{"commission_upper_ratio":0.35}`),
			[]byte(`{"brand":"agent"}`),
			createdAt,
			createdAt,
		))

	out, err := repo.GetByID(context.Background(), 88)
	require.NoError(t, err)
	require.Equal(t, int64(88), out.ID)
	require.Equal(t, 0.35, out.Config["commission_upper_ratio"])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDistributionOrganizationRepositoryUpdate_PersistsInput(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &distributionOrganizationRepository{db: db}

	ownerID := int64(7)
	createdAt := time.Date(2026, 5, 23, 12, 0, 0, 0, time.UTC)
	input := service.DistributionOrganizationInput{
		Type:        "reseller",
		Name:        "Independent Agent",
		OwnerUserID: &ownerID,
		Status:      "active",
		Config:      map[string]any{"commission_upper_ratio": 0.35},
		BrandConfig: map[string]any{"brand": "agent"},
	}

	mock.ExpectQuery("UPDATE channel_organizations").
		WithArgs(
			int64(88),
			input.Type,
			input.Name,
			ownerID,
			input.Status,
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"type",
			"name",
			"owner_user_id",
			"status",
			"config",
			"brand_config",
			"created_at",
			"updated_at",
		}).AddRow(
			int64(88),
			input.Type,
			input.Name,
			ownerID,
			input.Status,
			[]byte(`{"commission_upper_ratio":0.35}`),
			[]byte(`{"brand":"agent"}`),
			createdAt,
			createdAt,
		))

	out, err := repo.Update(context.Background(), 88, input)
	require.NoError(t, err)
	require.Equal(t, int64(88), out.ID)
	require.Equal(t, 0.35, out.Config["commission_upper_ratio"])
	require.NoError(t, mock.ExpectationsWereMet())
}
