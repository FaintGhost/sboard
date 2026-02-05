package db_test

import (
  "context"
  "errors"
  "path/filepath"
  "runtime"
  "testing"

  "sboard/panel/internal/db"
  "github.com/stretchr/testify/require"
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
