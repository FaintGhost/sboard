package db_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"sboard/panel/internal/db"
)

func mustCreateNodeInDBTests(t *testing.T, store *db.Store, name string) db.Node {
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

func TestTrafficAggregates_FilterAndTimeseries(t *testing.T) {
	store := setupStore(t)
	ctx := context.Background()

	node1 := mustCreateNodeInDBTests(t, store, "n1")
	node2 := mustCreateNodeInDBTests(t, store, "n2")
	user, err := store.CreateUser(ctx, "alice")
	require.NoError(t, err)

	t1 := time.Date(2026, 2, 10, 10, 10, 0, 0, time.UTC)
	t2 := time.Date(2026, 2, 10, 10, 40, 0, 0, time.UTC)
	t3 := time.Date(2026, 2, 10, 11, 5, 0, 0, time.UTC)

	_, err = store.InsertInboundTrafficDelta(ctx, node1.ID, "ss-in", 10, 20, t1)
	require.NoError(t, err)
	_, err = store.InsertInboundTrafficDelta(ctx, node1.ID, "vmess-in", 30, 40, t2)
	require.NoError(t, err)
	_, err = store.InsertInboundTrafficDelta(ctx, node2.ID, "ss-in", 5, 6, t3)
	require.NoError(t, err)
	// 用户级流量不应计入 node/global 聚合。
	_, err = store.InsertUserInboundTrafficDelta(ctx, user.ID, node1.ID, "ss-in", 100, 100, t3)
	require.NoError(t, err)

	summaries, err := store.ListNodeTrafficSummaries(ctx, time.Time{})
	require.NoError(t, err)
	require.Len(t, summaries, 2)

	require.Equal(t, node1.ID, summaries[0].NodeID)
	require.Equal(t, int64(40), summaries[0].Upload)
	require.Equal(t, int64(60), summaries[0].Download)
	require.Equal(t, int64(2), summaries[0].Samples)
	require.Equal(t, int64(2), summaries[0].Inbounds)
	require.WithinDuration(t, t2, summaries[0].LastRecordedAt, time.Second)

	require.Equal(t, node2.ID, summaries[1].NodeID)
	require.Equal(t, int64(5), summaries[1].Upload)
	require.Equal(t, int64(6), summaries[1].Download)
	require.Equal(t, int64(1), summaries[1].Samples)
	require.Equal(t, int64(1), summaries[1].Inbounds)
	require.WithinDuration(t, t3, summaries[1].LastRecordedAt, time.Second)

	total, err := store.GetTrafficTotalSummary(ctx, time.Time{})
	require.NoError(t, err)
	require.Equal(t, int64(45), total.Upload)
	require.Equal(t, int64(66), total.Download)
	require.Equal(t, int64(3), total.Samples)
	require.Equal(t, int64(2), total.Nodes)
	require.Equal(t, int64(2), total.Inbounds)
	require.WithinDuration(t, t3, total.LastRecordedAt, time.Second)

	totalRecent, err := store.GetTrafficTotalSummary(ctx, t3)
	require.NoError(t, err)
	require.Equal(t, int64(5), totalRecent.Upload)
	require.Equal(t, int64(6), totalRecent.Download)
	require.Equal(t, int64(1), totalRecent.Samples)
	require.Equal(t, int64(1), totalRecent.Nodes)
	require.Equal(t, int64(1), totalRecent.Inbounds)

	node1Series, err := store.ListTrafficTimeseries(ctx, time.Time{}, node1.ID, "hour")
	require.NoError(t, err)
	require.Len(t, node1Series, 1)
	require.Equal(t, int64(40), node1Series[0].Upload)
	require.Equal(t, int64(60), node1Series[0].Download)
	require.WithinDuration(t, time.Date(2026, 2, 10, 10, 0, 0, 0, time.UTC), node1Series[0].BucketStart, time.Second)

	allSeries, err := store.ListTrafficTimeseries(ctx, time.Time{}, 0, "hour")
	require.NoError(t, err)
	require.Len(t, allSeries, 2)
	require.Equal(t, int64(40), allSeries[0].Upload)
	require.Equal(t, int64(60), allSeries[0].Download)
	require.Equal(t, int64(5), allSeries[1].Upload)
	require.Equal(t, int64(6), allSeries[1].Download)

	_, err = store.ListTrafficTimeseries(ctx, time.Time{}, 0, "week")
	require.ErrorContains(t, err, "invalid bucket")
}

func TestTrafficStatsCRUDAndValidation(t *testing.T) {
	store := setupStore(t)
	ctx := context.Background()
	store.Now = func() time.Time {
		return time.Date(2026, 2, 10, 12, 0, 0, 0, time.UTC)
	}

	node := mustCreateNodeInDBTests(t, store, "n1")
	user, err := store.CreateUser(ctx, "bob")
	require.NoError(t, err)

	_, err = store.InsertNodeTrafficSample(ctx, 0, 1, 2, time.Time{})
	require.ErrorContains(t, err, "invalid node_id")
	_, err = store.InsertInboundTrafficDelta(ctx, node.ID, "", 1, 2, time.Time{})
	require.ErrorContains(t, err, "missing inbound_tag")
	_, err = store.InsertUserInboundTrafficDelta(ctx, 0, node.ID, "ss-in", 1, 2, time.Time{})
	require.ErrorContains(t, err, "invalid user_id")

	require.ErrorContains(t, store.AddUserTrafficUsed(ctx, 0, 100), "invalid user_id")
	require.NoError(t, store.AddUserTrafficUsed(ctx, user.ID, 0))
	require.NoError(t, store.AddUserTrafficUsed(ctx, user.ID, 123))
	updatedUser, err := store.GetUserByID(ctx, user.ID)
	require.NoError(t, err)
	require.Equal(t, int64(123), updatedUser.TrafficUsed)
	require.ErrorIs(t, store.AddUserTrafficUsed(ctx, 999999, 1), db.ErrNotFound)

	nodeSample, err := store.InsertNodeTrafficSample(ctx, node.ID, 10, 20, time.Time{})
	require.NoError(t, err)
	require.NotNil(t, nodeSample.NodeID)
	require.Equal(t, node.ID, *nodeSample.NodeID)
	require.Nil(t, nodeSample.UserID)
	require.True(t, nodeSample.RecordedAt.Equal(store.NowUTC()))

	inboundSample, err := store.InsertInboundTrafficDelta(ctx, node.ID, "ss-in", 30, 40, store.NowUTC().Add(-time.Minute))
	require.NoError(t, err)
	require.NotNil(t, inboundSample.InboundTag)
	require.Equal(t, "ss-in", *inboundSample.InboundTag)

	userSample, err := store.InsertUserInboundTrafficDelta(ctx, user.ID, node.ID, "ss-in", 50, 60, store.NowUTC())
	require.NoError(t, err)
	require.NotNil(t, userSample.UserID)
	require.Equal(t, user.ID, *userSample.UserID)

	got, err := store.GetTrafficStatByID(ctx, inboundSample.ID)
	require.NoError(t, err)
	require.Equal(t, inboundSample.ID, got.ID)

	_, err = store.GetTrafficStatByID(ctx, 999999)
	require.ErrorIs(t, err, db.ErrNotFound)

	_, err = store.ListNodeTrafficSamples(ctx, 0, 10, 0)
	require.ErrorContains(t, err, "invalid node_id")
	nodeSamples, err := store.ListNodeTrafficSamples(ctx, node.ID, 10, 0)
	require.NoError(t, err)
	require.Len(t, nodeSamples, 2)
	for _, sample := range nodeSamples {
		require.Nil(t, sample.UserID)
	}
	require.Equal(t, nodeSample.ID, nodeSamples[0].ID)
	require.Equal(t, inboundSample.ID, nodeSamples[1].ID)
}
