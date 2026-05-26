package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestDistributionMemberRepositoryGetByID_LoadsCommissionRate(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &distributionMemberRepository{db: db}

	createdAt := time.Date(2026, 5, 23, 12, 0, 0, 0, time.UTC)
	mock.ExpectQuery("FROM channel_members m").
		WithArgs(int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"user_id",
			"email",
			"username",
			"channel_org_id",
			"role_type",
			"parent_member_id",
			"level_code",
			"commission_rate",
			"status",
			"created_at",
			"updated_at",
		}).AddRow(
			int64(42),
			int64(7),
			"agent@example.com",
			"agent",
			int64(99),
			"agent",
			sql.NullInt64{},
			"A",
			0.15,
			"active",
			createdAt,
			createdAt,
		))

	out, err := repo.GetByID(context.Background(), 42)
	require.NoError(t, err)
	require.Equal(t, int64(42), out.MemberID)
	require.Equal(t, int64(99), out.ChannelOrgID)
	require.Equal(t, 0.15, out.CommissionRate)
	require.Equal(t, "active", out.Status)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDistributionMemberRepositoryListByUserID_LoadsExistingChannels(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &distributionMemberRepository{db: db}

	createdAt := time.Date(2026, 5, 23, 12, 0, 0, 0, time.UTC)
	mock.ExpectQuery("WHERE m.user_id = \\$1").
		WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"user_id",
			"email",
			"username",
			"channel_org_id",
			"role_type",
			"parent_member_id",
			"level_code",
			"commission_rate",
			"status",
			"created_at",
			"updated_at",
		}).AddRow(
			int64(42),
			int64(7),
			"agent@example.com",
			"agent",
			int64(99),
			"agent",
			sql.NullInt64{},
			"agent",
			0.15,
			"active",
			createdAt,
			createdAt,
		))

	out, err := repo.ListByUserID(context.Background(), 7)
	require.NoError(t, err)
	require.Len(t, out, 1)
	require.Equal(t, int64(99), out[0].ChannelOrgID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDistributionMemberRepositoryCreate_PersistsMemberInput(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &distributionMemberRepository{db: db}

	parentID := int64(20)
	createdAt := time.Date(2026, 5, 23, 12, 0, 0, 0, time.UTC)
	input := service.DistributionMemberInput{
		ChannelOrgID:   99,
		UserID:         7,
		RoleType:       "kol1",
		ParentMemberID: &parentID,
		LevelCode:      "agent/kol1",
		CommissionRate: 0.10,
		Status:         "active",
	}

	mock.ExpectQuery("INSERT INTO channel_members").
		WithArgs(
			input.ChannelOrgID,
			input.UserID,
			input.RoleType,
			parentID,
			input.LevelCode,
			input.CommissionRate,
			input.Status,
		).
		WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"user_id",
			"email",
			"username",
			"channel_org_id",
			"role_type",
			"parent_member_id",
			"level_code",
			"commission_rate",
			"status",
			"created_at",
			"updated_at",
		}).AddRow(
			int64(43),
			input.UserID,
			"kol@example.com",
			"kol",
			input.ChannelOrgID,
			input.RoleType,
			parentID,
			input.LevelCode,
			input.CommissionRate,
			input.Status,
			createdAt,
			createdAt,
		))

	out, err := repo.Create(context.Background(), input)
	require.NoError(t, err)
	require.Equal(t, int64(43), out.MemberID)
	require.Equal(t, "kol1", out.RoleType)
	require.NotNil(t, out.ParentMemberID)
	require.Equal(t, parentID, *out.ParentMemberID)
	require.NoError(t, mock.ExpectationsWereMet())
}
