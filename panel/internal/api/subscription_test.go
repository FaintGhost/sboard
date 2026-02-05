package api_test

import (
  "encoding/base64"
  "net/http"
  "net/http/httptest"
  "path/filepath"
  "runtime"
  "strings"
  "testing"

  "sboard/panel/internal/api"
  "sboard/panel/internal/config"
  "sboard/panel/internal/db"
  "github.com/google/uuid"
  "github.com/stretchr/testify/require"
)

func setupSubscriptionStore(t *testing.T) *db.Store {
  t.Helper()
  dir := t.TempDir()
  dbPath := filepath.Join(dir, "test.db")
  database, err := db.Open(dbPath)
  require.NoError(t, err)
  t.Cleanup(func() { _ = database.Close() })

  _, file, _, ok := runtime.Caller(0)
  require.True(t, ok)
  migrationsDir := filepath.Join(filepath.Dir(file), "..", "db", "migrations")
  err = db.MigrateUp(database, migrationsDir)
  require.NoError(t, err)

  return db.NewStore(database)
}

func seedSubscriptionData(t *testing.T, store *db.Store) string {
  t.Helper()
  user, err := store.CreateUser(t.Context(), "alice")
  require.NoError(t, err)

  res, err := store.DB.Exec(
    "INSERT INTO nodes (uuid, name, address, port, secret_key, api_address, api_port, public_address) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
    uuid.NewString(),
    "node-a",
    "api.local",
    2222,
    "secret",
    "api.local",
    2222,
    "a.example.com",
  )
  require.NoError(t, err)
  nodeID, err := res.LastInsertId()
  require.NoError(t, err)

  _, err = store.DB.Exec(
    "INSERT INTO inbounds (uuid, tag, node_id, protocol, listen_port, settings, public_port) VALUES (?, ?, ?, ?, ?, ?, ?)",
    uuid.NewString(),
    "vless-in",
    nodeID,
    "vless",
    443,
    `{}`,
    0,
  )
  require.NoError(t, err)

  _, err = store.DB.Exec("INSERT INTO user_nodes (user_id, node_id) VALUES (?, ?)", user.ID, nodeID)
  require.NoError(t, err)

  return user.UUID
}

func TestSubscriptionUAAndFormat(t *testing.T) {
  store := setupSubscriptionStore(t)
  userUUID := seedSubscriptionData(t, store)

  r := api.NewRouter(config.Config{}, store)

  req := httptest.NewRequest(http.MethodGet, "/api/sub/"+userUUID+"?format=singbox", nil)
  req.Header.Set("User-Agent", "clash-meta")
  w := httptest.NewRecorder()
  r.ServeHTTP(w, req)
  require.Equal(t, http.StatusOK, w.Code)
  require.Contains(t, w.Header().Get("Content-Type"), "application/json")

  req = httptest.NewRequest(http.MethodGet, "/api/sub/"+userUUID, nil)
  req.Header.Set("User-Agent", "Shadowrocket")
  w = httptest.NewRecorder()
  r.ServeHTTP(w, req)
  require.Equal(t, http.StatusOK, w.Code)
  require.NotContains(t, w.Header().Get("Content-Type"), "application/json")

  decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(w.Body.String()))
  require.NoError(t, err)
  require.Contains(t, string(decoded), "a.example.com")
}
