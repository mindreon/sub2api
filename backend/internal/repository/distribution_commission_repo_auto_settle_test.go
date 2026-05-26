package repository

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestDistributionCommissionRepositoryListAutoSettleCommissionIDs_ThawsAndListsAvailableAutoIDs(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &distributionCommissionRepository{db: db}

	mock.ExpectExec("UPDATE commission_ledger").
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectQuery("SELECT id FROM commission_ledger").
		WithArgs(50).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).
			AddRow(int64(1001)).
			AddRow(int64(1002)))

	ids, err := repo.ListAutoSettleCommissionIDs(context.Background(), 50)
	require.NoError(t, err)
	require.Equal(t, []int64{1001, 1002}, ids)
	require.NoError(t, mock.ExpectationsWereMet())
}
