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

func TestDistributionCommissionRepositoryCreate_PersistsLedgerInput(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &distributionCommissionRepository{db: db}

	usageLogID := int64(555)
	frozenUntil := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	createdAt := time.Date(2026, 5, 23, 12, 0, 0, 0, time.UTC)
	input := service.DistributionCommissionInput{
		ChannelOrgID:     99,
		MemberID:         42,
		UserID:           7,
		UsageLogID:       &usageLogID,
		CommissionType:   "direct",
		BaseAmount:       8.75,
		Rate:             0.15,
		Amount:           1.3125,
		Status:           "frozen",
		SettlementMethod: "balance",
		FrozenUntil:      &frozenUntil,
	}

	mock.ExpectQuery("INSERT INTO commission_ledger").
		WithArgs(
			input.ChannelOrgID,
			input.MemberID,
			input.UserID,
			usageLogID,
			input.CommissionType,
			input.BaseAmount,
			input.Rate,
			input.Amount,
			input.Status,
			input.SettlementMethod,
			"",
			"",
			frozenUntil,
			sqlmock.AnyArg(),
			nil,
			sqlmock.AnyArg(),
		).
		WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"channel_org_id",
			"member_id",
			"user_id",
			"usage_log_id",
			"commission_type",
			"base_amount",
			"rate",
			"amount",
			"status",
			"settlement_method",
			"settlement_reference_no",
			"settlement_note",
			"frozen_until",
			"settled_at",
			"settled_by_user_id",
			"reversed_from_id",
			"created_at",
			"updated_at",
		}).AddRow(
			int64(1001),
			input.ChannelOrgID,
			input.MemberID,
			input.UserID,
			usageLogID,
			input.CommissionType,
			input.BaseAmount,
			input.Rate,
			input.Amount,
			input.Status,
			input.SettlementMethod,
			"",
			"",
			frozenUntil,
			nil,
			nil,
			nil,
			createdAt,
			createdAt,
		))

	out, err := repo.Create(context.Background(), input)
	require.NoError(t, err)
	require.Equal(t, int64(1001), out.ID)
	require.Equal(t, input.BaseAmount, out.BaseAmount)
	require.Equal(t, input.Amount, out.Amount)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDistributionCommissionRepositoryGetByID_ThawsMaturedFrozenLedger(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &distributionCommissionRepository{db: db}

	updatedAt := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	mock.ExpectExec("UPDATE commission_ledger").
		WithArgs(int64(1001)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("SELECT id,.*FROM commission_ledger").
		WithArgs(int64(1001)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"channel_org_id",
			"member_id",
			"user_id",
			"usage_log_id",
			"commission_type",
			"base_amount",
			"rate",
			"amount",
			"status",
			"settlement_method",
			"settlement_reference_no",
			"settlement_note",
			"frozen_until",
			"settled_at",
			"settled_by_user_id",
			"reversed_from_id",
			"created_at",
			"updated_at",
		}).AddRow(
			int64(1001),
			int64(99),
			int64(42),
			int64(7),
			nil,
			"direct",
			8.75,
			0.15,
			1.3125,
			"available",
			"manual",
			"",
			"",
			nil,
			nil,
			nil,
			nil,
			updatedAt,
			updatedAt,
		))

	out, err := repo.GetByID(context.Background(), 1001)
	require.NoError(t, err)
	require.NotNil(t, out)
	require.Equal(t, "available", out.Status)
	require.Nil(t, out.FrozenUntil)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDistributionCommissionRepositoryListByChannelOrgID_ThawsMaturedFrozenLedgersBeforeListing(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &distributionCommissionRepository{db: db}

	updatedAt := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	mock.ExpectExec("UPDATE commission_ledger").
		WithArgs(int64(99)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("SELECT COUNT\\(\\*\\).*FROM commission_ledger").
		WithArgs(int64(99)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))
	mock.ExpectQuery("FROM commission_ledger cl").
		WithArgs(int64(99), 20, 0).
		WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"channel_org_id",
			"member_id",
			"user_id",
			"user_email",
			"username",
			"usage_log_id",
			"commission_type",
			"base_amount",
			"rate",
			"amount",
			"status",
			"settlement_method",
			"settlement_reference_no",
			"settlement_note",
			"frozen_until",
			"settled_at",
			"settled_by_user_id",
			"reversed_from_id",
			"created_at",
			"updated_at",
		}).AddRow(
			int64(1001),
			int64(99),
			int64(42),
			int64(7),
			"user@example.com",
			"demo",
			nil,
			"direct",
			8.75,
			0.15,
			1.3125,
			"available",
			"manual",
			"",
			"",
			nil,
			nil,
			nil,
			nil,
			updatedAt,
			updatedAt,
		))

	out, page, err := repo.ListByChannelOrgID(context.Background(), 99, pagination.PaginationParams{Page: 1, PageSize: 20})
	require.NoError(t, err)
	require.Len(t, out, 1)
	require.Equal(t, int64(1), page.Total)
	require.Equal(t, "available", out[0].Status)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDistributionCommissionRepositoryGetTotalCommissionByUserID_ReturnsNonCancelledTotal(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &distributionCommissionRepository{db: db}

	mock.ExpectQuery("SELECT COALESCE\\(SUM\\(amount\\), 0\\)::double precision FROM commission_ledger").
		WithArgs(int64(99), int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"sum"}).AddRow(18.5))

	total, err := repo.GetTotalCommissionByUserID(context.Background(), 99, 7)
	require.NoError(t, err)
	require.InDelta(t, 18.5, total, 0.0001)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDistributionCommissionRepositoryGetTotalCommissionByMemberID_ReturnsNonCancelledTotal(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &distributionCommissionRepository{db: db}

	mock.ExpectQuery("SELECT COALESCE\\(SUM\\(amount\\), 0\\)::double precision FROM commission_ledger").
		WithArgs(int64(99), int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"sum"}).AddRow(21.75))

	total, err := repo.GetTotalCommissionByMemberID(context.Background(), 99, 42)
	require.NoError(t, err)
	require.InDelta(t, 21.75, total, 0.0001)
	require.NoError(t, mock.ExpectationsWereMet())
}
