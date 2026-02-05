package db_test

import (
  "context"
  "path/filepath"
  "runtime"
  "testing"

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
  migrationsDir := filepath.Join(filepath.Dir(file), "migrations")
  err = db.MigrateUp(database, migrationsDir)
  require.NoError(t, err)

  return db.NewStore(database)
}

func insertNode(t *testing.T, store *db.Store, name, apiAddr string, apiPort int, publicAddr string) int64 {
  t.Helper()
  res, err := store.DB.Exec(
    "INSERT INTO nodes (uuid, name, address, port, secret_key, api_address, api_port, public_address) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
    uuid.NewString(),
    name,
    apiAddr,
    apiPort,
    "secret",
    apiAddr,
    apiPort,
    publicAddr,
  )
  require.NoError(t, err)
  id, err := res.LastInsertId()
  require.NoError(t, err)
  return id
}

func insertInbound(t *testing.T, store *db.Store, nodeID int64, typ string, listenPort int, publicPort int) int64 {
  t.Helper()
  res, err := store.DB.Exec(
    "INSERT INTO inbounds (uuid, tag, node_id, protocol, listen_port, settings, public_port) VALUES (?, ?, ?, ?, ?, ?, ?)",
    uuid.NewString(),
    typ+"-in",
    nodeID,
    typ,
    listenPort,
    `{"flow":"xtls-rprx-vision"}`,
    publicPort,
  )
  require.NoError(t, err)
  id, err := res.LastInsertId()
  require.NoError(t, err)
  return id
}

func insertUser(t *testing.T, store *db.Store, username string) int64 {
  t.Helper()
  user, err := store.CreateUser(context.Background(), username)
  require.NoError(t, err)
  return user.ID
}

func bindUserNode(t *testing.T, store *db.Store, userID, nodeID int64) {
  t.Helper()
  _, err := store.DB.Exec("INSERT INTO user_nodes (user_id, node_id) VALUES (?, ?)", userID, nodeID)
  require.NoError(t, err)
}

func TestSubscriptionQuery(t *testing.T) {
  store := setupSubscriptionStore(t)
  ctx := context.Background()

  nodeID := insertNode(t, store, "node-a", "api.local", 2222, "a.example.com")
  _ = insertInbound(t, store, nodeID, "vless", 443, 0)
  userID := insertUser(t, store, "alice")
  bindUserNode(t, store, userID, nodeID)

  got, err := store.ListUserInbounds(ctx, userID)
  require.NoError(t, err)
  require.Len(t, got, 1)
  require.Equal(t, "vless", got[0].InboundType)
  require.Equal(t, "a.example.com", got[0].NodePublicAddress)
}
