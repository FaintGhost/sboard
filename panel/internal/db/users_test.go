package db_test

import (
	"context"
	"errors"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"sboard/panel/internal/db"
)

func setupStore(t *testing.T) *db.Store {
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

	return db.NewStore(database)
}

func TestUserCreateAndUnique(t *testing.T) {
	store := setupStore(t)
	user, err := store.CreateUser(context.Background(), "alice")
	require.NoError(t, err)
	require.NotEmpty(t, user.UUID)

	_, err = store.CreateUser(context.Background(), "alice")
	require.Error(t, err)
	require.True(t, errors.Is(err, db.ErrConflict))
}

func TestUserUpdateExpireAt(t *testing.T) {
	store := setupStore(t)
	user, err := store.CreateUser(context.Background(), "alice")
	require.NoError(t, err)

	exp := time.Now().UTC().Truncate(time.Second)
	updated, err := store.UpdateUser(context.Background(), user.ID, db.UserUpdate{
		ExpireAt:    &exp,
		ExpireAtSet: true,
	})
	require.NoError(t, err)
	require.NotNil(t, updated.ExpireAt)
	require.WithinDuration(t, exp, *updated.ExpireAt, time.Second)

	updated, err = store.UpdateUser(context.Background(), user.ID, db.UserUpdate{
		ExpireAtSet: true,
	})
	require.NoError(t, err)
	require.Nil(t, updated.ExpireAt)
}

func TestListUsersIncludesDisabledByDefault(t *testing.T) {
	store := setupStore(t)
	user, err := store.CreateUser(context.Background(), "alice")
	require.NoError(t, err)

	require.NoError(t, store.DisableUser(context.Background(), user.ID))

	listed, err := store.ListUsers(context.Background(), 10, 0, "")
	require.NoError(t, err)
	require.Len(t, listed, 1)
	require.Equal(t, "disabled", listed[0].Status)
}

func TestTrafficResetClampsToLastDayAndResetsAfterBoundary(t *testing.T) {
	store := setupStore(t)
	now := time.Date(2026, 2, 28, 12, 0, 0, 0, time.UTC)
	store.Now = func() time.Time { return now }

	user, err := store.CreateUser(context.Background(), "alice")
	require.NoError(t, err)

	lastReset := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
	_, err = store.DB.Exec(
		"UPDATE users SET traffic_used = ?, traffic_reset_day = ?, traffic_last_reset_at = ? WHERE id = ?",
		int64(123),
		31,
		lastReset.Format(time.RFC3339),
		user.ID,
	)
	require.NoError(t, err)

	got, err := store.GetUserByID(context.Background(), user.ID)
	require.NoError(t, err)
	require.Equal(t, int64(0), got.TrafficUsed)
	require.NotNil(t, got.TrafficLastResetAt)
	require.Equal(t, time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC), *got.TrafficLastResetAt)
}

func TestTrafficResetDoesNotRunBeforeBoundary(t *testing.T) {
	store := setupStore(t)
	now := time.Date(2026, 2, 6, 12, 0, 0, 0, time.UTC)
	store.Now = func() time.Time { return now }

	user, err := store.CreateUser(context.Background(), "alice")
	require.NoError(t, err)

	lastReset := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
	_, err = store.DB.Exec(
		"UPDATE users SET traffic_used = ?, traffic_reset_day = ?, traffic_last_reset_at = ? WHERE id = ?",
		int64(123),
		31,
		lastReset.Format(time.RFC3339),
		user.ID,
	)
	require.NoError(t, err)

	got, err := store.GetUserByID(context.Background(), user.ID)
	require.NoError(t, err)
	require.Equal(t, int64(123), got.TrafficUsed)
}

func TestTrafficResetInitDoesNotWipeUsageWhenLastResetNull(t *testing.T) {
	store := setupStore(t)
	now := time.Date(2026, 2, 6, 12, 0, 0, 0, time.UTC)
	store.Now = func() time.Time { return now }

	user, err := store.CreateUser(context.Background(), "alice")
	require.NoError(t, err)

	_, err = store.DB.Exec(
		"UPDATE users SET traffic_used = ?, traffic_reset_day = ?, traffic_last_reset_at = NULL WHERE id = ?",
		int64(123),
		1,
		user.ID,
	)
	require.NoError(t, err)

	got, err := store.GetUserByID(context.Background(), user.ID)
	require.NoError(t, err)
	require.Equal(t, int64(123), got.TrafficUsed)
	require.NotNil(t, got.TrafficLastResetAt)
}
