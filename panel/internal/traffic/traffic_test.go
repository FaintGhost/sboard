package traffic_test

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"sboard/panel/internal/db"
	"sboard/panel/internal/traffic"
)

func setupTrafficStore(t *testing.T) *db.Store {
	t.Helper()

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	database, err := db.Open(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { _ = database.Close() })

	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)
	migrationsDir := filepath.Join(filepath.Dir(file), "../db/migrations")
	require.NoError(t, db.MigrateUp(database, migrationsDir))

	store := db.NewStore(database)
	store.Now = func() time.Time {
		return time.Date(2026, 2, 10, 12, 0, 0, 0, time.UTC)
	}
	return store
}

func mustCreateNode(t *testing.T, store *db.Store, name string) db.Node {
	t.Helper()
	node, err := store.CreateNode(context.Background(), db.NodeCreate{
		Name:          name,
		APIAddress:    "127.0.0.1",
		APIPort:       3000,
		SecretKey:     "secret",
		PublicAddress: "example.com",
	})
	require.NoError(t, err)
	return node
}

func TestSQLiteProvider_Validation(t *testing.T) {
	ctx := context.Background()
	p := traffic.NewSQLiteProvider(nil)

	_, err := p.NodesSummary(ctx, time.Hour)
	require.ErrorContains(t, err, "missing store")

	_, err = p.TotalSummary(ctx, time.Hour)
	require.ErrorContains(t, err, "missing store")

	_, err = p.Timeseries(ctx, traffic.TimeseriesQuery{Window: time.Hour, Bucket: traffic.BucketHour})
	require.ErrorContains(t, err, "missing store")
}

func TestSQLiteProvider_Aggregations(t *testing.T) {
	ctx := context.Background()
	store := setupTrafficStore(t)

	node1 := mustCreateNode(t, store, "n1")
	node2 := mustCreateNode(t, store, "n2")

	user, err := store.CreateUser(ctx, "alice")
	require.NoError(t, err)

	now := store.NowUTC()
	_, err = store.InsertInboundTrafficDelta(ctx, node1.ID, "ss-in", 100, 200, now.Add(-90*time.Minute))
	require.NoError(t, err)
	_, err = store.InsertInboundTrafficDelta(ctx, node1.ID, "vmess-in", 50, 70, now.Add(-20*time.Minute))
	require.NoError(t, err)
	_, err = store.InsertInboundTrafficDelta(ctx, node2.ID, "ss-in", 30, 40, now.Add(-3*time.Hour))
	require.NoError(t, err)

	_, err = store.InsertUserInboundTrafficDelta(ctx, user.ID, node1.ID, "ss-in", 999, 999, now.Add(-10*time.Minute))
	require.NoError(t, err)

	p := traffic.NewSQLiteProvider(store)

	nodes, err := p.NodesSummary(ctx, time.Hour)
	require.NoError(t, err)
	require.Len(t, nodes, 2)

	require.Equal(t, node1.ID, nodes[0].NodeID)
	require.Equal(t, int64(50), nodes[0].Upload)
	require.Equal(t, int64(70), nodes[0].Download)
	require.Equal(t, int64(1), nodes[0].Samples)
	require.Equal(t, int64(1), nodes[0].Inbounds)

	require.Equal(t, node2.ID, nodes[1].NodeID)
	require.Equal(t, int64(0), nodes[1].Upload)
	require.Equal(t, int64(0), nodes[1].Download)

	total, err := p.TotalSummary(ctx, time.Hour)
	require.NoError(t, err)
	require.Equal(t, int64(50), total.Upload)
	require.Equal(t, int64(70), total.Download)
	require.Equal(t, int64(1), total.Samples)
	require.Equal(t, int64(1), total.Nodes)
	require.Equal(t, int64(1), total.Inbounds)

	series, err := p.Timeseries(ctx, traffic.TimeseriesQuery{
		Window: 2 * time.Hour,
		Bucket: traffic.BucketHour,
		NodeID: node1.ID,
	})
	require.NoError(t, err)
	require.Len(t, series, 2)
	require.Equal(t, int64(100), series[0].Upload)
	require.Equal(t, int64(200), series[0].Download)
	require.Equal(t, int64(50), series[1].Upload)
	require.Equal(t, int64(70), series[1].Download)
}

func TestSQLiteProvider_InvalidQuery(t *testing.T) {
	ctx := context.Background()
	store := setupTrafficStore(t)
	p := traffic.NewSQLiteProvider(store)

	_, err := p.NodesSummary(ctx, -time.Minute)
	require.ErrorContains(t, err, "invalid window")

	_, err = p.TotalSummary(ctx, -time.Minute)
	require.ErrorContains(t, err, "invalid window")

	_, err = p.Timeseries(ctx, traffic.TimeseriesQuery{Window: -time.Minute, Bucket: traffic.BucketHour})
	require.ErrorContains(t, err, "invalid window")

	_, err = p.Timeseries(ctx, traffic.TimeseriesQuery{Window: time.Minute, Bucket: "week"})
	require.ErrorContains(t, err, "invalid bucket")
}
