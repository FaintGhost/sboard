package db

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMigrateUp(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	database, err := Open(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { _ = database.Close() })

	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)
	migrationsDir := filepath.Join(filepath.Dir(file), "migrations")
	err = MigrateUp(database, migrationsDir)
	require.NoError(t, err)

	rows, err := database.Query("SELECT name FROM sqlite_master WHERE type='table' AND name='users'")
	require.NoError(t, err)
	defer rows.Close()
	require.True(t, rows.Next())
}

func TestMigrateAddsUserNodes(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	database, err := Open(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { _ = database.Close() })

	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)
	migrationsDir := filepath.Join(filepath.Dir(file), "migrations")
	err = MigrateUp(database, migrationsDir)
	require.NoError(t, err)

	_, err = database.Exec("SELECT user_id, node_id FROM user_nodes LIMIT 1")
	require.NoError(t, err)

	_, err = database.Exec("SELECT api_address, api_port, public_address FROM nodes LIMIT 1")
	require.NoError(t, err)

	_, err = database.Exec("SELECT public_port FROM inbounds LIMIT 1")
	require.NoError(t, err)

	_, err = database.Exec("SELECT traffic_last_reset_at FROM users LIMIT 1")
	require.NoError(t, err)
}

func TestMigrateAddsGroups(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	database, err := Open(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { _ = database.Close() })

	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)
	migrationsDir := filepath.Join(filepath.Dir(file), "migrations")
	err = MigrateUp(database, migrationsDir)
	require.NoError(t, err)

	_, err = database.Exec("SELECT id, name, description FROM groups LIMIT 1")
	require.NoError(t, err)

	_, err = database.Exec("SELECT user_id, group_id FROM user_groups LIMIT 1")
	require.NoError(t, err)

	_, err = database.Exec("SELECT group_id FROM nodes LIMIT 1")
	require.NoError(t, err)

	_, err = database.Exec("SELECT id, node_id, trigger_source, status FROM sync_jobs LIMIT 1")
	require.NoError(t, err)

	_, err = database.Exec("SELECT parent_job_id FROM sync_jobs LIMIT 1")
	require.NoError(t, err)

	_, err = database.Exec("SELECT id, job_id, attempt_no, status FROM sync_attempts LIMIT 1")
	require.NoError(t, err)
}
