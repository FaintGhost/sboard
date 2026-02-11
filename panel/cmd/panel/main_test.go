package main

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"sboard/panel/internal/config"
	"sboard/panel/internal/db"
)

func setupPanelStore(t *testing.T) *db.Store {
	t.Helper()
	dir := t.TempDir()
	database, err := db.Open(filepath.Join(dir, "test.db"))
	require.NoError(t, err)
	t.Cleanup(func() { _ = database.Close() })
	require.NoError(t, db.MigrateUp(database, ""))
	return db.NewStore(database)
}

func TestEnsureSetupTokenIfNeeded_WhenNoAdminAndNoToken(t *testing.T) {
	store := setupPanelStore(t)
	cfg := config.Config{}

	token, err := ensureSetupTokenIfNeeded(&cfg, store)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	require.Equal(t, token, cfg.SetupToken)
}

func TestEnsureSetupTokenIfNeeded_WhenNoAdminAndPresetToken(t *testing.T) {
	store := setupPanelStore(t)
	cfg := config.Config{SetupToken: "preset-token"}

	token, err := ensureSetupTokenIfNeeded(&cfg, store)
	require.NoError(t, err)
	require.Equal(t, "preset-token", token)
	require.Equal(t, "preset-token", cfg.SetupToken)
}

func TestEnsureSetupTokenIfNeeded_WhenAdminExists(t *testing.T) {
	store := setupPanelStore(t)
	created, err := db.AdminCreateIfNone(store, "admin", "hash")
	require.NoError(t, err)
	require.True(t, created)

	cfg := config.Config{}
	token, err := ensureSetupTokenIfNeeded(&cfg, store)
	require.NoError(t, err)
	require.Empty(t, token)
	require.Empty(t, cfg.SetupToken)
}
