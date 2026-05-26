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

func distributionAttributionRows(
	userID int64,
	channelOrgID int64,
	referrerMemberID any,
	promotionLinkID any,
	boundAt time.Time,
	boundSource string,
	boundBy string,
	auditID any,
	createdAt time.Time,
	updatedAt time.Time,
) *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"user_id",
		"channel_org_id",
		"referrer_member_id",
		"promotion_link_id",
		"bound_at",
		"bound_source",
		"bound_by",
		"audit_id",
		"created_at",
		"updated_at",
	}).AddRow(
		userID,
		channelOrgID,
		referrerMemberID,
		promotionLinkID,
		boundAt,
		boundSource,
		boundBy,
		auditID,
		createdAt,
		updatedAt,
	)
}

func distributionAttributionAuditRows(
	id int64,
	userID int64,
	userEmail string,
	username string,
	previousChannelOrgID any,
	previousReferrerMemberID any,
	previousPromotionLinkID any,
	previousBoundSource string,
	previousBoundBy string,
	newChannelOrgID int64,
	newReferrerMemberID any,
	newPromotionLinkID any,
	newBoundSource string,
	newBoundBy string,
	note string,
	operatorUserID any,
	operatorUserEmail string,
	operatorUsername string,
	createdAt time.Time,
) *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id",
		"user_id",
		"user_email",
		"username",
		"previous_channel_org_id",
		"previous_referrer_member_id",
		"previous_promotion_link_id",
		"previous_bound_source",
		"previous_bound_by",
		"new_channel_org_id",
		"new_referrer_member_id",
		"new_promotion_link_id",
		"new_bound_source",
		"new_bound_by",
		"note",
		"operator_user_id",
		"operator_user_email",
		"operator_username",
		"created_at",
	}).AddRow(
		id,
		userID,
		userEmail,
		username,
		previousChannelOrgID,
		previousReferrerMemberID,
		previousPromotionLinkID,
		previousBoundSource,
		previousBoundBy,
		newChannelOrgID,
		newReferrerMemberID,
		newPromotionLinkID,
		newBoundSource,
		newBoundBy,
		note,
		operatorUserID,
		operatorUserEmail,
		operatorUsername,
		createdAt,
	)
}

func TestDistributionAttributionRepositoryUpdateByAdmin_AuditsAndRebindsAttribution(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &distributionAttributionRepository{sql: db}

	now := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	operatorUserID := int64(9)
	promotionLinkID := int64(202)

	mock.ExpectBegin()
	mock.ExpectQuery("FROM user_attributions").
		WithArgs(int64(7)).
		WillReturnRows(distributionAttributionRows(
			7,
			88,
			int64(101),
			int64(201),
			now.Add(-24*time.Hour),
			"registration",
			"system",
			int64(55),
			now.Add(-24*time.Hour),
			now.Add(-24*time.Hour),
		))
	mock.ExpectQuery("FROM promotion_links").
		WithArgs(promotionLinkID).
		WillReturnRows(sqlmock.NewRows([]string{"channel_org_id", "status"}).AddRow(int64(88), "active"))
	mock.ExpectQuery("SELECT member_id\\s+FROM promotion_links").
		WithArgs(promotionLinkID, int64(88)).
		WillReturnRows(sqlmock.NewRows([]string{"member_id"}).AddRow(int64(102)))
	mock.ExpectQuery("INSERT INTO distribution_attribution_audits").
		WithArgs(
			int64(7),
			int64(88),
			int64(101),
			int64(201),
			"registration",
			"system",
			int64(88),
			int64(102),
			int64(202),
			"manual",
			"admin",
			"manual reassignment",
			operatorUserID,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(66)))
	mock.ExpectQuery("INSERT INTO user_attributions").
		WithArgs(
			int64(7),
			int64(88),
			int64(102),
			int64(202),
			"manual",
			"admin",
			int64(66),
		).
		WillReturnRows(distributionAttributionRows(
			7,
			88,
			int64(102),
			int64(202),
			now,
			"manual",
			"admin",
			int64(66),
			now.Add(-24*time.Hour),
			now,
		))
	mock.ExpectCommit()

	out, err := repo.UpdateByAdmin(context.Background(), service.DistributionAttributionAdminUpdateInput{
		UserID:          7,
		ChannelOrgID:    88,
		PromotionLinkID: &promotionLinkID,
		BoundSource:     "manual",
		BoundBy:         "admin",
		OperatorUserID:  &operatorUserID,
		Note:            "manual reassignment",
	})
	require.NoError(t, err)
	require.NotNil(t, out)
	require.Equal(t, int64(7), out.UserID)
	require.Equal(t, int64(88), out.ChannelOrgID)
	require.NotNil(t, out.ReferrerMemberID)
	require.Equal(t, int64(102), *out.ReferrerMemberID)
	require.Equal(t, "manual", out.BoundSource)
	require.Equal(t, "admin", out.BoundBy)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDistributionAttributionRepositoryListAuditsAdmin(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &distributionAttributionRepository{sql: db}

	now := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)

	mock.ExpectQuery("SELECT COUNT\\(\\*\\)\\s+FROM distribution_attribution_audits da").
		WithArgs(int64(88), int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))
	mock.ExpectQuery("SELECT da.id,").
		WithArgs(int64(88), int64(7), 20, 0).
		WillReturnRows(distributionAttributionAuditRows(
			1,
			7,
			"user@example.com",
			"example-user",
			int64(66),
			int64(77),
			int64(88),
			"registration",
			"system",
			88,
			int64(99),
			int64(111),
			"manual",
			"admin",
			"manual reassignment",
			int64(9),
			"admin@example.com",
			"admin-user",
			now,
		))

	items, page, err := repo.ListAuditsAdmin(context.Background(), service.DistributionAdminListFilter{
		ChannelOrgID: 88,
		UserID:       7,
	}, pagination.PaginationParams{Page: 1, PageSize: 20})
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, int64(1), items[0].ID)
	require.Equal(t, int64(7), items[0].UserID)
	require.Equal(t, int64(88), items[0].NewChannelOrgID)
	require.NotNil(t, items[0].OperatorUserID)
	require.Equal(t, int64(9), *items[0].OperatorUserID)
	require.Equal(t, "admin@example.com", items[0].OperatorUserEmail)
	require.Equal(t, int64(1), page.Total)
	require.NoError(t, mock.ExpectationsWereMet())
}
