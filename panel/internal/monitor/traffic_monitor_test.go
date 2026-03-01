package monitor

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"sboard/panel/internal/db"
	"sboard/panel/internal/node"
	nodev1 "sboard/panel/internal/rpc/gen/sboard/node/v1"
)

type trafficMonitorFakeDoer struct {
	resp *nodev1.GetInboundTrafficResponse
}

func (d *trafficMonitorFakeDoer) Do(req *http.Request) (*http.Response, error) {
	return serveNodeRPCRequest(req, nodeRPCServiceStub{
		inboundTrafficFunc: func(context.Context, *nodev1.GetInboundTrafficRequest) (*nodev1.GetInboundTrafficResponse, error) {
			return d.resp, nil
		},
	})
}

func TestTrafficMonitor_SampleOnce_AccumulatesUserTrafficUsed(t *testing.T) {
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

	alice, err := store.CreateUser(ctx, "alice")
	require.NoError(t, err)
	bob, err := store.CreateUser(ctx, "bob")
	require.NoError(t, err)
	require.NoError(t, store.ReplaceUserGroups(ctx, alice.ID, []int64{g.ID}))
	require.NoError(t, store.ReplaceUserGroups(ctx, bob.ID, []int64{g.ID}))

	now := time.Date(2026, 2, 8, 4, 0, 0, 0, time.UTC).Format(time.RFC3339)
	doer := &trafficMonitorFakeDoer{resp: &nodev1.GetInboundTrafficResponse{
		Data: []*nodev1.InboundTraffic{
			{Tag: "ss-in", User: "alice", Uplink: 900, Downlink: 300, At: now},
			{Tag: "ss-in", User: "bob", Uplink: 120, Downlink: 80, At: now},
		},
		Reset_: true,
		Meta:   &nodev1.InboundTrafficMeta{TrackedTags: 1, TcpConns: 1, UdpConns: 0},
	}}
	m := NewTrafficMonitor(store, node.NewClient(doer))

	require.NoError(t, m.SampleOnce(ctx))

	gotAlice, err := store.GetUserByID(ctx, alice.ID)
	require.NoError(t, err)
	gotBob, err := store.GetUserByID(ctx, bob.ID)
	require.NoError(t, err)
	require.Equal(t, int64(1200), gotAlice.TrafficUsed)
	require.Equal(t, int64(200), gotBob.TrafficUsed)

	row := store.DB.QueryRowContext(ctx, "SELECT COUNT(1) FROM traffic_stats WHERE user_id IS NOT NULL")
	var userTrafficRows int
	require.NoError(t, row.Scan(&userTrafficRows))
	require.Equal(t, 2, userTrafficRows)
}
