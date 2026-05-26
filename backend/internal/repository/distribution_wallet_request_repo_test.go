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

func distributionWalletRequestRows(id int64, channelOrgID int64, requestType string, amount float64, status string, createdByUserID int64, reviewedByUserID *int64, reviewedAt *time.Time, reviewNote string, referenceNo string, note string, createdAt time.Time) *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id",
		"channel_org_id",
		"organization_name",
		"organization_type",
		"request_type",
		"amount",
		"reference_no",
		"note",
		"status",
		"created_by_user_id",
		"created_by_user_email",
		"created_by_username",
		"reviewed_by_user_id",
		"reviewed_by_user_email",
		"reviewed_by_username",
		"review_note",
		"reviewed_at",
		"created_at",
	}).AddRow(
		id,
		channelOrgID,
		"Independent Agent",
		"reseller",
		requestType,
		amount,
		referenceNo,
		note,
		status,
		createdByUserID,
		"manager@example.com",
		"manager",
		reviewedByUserID,
		"admin@example.com",
		"admin",
		reviewNote,
		reviewedAt,
		createdAt,
	)
}

func TestDistributionWalletRequestRepository_ListRequests(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &distributionWalletRequestRepository{db: db}
	now := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)

	mock.ExpectQuery("SELECT COUNT\\(\\*\\).*FROM channel_wallet_requests r").
		WithArgs(int64(88), "recharge", "pending").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))
	mock.ExpectQuery("FROM channel_wallet_requests r").
		WithArgs(int64(88), "recharge", "pending", 20, 0).
		WillReturnRows(distributionWalletRequestRows(1, 88, "recharge", 120, "pending", 7, nil, nil, "", "BANK-1", "bank transfer", now))

	items, page, err := repo.ListRequests(context.Background(), service.DistributionWalletRequestListFilter{
		ChannelOrgID: 88,
		RequestType:  "recharge",
		Status:       "pending",
	}, pagination.PaginationParams{Page: 1, PageSize: 20})
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, int64(1), page.Total)
	require.Equal(t, "recharge", items[0].RequestType)
	require.Equal(t, "pending", items[0].Status)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDistributionWalletRequestRepository_CreateRequest(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &distributionWalletRequestRepository{db: db}
	now := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)

	mock.ExpectQuery("INSERT INTO channel_wallet_requests").
		WithArgs(int64(88), "refund", 50.0, "RF-1", "return balance", int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(9)))
	mock.ExpectQuery("FROM channel_wallet_requests r").
		WithArgs(int64(9)).
		WillReturnRows(distributionWalletRequestRows(9, 88, "refund", 50, "pending", 7, nil, nil, "", "RF-1", "return balance", now))

	out, err := repo.CreateRequest(context.Background(), service.DistributionWalletRequestCreateInput{
		ChannelOrgID:    88,
		RequestType:     "refund",
		Amount:          50,
		ReferenceNo:     "RF-1",
		Note:            "return balance",
		CreatedByUserID: 7,
	})
	require.NoError(t, err)
	require.NotNil(t, out)
	require.Equal(t, int64(9), out.ID)
	require.Equal(t, "pending", out.Status)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDistributionWalletRequestRepository_ApproveRechargeRequestMutatesWalletAndRequest(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &distributionWalletRequestRepository{db: db}
	now := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	reviewedBy := int64(9)

	mock.ExpectBegin()
	mock.ExpectQuery("FROM channel_wallet_requests r").
		WithArgs(int64(31)).
		WillReturnRows(distributionWalletRequestRows(31, 88, "recharge", 120, "pending", 7, nil, nil, "", "BANK-1", "bank transfer", now))
	mock.ExpectQuery("FROM channel_wallets w").
		WithArgs(int64(88)).
		WillReturnRows(distributionWalletRows(88, "Independent Agent", "reseller", 80, 0, 200, 120, 50, "active", now, now))
	mock.ExpectQuery("FROM channel_organizations").
		WithArgs(int64(88)).
		WillReturnRows(distributionOrganizationRows(int64(88), "Independent Agent", "reseller", "active", `{}`, now, now))
	mock.ExpectQuery("UPDATE channel_wallets").
		WithArgs(int64(88), 200.0, 0.0, 320.0, 120.0, "active").
		WillReturnRows(distributionWalletRows(88, "Independent Agent", "reseller", 200, 0, 320, 120, 50, "active", now, now))
	mock.ExpectQuery("FROM channel_alert_events").
		WithArgs(int64(88)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "alert_type", "severity", "status", "details"}))
	mock.ExpectQuery("INSERT INTO channel_wallet_transactions").
		WithArgs(int64(88), "recharge", 120.0, 80.0, 200.0, 0.0, 0.0, "BANK-1", "confirmed request_id=31", reviewedBy).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1001)))
	mock.ExpectQuery("UPDATE channel_wallet_requests").
		WithArgs("approved", "到账确认", reviewedBy, int64(31)).
		WillReturnRows(distributionWalletRequestRows(31, 88, "recharge", 120, "approved", 7, &reviewedBy, &now, "到账确认", "BANK-1", "bank transfer", now))
	mock.ExpectCommit()

	out, err := repo.ApproveRechargeRequest(context.Background(), 31, service.DistributionWalletRequestReviewInput{
		ReviewNote:       "到账确认",
		ReviewedByUserID: reviewedBy,
	}, service.DistributionWalletRechargeInput{
		Amount:         120,
		ReferenceNo:    "BANK-1",
		Note:           "confirmed request_id=31",
		OperatorUserID: &reviewedBy,
	})
	require.NoError(t, err)
	require.NotNil(t, out)
	require.Equal(t, "approved", out.Status)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDistributionWalletRequestRepository_ApproveRefundRequestMutatesWalletAndRequest(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &distributionWalletRequestRepository{db: db}
	now := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	reviewedBy := int64(9)

	mock.ExpectBegin()
	mock.ExpectQuery("FROM channel_wallet_requests r").
		WithArgs(int64(45)).
		WillReturnRows(distributionWalletRequestRows(45, 88, "refund", 100, "pending", 7, nil, nil, "", "RF-1", "refund request", now))
	mock.ExpectQuery("FROM channel_wallets w").
		WithArgs(int64(88)).
		WillReturnRows(distributionWalletRows(88, "Independent Agent", "reseller", 200, 20, 300, 100, 50, "active", now, now))
	mock.ExpectQuery("FROM channel_organizations").
		WithArgs(int64(88)).
		WillReturnRows(distributionOrganizationRows(int64(88), "Independent Agent", "reseller", "active", `{}`, now, now))
	mock.ExpectQuery("UPDATE channel_wallets").
		WithArgs(int64(88), 100.0, 20.0, 300.0, 100.0, "active").
		WillReturnRows(distributionWalletRows(88, "Independent Agent", "reseller", 100, 20, 300, 100, 50, "active", now, now))
	mock.ExpectQuery("FROM channel_alert_events").
		WithArgs(int64(88)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "alert_type", "severity", "status", "details"}))
	mock.ExpectQuery("INSERT INTO channel_wallet_transactions").
		WithArgs(int64(88), "refund", 100.0, 200.0, 100.0, 20.0, 20.0, "RF-1", "mock_refund fee_amount=10 net_amount=90 request_id=45", reviewedBy).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1002)))
	mock.ExpectQuery("UPDATE channel_wallet_requests").
		WithArgs("approved", "同意退款", reviewedBy, int64(45)).
		WillReturnRows(distributionWalletRequestRows(45, 88, "refund", 100, "approved", 7, &reviewedBy, &now, "同意退款", "RF-1", "refund request", now))
	mock.ExpectCommit()

	out, err := repo.ApproveRefundRequest(context.Background(), 45, service.DistributionWalletRequestReviewInput{
		ReviewNote:       "同意退款",
		ReviewedByUserID: reviewedBy,
	}, service.DistributionWalletRefundInput{
		Amount:         100,
		ReferenceNo:    "RF-1",
		Note:           "mock_refund fee_amount=10 net_amount=90 request_id=45",
		OperatorUserID: &reviewedBy,
	})
	require.NoError(t, err)
	require.NotNil(t, out)
	require.Equal(t, "approved", out.Status)
	require.NoError(t, mock.ExpectationsWereMet())
}
