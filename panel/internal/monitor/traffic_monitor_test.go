package monitor

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"sboard/panel/internal/db"
	"sboard/panel/internal/node"
)

type trafficMonitorFakeDoer struct {
	body string
}

func (d *trafficMonitorFakeDoer) Do(req *http.Request) (*http.Response, error) {
	if req.URL.Path != "/api/stats/inbounds" {
		return &http.Response{
			StatusCode: http.StatusNotFound,
			Body:       io.NopCloser(strings.NewReader("not found")),
			Header:     http.Header{"Content-Type": []string{"application/json"}},
		}, nil
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(d.body)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
	}, nil
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
	doer := &trafficMonitorFakeDoer{body: `{"data":[{"tag":"ss-in","user":"alice","uplink":900,"downlink":300,"at":"` + now + `"},{"tag":"ss-in","user":"bob","uplink":120,"downlink":80,"at":"` + now + `"}],"reset":true,"meta":{"tracked_tags":1,"tcp_conns":1,"udp_conns":0}}`}
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
