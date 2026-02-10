package db_test

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"sboard/panel/internal/db"
)

func setupStoreForSyncJobs(t *testing.T) *db.Store {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	database, err := db.Open(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { _ = database.Close() })

	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)
	migrationsDir := filepath.Join(filepath.Dir(file), "migrations")
	err = db.MigrateUp(database, migrationsDir)
	require.NoError(t, err)

	store := db.NewStore(database)
	store.Now = func() time.Time {
		return time.Date(2026, 2, 8, 12, 0, 0, 0, time.UTC)
	}
	return store
}

func TestSyncJobs_CreateAndFinish(t *testing.T) {
	ctx := context.Background()
	store := setupStoreForSyncJobs(t)

	g, err := store.CreateGroup(ctx, "g1", "")
	require.NoError(t, err)
	n, err := store.CreateNode(ctx, db.NodeCreate{
		Name:          "n1",
		APIAddress:    "127.0.0.1",
		APIPort:       3003,
		SecretKey:     "secret",
		PublicAddress: "example.com",
		GroupID:       &g.ID,
	})
	require.NoError(t, err)

	job, err := store.CreateSyncJob(ctx, db.SyncJobCreate{
		NodeID:        n.ID,
		TriggerSource: "auto_inbound_change",
	})
	require.NoError(t, err)
	require.Equal(t, db.SyncJobStatusQueued, job.Status)

	started := time.Date(2026, 2, 8, 12, 1, 0, 0, time.UTC)
	require.NoError(t, store.UpdateSyncJobStart(ctx, job.ID, db.SyncJobStartUpdate{
		InboundCount:    3,
		ActiveUserCount: 5,
		PayloadHash:     "abc",
		StartedAt:       started,
	}))

	attempt, err := store.CreateSyncAttempt(ctx, db.SyncAttemptCreate{
		JobID:     job.ID,
		AttemptNo: 1,
		StartedAt: started,
	})
	require.NoError(t, err)

	finished := started.Add(2 * time.Second)
	require.NoError(t, store.UpdateSyncAttemptFinish(ctx, attempt.ID, db.SyncAttemptFinishUpdate{
		Status:     db.SyncAttemptStatusSuccess,
		HTTPStatus: 200,
		FinishedAt: finished,
		DurationMS: 2000,
	}))
	require.NoError(t, store.UpdateSyncJobFinish(ctx, job.ID, db.SyncJobFinishUpdate{
		Status:       db.SyncJobStatusSuccess,
		AttemptCount: 1,
		FinishedAt:   finished,
		DurationMS:   2000,
	}))

	got, err := store.GetSyncJobByID(ctx, job.ID)
	require.NoError(t, err)
	require.Equal(t, db.SyncJobStatusSuccess, got.Status)
	require.Equal(t, 3, got.InboundCount)
	require.Equal(t, 5, got.ActiveUserCount)
	require.Equal(t, "abc", got.PayloadHash)
	require.Equal(t, 1, got.AttemptCount)

	attempts, err := store.ListSyncAttemptsByJobID(ctx, job.ID)
	require.NoError(t, err)
	require.Len(t, attempts, 1)
	require.Equal(t, db.SyncAttemptStatusSuccess, attempts[0].Status)
	require.Equal(t, 200, attempts[0].HTTPStatus)
}

func TestSyncJobs_PruneByNode(t *testing.T) {
	ctx := context.Background()
	store := setupStoreForSyncJobs(t)

	g, err := store.CreateGroup(ctx, "g1", "")
	require.NoError(t, err)
	n, err := store.CreateNode(ctx, db.NodeCreate{
		Name:          "n1",
		APIAddress:    "127.0.0.1",
		APIPort:       3003,
		SecretKey:     "secret",
		PublicAddress: "example.com",
		GroupID:       &g.ID,
	})
	require.NoError(t, err)

	ids := make([]int64, 0, 3)
	for i := 0; i < 3; i++ {
		job, err := store.CreateSyncJob(ctx, db.SyncJobCreate{NodeID: n.ID, TriggerSource: "auto"})
		require.NoError(t, err)
		ids = append(ids, job.ID)
		// bump store.Now so created_at can sort deterministically
		ii := i
		store.Now = func() time.Time {
			return time.Date(2026, 2, 8, 12, 0, ii+1, 0, time.UTC)
		}
	}

	require.NoError(t, store.PruneSyncJobsByNode(ctx, n.ID, 2))

	_, err = store.GetSyncJobByID(ctx, ids[0])
	require.ErrorIs(t, err, db.ErrNotFound)
	_, err = store.GetSyncJobByID(ctx, ids[1])
	require.NoError(t, err)
	_, err = store.GetSyncJobByID(ctx, ids[2])
	require.NoError(t, err)
}

func TestSyncJobs_ListFilters(t *testing.T) {
	ctx := context.Background()
	store := setupStoreForSyncJobs(t)

	current := time.Date(2026, 2, 8, 12, 0, 0, 0, time.UTC)
	store.Now = func() time.Time { return current }

	g, err := store.CreateGroup(ctx, "g-list", "")
	require.NoError(t, err)
	n, err := store.CreateNode(ctx, db.NodeCreate{
		Name:          "n-list",
		APIAddress:    "127.0.0.1",
		APIPort:       3003,
		SecretKey:     "secret",
		PublicAddress: "example.com",
		GroupID:       &g.ID,
	})
	require.NoError(t, err)

	current = current.Add(1 * time.Minute)
	job1, err := store.CreateSyncJob(ctx, db.SyncJobCreate{NodeID: n.ID, TriggerSource: "manual_sync_node"})
	require.NoError(t, err)
	require.NoError(t, store.UpdateSyncJobFinish(ctx, job1.ID, db.SyncJobFinishUpdate{
		Status:       db.SyncJobStatusSuccess,
		AttemptCount: 1,
		FinishedAt:   current,
		DurationMS:   100,
	}))

	current = current.Add(1 * time.Minute)
	job2, err := store.CreateSyncJob(ctx, db.SyncJobCreate{NodeID: n.ID, TriggerSource: "manual_retry"})
	require.NoError(t, err)
	require.NoError(t, store.UpdateSyncJobFinish(ctx, job2.ID, db.SyncJobFinishUpdate{
		Status:       db.SyncJobStatusFailed,
		AttemptCount: 2,
		ErrorSummary: "boom",
		FinishedAt:   current,
		DurationMS:   200,
	}))

	current = current.Add(1 * time.Minute)
	job3, err := store.CreateSyncJob(ctx, db.SyncJobCreate{NodeID: n.ID, TriggerSource: "auto_inbound_change"})
	require.NoError(t, err)

	all, err := store.ListSyncJobs(ctx, db.SyncJobsListFilter{Limit: 10})
	require.NoError(t, err)
	require.Len(t, all, 3)
	require.Equal(t, job3.ID, all[0].ID)
	require.Equal(t, job2.ID, all[1].ID)
	require.Equal(t, job1.ID, all[2].ID)

	onlyFailed, err := store.ListSyncJobs(ctx, db.SyncJobsListFilter{Status: db.SyncJobStatusFailed, Limit: 10})
	require.NoError(t, err)
	require.Len(t, onlyFailed, 1)
	require.Equal(t, job2.ID, onlyFailed[0].ID)

	onlyManualRetry, err := store.ListSyncJobs(ctx, db.SyncJobsListFilter{TriggerSource: "manual_retry", Limit: 10})
	require.NoError(t, err)
	require.Len(t, onlyManualRetry, 1)
	require.Equal(t, job2.ID, onlyManualRetry[0].ID)

	from := current.Add(-30 * time.Second)
	to := current.Add(30 * time.Second)
	windowed, err := store.ListSyncJobs(ctx, db.SyncJobsListFilter{From: &from, To: &to, Limit: 10})
	require.NoError(t, err)
	require.Len(t, windowed, 1)
	require.Equal(t, job3.ID, windowed[0].ID)

	paged, err := store.ListSyncJobs(ctx, db.SyncJobsListFilter{Limit: 1, Offset: 1})
	require.NoError(t, err)
	require.Len(t, paged, 1)
	require.Equal(t, job2.ID, paged[0].ID)
}
