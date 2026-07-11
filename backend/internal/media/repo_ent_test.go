package media

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/enttest"
	"github.com/stretchr/testify/require"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "modernc.org/sqlite"
)

func TestEntTaskStoreUpdatePersistsAccountID(t *testing.T) {
	db, err := sql.Open("sqlite", "file:media_task_store_account?mode=memory&cache=shared&_fk=1")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err)

	driver := entsql.OpenDB(dialect.SQLite, db)
	client := enttest.NewClient(t, enttest.WithOptions(dbent.Driver(driver)))
	t.Cleanup(func() { _ = client.Close() })

	store := NewEntTaskStore(client)
	task := &Task{
		TaskID:          "task-account-persist",
		UserID:          1,
		APIKeyID:        2,
		Model:           "seedance2.0-fast-p5",
		MediaType:       "video",
		Status:          TaskPending,
		ReservedCost:    0.25,
		RateMultiplier:  1,
		BillingCurrency: CurrencyUSD,
		ExpiresAt:       time.Now().Add(time.Hour),
	}
	require.NoError(t, store.Create(context.Background(), task))

	accountID := int64(6)
	task.AccountID = &accountID
	task.UpstreamTaskID = "upstream-task-1"
	task.Status = TaskInProgress
	require.NoError(t, store.Update(context.Background(), task))

	stored, err := store.GetByTaskID(context.Background(), task.TaskID)
	require.NoError(t, err)
	require.NotNil(t, stored.AccountID)
	require.Equal(t, accountID, *stored.AccountID)
}

func TestEntTaskStoreListFiltersCreatedAtInclusive(t *testing.T) {
	db, err := sql.Open("sqlite", "file:media_task_store_created_at?mode=memory&cache=shared&_fk=1")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err)

	driver := entsql.OpenDB(dialect.SQLite, db)
	client := enttest.NewClient(t, enttest.WithOptions(dbent.Driver(driver)))
	t.Cleanup(func() { _ = client.Close() })

	from := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 7, 31, 23, 59, 59, 0, time.UTC)
	createdAt := []time.Time{from.Add(-time.Second), from, to}
	for i, created := range createdAt {
		_, err = client.MediaGenerationTask.Create().
			SetTaskID(fmt.Sprintf("task-created-at-%d", i)).
			SetUserID(1).
			SetAPIKeyID(2).
			SetModel("seedance2.0-fast-p5").
			SetMediaType("video").
			SetStatus(string(TaskPending)).
			SetReservedCost(0.25).
			SetRateMultiplier(1).
			SetBillingCurrency(string(CurrencyUSD)).
			SetExpiresAt(created.Add(time.Hour)).
			SetCreatedAt(created).
			SetUpdatedAt(created).
			Save(context.Background())
		require.NoError(t, err)
	}

	result, err := NewEntTaskStore(client).List(context.Background(), TaskListQuery{
		CreatedFrom: &from,
		CreatedTo:   &to,
		Page:        1,
		PageSize:    1,
	})
	require.NoError(t, err)
	require.Equal(t, 2, result.Total)
	require.Len(t, result.Tasks, 1)
	require.Equal(t, "task-created-at-2", result.Tasks[0].TaskID)
}
