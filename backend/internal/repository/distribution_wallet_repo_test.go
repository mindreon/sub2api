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

func distributionWalletRows(channelOrgID int64, name string, orgType string, prepaidBalance float64, commissionReserved float64, totalRecharged float64, totalConsumed float64, warningThreshold float64, status string, createdAt time.Time, updatedAt time.Time) *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"channel_org_id",
		"name",
		"type",
		"prepaid_balance",
		"commission_reserved",
		"total_recharged",
		"total_consumed",
		"warning_threshold",
		"status",
		"created_at",
		"updated_at",
	}).AddRow(
		channelOrgID,
		name,
		orgType,
		prepaidBalance,
		commissionReserved,
		totalRecharged,
		totalConsumed,
		warningThreshold,
		status,
		createdAt,
		updatedAt,
	)
}

func distributionOrganizationRows(id int64, name string, orgType string, status string, config string, createdAt time.Time, updatedAt time.Time) *sqlmock.Rows {
	return sqlmock.NewRows([]string{
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
		id,
		orgType,
		name,
		nil,
		status,
		[]byte(config),
		[]byte(`{}`),
		createdAt,
		updatedAt,
	)
}

func expectWalletMutation(
	mock sqlmock.Sqlmock,
	channelOrgID int64,
	amount float64,
	transactionType string,
	before *service.DistributionWallet,
	after *service.DistributionWallet,
	orgConfig string,
) {
	mock.ExpectBegin()
	mock.ExpectQuery("FROM channel_wallets w").
		WithArgs(channelOrgID).
		WillReturnRows(distributionWalletRows(
			before.ChannelOrgID,
			before.OrganizationName,
			before.OrganizationType,
			before.PrepaidBalance,
			before.CommissionReserved,
			before.TotalRecharged,
			before.TotalConsumed,
			before.WarningThreshold,
			before.Status,
			before.CreatedAt,
			before.UpdatedAt,
		))
	mock.ExpectQuery("FROM channel_organizations").
		WithArgs(channelOrgID).
		WillReturnRows(distributionOrganizationRows(
			before.ChannelOrgID,
			before.OrganizationName,
			before.OrganizationType,
			"active",
			orgConfig,
			before.CreatedAt,
			before.UpdatedAt,
		))
	mock.ExpectQuery("UPDATE channel_wallets").
		WithArgs(channelOrgID, after.PrepaidBalance, after.CommissionReserved, after.TotalRecharged, after.TotalConsumed, after.Status).
		WillReturnRows(distributionWalletRows(
			after.ChannelOrgID,
			after.OrganizationName,
			after.OrganizationType,
			after.PrepaidBalance,
			after.CommissionReserved,
			after.TotalRecharged,
			after.TotalConsumed,
			after.WarningThreshold,
			after.Status,
			after.CreatedAt,
			after.UpdatedAt,
		))
	mock.ExpectQuery("FROM channel_alert_events").
		WithArgs(channelOrgID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "alert_type", "severity", "status", "details"}))
	mock.ExpectQuery("INSERT INTO channel_wallet_transactions").
		WithArgs(
			channelOrgID,
			transactionType,
			amount,
			before.PrepaidBalance,
			after.PrepaidBalance,
			before.CommissionReserved,
			after.CommissionReserved,
			"",
			"",
			nil,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))
	mock.ExpectCommit()
}

func TestDistributionWalletRepositoryList_ReturnsWallets(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &distributionWalletRepository{db: db}

	createdAt := time.Date(2026, 5, 23, 12, 0, 0, 0, time.UTC)
	mock.ExpectQuery("SELECT COUNT\\(\\*\\).*FROM channel_wallets").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))
	mock.ExpectQuery("FROM channel_wallets").
		WithArgs(20, 20).
		WillReturnRows(distributionWalletRows(int64(88), "Independent Agent", "reseller", 100.5, 12.25, 200.0, 87.75, 50.0, "active", createdAt, createdAt))

	out, page, err := repo.List(context.Background(), service.DistributionAdminListFilter{}, pagination.PaginationParams{Page: 2, PageSize: 20})
	require.NoError(t, err)
	require.Len(t, out, 1)
	require.Equal(t, int64(1), page.Total)
	require.Equal(t, "Independent Agent", out[0].OrganizationName)
	require.Equal(t, 100.5, out[0].PrepaidBalance)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDistributionWalletRepositoryListTransactions_ReturnsLedger(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &distributionWalletRepository{db: db}

	createdAt := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	mock.ExpectQuery("SELECT COUNT\\(\\*\\).*FROM channel_wallet_transactions").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))
	mock.ExpectQuery("FROM channel_wallet_transactions t").
		WithArgs(int64(88), "recharge", 20, 0).
		WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"channel_org_id",
			"name",
			"type",
			"transaction_type",
			"amount",
			"prepaid_balance_before",
			"prepaid_balance_after",
			"commission_reserved_before",
			"commission_reserved_after",
			"reference_no",
			"note",
			"operator_user_id",
			"created_at",
		}).AddRow(int64(1), int64(88), "Independent Agent", "reseller", "recharge", 120.0, 10.0, 130.0, 0.0, 0.0, "BANK-1", "confirmed", int64(9), createdAt))

	out, page, err := repo.ListTransactions(context.Background(), service.DistributionAdminListFilter{ChannelOrgID: 88, TransactionType: "recharge"}, pagination.PaginationParams{Page: 1, PageSize: 20})
	require.NoError(t, err)
	require.Len(t, out, 1)
	require.Equal(t, "recharge", out[0].TransactionType)
	require.NotNil(t, out[0].OperatorUserID)
	require.Equal(t, int64(1), page.Total)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDistributionWalletRepositoryGetAdminStats_ReturnsAggregates(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &distributionWalletRepository{db: db}

	mock.ExpectExec("UPDATE commission_ledger").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("FROM channel_organizations").
		WillReturnRows(sqlmock.NewRows([]string{"organization_count", "platform_count", "reseller_count", "oem_count"}).AddRow(int64(3), int64(1), int64(1), int64(1)))
	mock.ExpectQuery("FROM channel_members").
		WillReturnRows(sqlmock.NewRows([]string{"member_count", "agent_count", "kol1_count", "kol2_count"}).AddRow(int64(4), int64(2), int64(1), int64(1)))
	mock.ExpectQuery("FROM promotion_links").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(5)))
	mock.ExpectQuery("FROM user_attributions").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(6)))
	mock.ExpectQuery("FROM commission_ledger").
		WillReturnRows(sqlmock.NewRows([]string{"commission_count", "frozen_amount", "available_amount", "settled_amount"}).AddRow(int64(7), 1.5, 2.5, 3.5))
	mock.ExpectQuery("FROM channel_wallets").
		WillReturnRows(sqlmock.NewRows([]string{"wallet_count", "prepaid_balance_total", "commission_reserved_total", "total_recharged", "total_consumed"}).AddRow(int64(2), 100.0, 10.0, 120.0, 20.0))
	mock.ExpectQuery("SELECT value FROM settings").
		WithArgs(service.SettingKeyDistributionCommissionUpperRatio).
		WillReturnRows(sqlmock.NewRows([]string{"value"}).AddRow("35"))

	stats, err := repo.GetAdminStats(context.Background())
	require.NoError(t, err)
	require.Equal(t, int64(3), stats.OrganizationCount)
	require.Equal(t, float64(100.0), stats.PrepaidBalanceTotal)
	require.Equal(t, int64(7), stats.CommissionCount)
	require.InDelta(t, 0.375, stats.CommissionExpenseRatio, 0.0001)
	require.InDelta(t, 0.35, stats.CommissionUpperRatio, 0.0001)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDistributionWalletRepositoryReserveCommission_UpdatesReservedAmount(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &distributionWalletRepository{db: db}

	createdAt := time.Date(2026, 5, 23, 12, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 5, 24, 9, 0, 0, 0, time.UTC)
	expectWalletMutation(mock, 88, 12.5, "commission_reserve",
		&service.DistributionWallet{ChannelOrgID: 88, OrganizationName: "Independent Agent", OrganizationType: "reseller", PrepaidBalance: 100.0, CommissionReserved: 10.0, TotalRecharged: 200.0, TotalConsumed: 87.75, WarningThreshold: 50.0, Status: "active", CreatedAt: createdAt, UpdatedAt: createdAt},
		&service.DistributionWallet{ChannelOrgID: 88, OrganizationName: "Independent Agent", OrganizationType: "reseller", PrepaidBalance: 100.0, CommissionReserved: 22.5, TotalRecharged: 200.0, TotalConsumed: 87.75, WarningThreshold: 50.0, Status: "active", CreatedAt: createdAt, UpdatedAt: updatedAt},
		`{}`,
	)

	out, err := repo.ReserveCommission(context.Background(), 88, 12.5)
	require.NoError(t, err)
	require.NotNil(t, out)
	require.Equal(t, 22.5, out.CommissionReserved)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDistributionWalletRepositoryReleaseCommission_UpdatesReservedAmount(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &distributionWalletRepository{db: db}

	createdAt := time.Date(2026, 5, 23, 12, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 5, 24, 9, 0, 0, 0, time.UTC)
	expectWalletMutation(mock, 88, 12.5, "commission_release",
		&service.DistributionWallet{ChannelOrgID: 88, OrganizationName: "Independent Agent", OrganizationType: "reseller", PrepaidBalance: 100.0, CommissionReserved: 22.5, TotalRecharged: 200.0, TotalConsumed: 87.75, WarningThreshold: 50.0, Status: "active", CreatedAt: createdAt, UpdatedAt: createdAt},
		&service.DistributionWallet{ChannelOrgID: 88, OrganizationName: "Independent Agent", OrganizationType: "reseller", PrepaidBalance: 100.0, CommissionReserved: 10.0, TotalRecharged: 200.0, TotalConsumed: 87.75, WarningThreshold: 50.0, Status: "active", CreatedAt: createdAt, UpdatedAt: updatedAt},
		`{}`,
	)

	out, err := repo.ReleaseCommission(context.Background(), 88, 12.5)
	require.NoError(t, err)
	require.NotNil(t, out)
	require.Equal(t, 10.0, out.CommissionReserved)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDistributionWalletRepositorySettleReservedCommission_DeductsPrepaidBalance(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &distributionWalletRepository{db: db}

	createdAt := time.Date(2026, 5, 23, 12, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 5, 24, 9, 0, 0, 0, time.UTC)
	expectWalletMutation(mock, 88, 12.5, "commission_settle",
		&service.DistributionWallet{ChannelOrgID: 88, OrganizationName: "Independent Agent", OrganizationType: "reseller", PrepaidBalance: 100.0, CommissionReserved: 12.5, TotalRecharged: 200.0, TotalConsumed: 87.75, WarningThreshold: 50.0, Status: "active", CreatedAt: createdAt, UpdatedAt: createdAt},
		&service.DistributionWallet{ChannelOrgID: 88, OrganizationName: "Independent Agent", OrganizationType: "reseller", PrepaidBalance: 87.5, CommissionReserved: 0.0, TotalRecharged: 200.0, TotalConsumed: 87.75, WarningThreshold: 50.0, Status: "active", CreatedAt: createdAt, UpdatedAt: updatedAt},
		`{}`,
	)

	out, err := repo.SettleReservedCommission(context.Background(), 88, 12.5)
	require.NoError(t, err)
	require.NotNil(t, out)
	require.Equal(t, 87.5, out.PrepaidBalance)
	require.Equal(t, 0.0, out.CommissionReserved)
	require.Equal(t, 87.75, out.TotalConsumed)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDistributionWalletRepositoryRefundCommission_ReplenishesPrepaidBalance(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &distributionWalletRepository{db: db}

	createdAt := time.Date(2026, 5, 23, 12, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 5, 24, 9, 0, 0, 0, time.UTC)
	expectWalletMutation(mock, 88, 12.5, "commission_refund",
		&service.DistributionWallet{ChannelOrgID: 88, OrganizationName: "Independent Agent", OrganizationType: "reseller", PrepaidBalance: 87.5, CommissionReserved: 0.0, TotalRecharged: 200.0, TotalConsumed: 87.75, WarningThreshold: 50.0, Status: "active", CreatedAt: createdAt, UpdatedAt: createdAt},
		&service.DistributionWallet{ChannelOrgID: 88, OrganizationName: "Independent Agent", OrganizationType: "reseller", PrepaidBalance: 100.0, CommissionReserved: 0.0, TotalRecharged: 200.0, TotalConsumed: 87.75, WarningThreshold: 50.0, Status: "active", CreatedAt: createdAt, UpdatedAt: updatedAt},
		`{}`,
	)

	out, err := repo.RefundCommission(context.Background(), 88, 12.5)
	require.NoError(t, err)
	require.NotNil(t, out)
	require.Equal(t, 100.0, out.PrepaidBalance)
	require.Equal(t, 87.75, out.TotalConsumed)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDistributionWalletRepositoryRefundPrepaidBalance_DeductsAvailableBalance(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &distributionWalletRepository{db: db}

	createdAt := time.Date(2026, 5, 23, 12, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 5, 24, 9, 0, 0, 0, time.UTC)
	before := &service.DistributionWallet{ChannelOrgID: 88, OrganizationName: "Independent Agent", OrganizationType: "reseller", PrepaidBalance: 200.0, CommissionReserved: 30.0, TotalRecharged: 200.0, TotalConsumed: 87.75, WarningThreshold: 50.0, Status: "active", CreatedAt: createdAt, UpdatedAt: createdAt}
	after := &service.DistributionWallet{ChannelOrgID: 88, OrganizationName: "Independent Agent", OrganizationType: "reseller", PrepaidBalance: 120.0, CommissionReserved: 30.0, TotalRecharged: 200.0, TotalConsumed: 87.75, WarningThreshold: 50.0, Status: "active", CreatedAt: createdAt, UpdatedAt: updatedAt}

	mock.ExpectBegin()
	mock.ExpectQuery("FROM channel_wallets w").
		WithArgs(int64(88)).
		WillReturnRows(distributionWalletRows(
			before.ChannelOrgID,
			before.OrganizationName,
			before.OrganizationType,
			before.PrepaidBalance,
			before.CommissionReserved,
			before.TotalRecharged,
			before.TotalConsumed,
			before.WarningThreshold,
			before.Status,
			before.CreatedAt,
			before.UpdatedAt,
		))
	mock.ExpectQuery("FROM channel_organizations").
		WithArgs(int64(88)).
		WillReturnRows(distributionOrganizationRows(int64(88), "Independent Agent", "reseller", "active", `{}`, createdAt, createdAt))
	mock.ExpectQuery("UPDATE channel_wallets").
		WithArgs(int64(88), after.PrepaidBalance, after.CommissionReserved, after.TotalRecharged, after.TotalConsumed, after.Status).
		WillReturnRows(distributionWalletRows(
			after.ChannelOrgID,
			after.OrganizationName,
			after.OrganizationType,
			after.PrepaidBalance,
			after.CommissionReserved,
			after.TotalRecharged,
			after.TotalConsumed,
			after.WarningThreshold,
			after.Status,
			after.CreatedAt,
			after.UpdatedAt,
		))
	mock.ExpectQuery("FROM channel_alert_events").
		WithArgs(int64(88)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "alert_type", "severity", "status", "details"}))
	mock.ExpectQuery("INSERT INTO channel_wallet_transactions").
		WithArgs(
			int64(88),
			"refund",
			80.0,
			before.PrepaidBalance,
			after.PrepaidBalance,
			before.CommissionReserved,
			after.CommissionReserved,
			"RF-1",
			"manual refund",
			int64(9),
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))
	mock.ExpectCommit()

	operatorUserID := int64(9)
	out, err := repo.RefundPrepaidBalance(context.Background(), 88, service.DistributionWalletRefundInput{
		Amount:         80,
		ReferenceNo:    "RF-1",
		Note:           "manual refund",
		OperatorUserID: &operatorUserID,
	})
	require.NoError(t, err)
	require.NotNil(t, out)
	require.InDelta(t, 120, out.PrepaidBalance, 0.0001)
	require.InDelta(t, 30, out.CommissionReserved, 0.0001)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDistributionWalletRepositoryConsumeUsage_DeductsBalanceAndTracksConsumption(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &distributionWalletRepository{db: db}

	createdAt := time.Date(2026, 5, 23, 12, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 5, 24, 9, 0, 0, 0, time.UTC)
	before := &service.DistributionWallet{ChannelOrgID: 88, OrganizationName: "Independent Agent", OrganizationType: "reseller", PrepaidBalance: 100.0, CommissionReserved: 10.0, TotalRecharged: 200.0, TotalConsumed: 30.0, WarningThreshold: 50.0, Status: "active", CreatedAt: createdAt, UpdatedAt: createdAt}
	after := &service.DistributionWallet{ChannelOrgID: 88, OrganizationName: "Independent Agent", OrganizationType: "reseller", PrepaidBalance: 87.5, CommissionReserved: 10.0, TotalRecharged: 200.0, TotalConsumed: 42.5, WarningThreshold: 50.0, Status: "active", CreatedAt: createdAt, UpdatedAt: updatedAt}

	mock.ExpectBegin()
	mock.ExpectQuery("FROM channel_wallets w").
		WithArgs(int64(88)).
		WillReturnRows(distributionWalletRows(
			before.ChannelOrgID,
			before.OrganizationName,
			before.OrganizationType,
			before.PrepaidBalance,
			before.CommissionReserved,
			before.TotalRecharged,
			before.TotalConsumed,
			before.WarningThreshold,
			before.Status,
			before.CreatedAt,
			before.UpdatedAt,
		))
	mock.ExpectQuery("FROM channel_organizations").
		WithArgs(int64(88)).
		WillReturnRows(distributionOrganizationRows(int64(88), "Independent Agent", "reseller", "active", `{"consumption_limit":100}`, createdAt, createdAt))
	mock.ExpectQuery("UPDATE channel_wallets").
		WithArgs(int64(88), after.PrepaidBalance, after.CommissionReserved, after.TotalRecharged, after.TotalConsumed, after.Status).
		WillReturnRows(distributionWalletRows(
			after.ChannelOrgID,
			after.OrganizationName,
			after.OrganizationType,
			after.PrepaidBalance,
			after.CommissionReserved,
			after.TotalRecharged,
			after.TotalConsumed,
			after.WarningThreshold,
			after.Status,
			after.CreatedAt,
			after.UpdatedAt,
		))
	mock.ExpectQuery("FROM channel_alert_events").
		WithArgs(int64(88)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "alert_type", "severity", "status", "details"}))
	mock.ExpectQuery("INSERT INTO channel_wallet_transactions").
		WithArgs(int64(88), "consume", 12.5, before.PrepaidBalance, after.PrepaidBalance, before.CommissionReserved, after.CommissionReserved, "", "", nil).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))
	mock.ExpectCommit()

	out, err := repo.ConsumeUsage(context.Background(), 88, 12.5, 100.0)
	require.NoError(t, err)
	require.NotNil(t, out)
	require.Equal(t, 87.5, out.PrepaidBalance)
	require.Equal(t, 42.5, out.TotalConsumed)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDistributionWalletRepositorySyncStatusMarksWalletInactiveWhenBalanceExhausted(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &distributionWalletRepository{db: db}

	now := time.Date(2026, 5, 24, 9, 0, 0, 0, time.UTC)
	mock.ExpectBegin()
	mock.ExpectQuery("FROM channel_wallets w").
		WithArgs(int64(88)).
		WillReturnRows(distributionWalletRows(88, "Independent Agent", "reseller", 10, 10, 200, 30, 50, "active", now, now))
	mock.ExpectQuery("FROM channel_organizations").
		WithArgs(int64(88)).
		WillReturnRows(distributionOrganizationRows(int64(88), "Independent Agent", "reseller", "active", `{}`, now, now))
	mock.ExpectQuery("UPDATE channel_wallets\\s+SET status = \\$2").
		WithArgs(int64(88), "inactive").
		WillReturnRows(distributionWalletRows(88, "Independent Agent", "reseller", 10, 10, 200, 30, 50, "inactive", now, now))
	mock.ExpectQuery("FROM channel_alert_events").
		WithArgs(int64(88)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "alert_type", "severity", "status", "details"}))
	mock.ExpectExec("INSERT INTO channel_alert_events").
		WithArgs(int64(88), "balance_exhausted", "critical", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	out, err := repo.SyncStatus(context.Background(), 88, 100)
	require.NoError(t, err)
	require.NotNil(t, out)
	require.Equal(t, "inactive", out.Status)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDistributionWalletRepositoryRechargeReactivatesWalletWhenBalanceRecovers(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &distributionWalletRepository{db: db}

	createdAt := time.Date(2026, 5, 23, 12, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 5, 24, 9, 0, 0, 0, time.UTC)
	mock.ExpectBegin()
	mock.ExpectQuery("FROM channel_wallets w").
		WithArgs(int64(88)).
		WillReturnRows(distributionWalletRows(88, "Independent Agent", "reseller", 0, 0, 200, 30, 50, "inactive", createdAt, createdAt))
	mock.ExpectQuery("FROM channel_organizations").
		WithArgs(int64(88)).
		WillReturnRows(distributionOrganizationRows(int64(88), "Independent Agent", "reseller", "active", `{"consumption_limit":100}`, createdAt, createdAt))
	mock.ExpectQuery("UPDATE channel_wallets").
		WithArgs(int64(88), 50.0, 0.0, 250.0, 30.0, "active").
		WillReturnRows(distributionWalletRows(88, "Independent Agent", "reseller", 50, 0, 250, 30, 50, "active", createdAt, updatedAt))
	mock.ExpectQuery("FROM channel_alert_events").
		WithArgs(int64(88)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "alert_type", "severity", "status", "details"}))
	mock.ExpectExec("INSERT INTO channel_alert_events").
		WithArgs(int64(88), "low_balance", "warning", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery("INSERT INTO channel_wallet_transactions").
		WithArgs(int64(88), "recharge", 50.0, 0.0, 50.0, 0.0, 0.0, "", "", nil).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))
	mock.ExpectCommit()

	out, err := repo.Recharge(context.Background(), 88, service.DistributionWalletRechargeInput{Amount: 50})
	require.NoError(t, err)
	require.NotNil(t, out)
	require.Equal(t, "active", out.Status)
	require.NoError(t, mock.ExpectationsWereMet())
}
