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

func TestDistributionPromotionRepositoryCreate_PersistsPromotionLink(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &distributionPromotionRepository{db: db}

	createdAt := time.Date(2026, 5, 23, 12, 0, 0, 0, time.UTC)
	input := service.DistributionPromotionLinkInput{
		MemberID:    11,
		Code:        "LINK-123",
		TargetType:  "registration",
		Status:      "active",
	}

	mock.ExpectQuery("INSERT INTO promotion_links").
		WithArgs(
			input.MemberID,
			input.Code,
			input.TargetType,
			input.Status,
		).
		WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"channel_org_id",
			"member_id",
			"code",
			"target_type",
			"status",
			"user_id",
			"email",
			"username",
			"role_type",
			"created_at",
			"updated_at",
		}).AddRow(
			int64(91),
			int64(88),
			input.MemberID,
			input.Code,
			input.TargetType,
			input.Status,
			int64(7),
			"agent@example.com",
			"agent",
			"agent",
			createdAt,
			createdAt,
		))

	out, err := repo.Create(context.Background(), input)
	require.NoError(t, err)
	require.Equal(t, int64(91), out.ID)
	require.Equal(t, int64(88), out.ChannelOrgID)
	require.Equal(t, input.Code, out.Code)
	require.Equal(t, "agent@example.com", out.UserEmail)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDistributionPromotionRepositoryGetByCode_LoadsPromotionLink(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &distributionPromotionRepository{db: db}

	createdAt := time.Date(2026, 5, 23, 12, 0, 0, 0, time.UTC)
	mock.ExpectQuery("FROM promotion_links pl").
		WithArgs("LINK-123").
		WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"channel_org_id",
			"member_id",
			"code",
			"target_type",
			"status",
			"user_id",
			"email",
			"username",
			"role_type",
			"created_at",
			"updated_at",
		}).AddRow(
			int64(91),
			int64(88),
			int64(11),
			"LINK-123",
			"registration",
			"active",
			int64(7),
			"agent@example.com",
			"agent",
			"agent",
			createdAt,
			createdAt,
		))

	out, err := repo.GetByCode(context.Background(), "LINK-123")
	require.NoError(t, err)
	require.Equal(t, int64(91), out.ID)
	require.Equal(t, "LINK-123", out.Code)
	require.Equal(t, "agent@example.com", out.UserEmail)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDistributionPromotionRepositoryListByChannelOrgID_ReturnsPromotionLinks(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &distributionPromotionRepository{db: db}

	createdAt := time.Date(2026, 5, 23, 12, 0, 0, 0, time.UTC)
	mock.ExpectQuery("SELECT COUNT\\(\\*\\).*FROM promotion_links").
		WithArgs(int64(88)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))
	mock.ExpectQuery("FROM promotion_links").
		WithArgs(int64(88), 20, 20).
		WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"channel_org_id",
			"member_id",
			"code",
			"target_type",
			"status",
			"user_id",
			"email",
			"username",
			"role_type",
			"created_at",
			"updated_at",
		}).AddRow(
			int64(91),
			int64(88),
			int64(11),
			"LINK-123",
			"registration",
			"active",
			int64(7),
			"agent@example.com",
			"agent",
			"agent",
			createdAt,
			createdAt,
		))

	out, page, err := repo.ListByChannelOrgID(context.Background(), 88, pagination.PaginationParams{Page: 2, PageSize: 20})
	require.NoError(t, err)
	require.Len(t, out, 1)
	require.Equal(t, int64(1), page.Total)
	require.Equal(t, "LINK-123", out[0].Code)
	require.Equal(t, "agent@example.com", out[0].UserEmail)
	require.NoError(t, mock.ExpectationsWereMet())
}
