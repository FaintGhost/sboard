package monitor

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"sboard/panel/internal/db"
)

type fakeNodeClient struct {
	healthErrs []error
	syncCalls  int
}

func (f *fakeNodeClient) Health(ctx context.Context, node db.Node) error {
	if len(f.healthErrs) == 0 {
		return nil
	}
	err := f.healthErrs[0]
	f.healthErrs = f.healthErrs[1:]
	return err
}

func (f *fakeNodeClient) SyncConfig(ctx context.Context, node db.Node, payload any) error {
	f.syncCalls++
	return nil
}

func setupStore(t *testing.T) *db.Store {
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

	store := db.NewStore(database)
	// deterministic "now" for assertions
	store.Now = func() time.Time { return time.Date(2026, 2, 7, 12, 0, 0, 0, time.UTC) }
	return store
}

func TestNodesMonitor_HealthTransitionTriggersSync(t *testing.T) {
	ctx := context.Background()
	store := setupStore(t)

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
	require.Equal(t, "offline", n.Status)

	// add an active user in the same group so sync payload has users
	u, err := store.CreateUser(ctx, "alice")
	require.NoError(t, err)
	require.NoError(t, store.ReplaceUserGroups(ctx, u.ID, []int64{g.ID}))

	client := &fakeNodeClient{
		// 1st ok -> online + sync
		// 2nd ok -> still online, no extra sync
		// 3rd fail -> keep online (needs 2 consecutive failures)
		// 4th fail -> offline
		// 5th ok -> online + sync again
		healthErrs: []error{nil, nil, context.DeadlineExceeded, context.DeadlineExceeded, nil},
	}
	m := NewNodesMonitor(store, client)

	require.NoError(t, m.CheckOnce(ctx))
	got, err := store.GetNodeByID(ctx, n.ID)
	require.NoError(t, err)
	require.Equal(t, "online", got.Status)
	require.NotNil(t, got.LastSeenAt)
	require.Equal(t, 1, client.syncCalls)

	require.NoError(t, m.CheckOnce(ctx))
	require.Equal(t, 1, client.syncCalls)

	require.NoError(t, m.CheckOnce(ctx))
	got, err = store.GetNodeByID(ctx, n.ID)
	require.NoError(t, err)
	require.Equal(t, "online", got.Status)

	require.NoError(t, m.CheckOnce(ctx))
	got, err = store.GetNodeByID(ctx, n.ID)
	require.NoError(t, err)
	require.Equal(t, "offline", got.Status)

	require.NoError(t, m.CheckOnce(ctx))
	got, err = store.GetNodeByID(ctx, n.ID)
	require.NoError(t, err)
	require.Equal(t, "online", got.Status)
	require.Equal(t, 2, client.syncCalls)
}

func TestNodesMonitor_FirstHealthyCheckSyncsWhenNodeAlreadyOnline(t *testing.T) {
	ctx := context.Background()
	store := setupStore(t)

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

	_, err = store.CreateInbound(ctx, db.InboundCreate{
		Tag:        "ss-in",
		NodeID:     n.ID,
		Protocol:   "shadowsocks",
		ListenPort: 8388,
		PublicPort: 8388,
		Settings:   []byte(`{"method":"2022-blake3-aes-128-gcm"}`),
	})
	require.NoError(t, err)

	u, err := store.CreateUser(ctx, "alice")
	require.NoError(t, err)
	require.NoError(t, store.ReplaceUserGroups(ctx, u.ID, []int64{g.ID}))

	require.NoError(t, store.MarkNodeOnline(ctx, n.ID, store.Now().UTC()))

	client := &fakeNodeClient{healthErrs: []error{nil, nil}}
	m := NewNodesMonitor(store, client)

	require.NoError(t, m.CheckOnce(ctx))
	require.Equal(t, 1, client.syncCalls)

	require.NoError(t, m.CheckOnce(ctx))
	require.Equal(t, 1, client.syncCalls)
}
