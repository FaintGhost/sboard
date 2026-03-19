package monitor

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"sboard/panel/internal/db"
	"sboard/panel/internal/node"
	nodev1 "sboard/panel/internal/rpc/gen/sboard/node/v1"
	nodev1connect "sboard/panel/internal/rpc/gen/sboard/node/v1/nodev1connect"
)

type nodeTelemetryRPCStub struct {
	mu        sync.Mutex
	syncCalls int
	resetSeen bool
}

func (s *nodeTelemetryRPCStub) Health(ctx context.Context, req *nodev1.HealthRequest) (*nodev1.HealthResponse, error) {
	return &nodev1.HealthResponse{Status: "ok"}, nil
}

func (s *nodeTelemetryRPCStub) SyncConfig(ctx context.Context, req *nodev1.SyncConfigRequest) (*nodev1.SyncConfigResponse, error) {
	s.mu.Lock()
	s.syncCalls++
	s.mu.Unlock()
	return &nodev1.SyncConfigResponse{Status: "ok"}, nil
}

func (s *nodeTelemetryRPCStub) GetTraffic(ctx context.Context, req *nodev1.GetTrafficRequest) (*nodev1.GetTrafficResponse, error) {
	return &nodev1.GetTrafficResponse{
		Interface: "eth0",
		RxBytes:   100,
		TxBytes:   200,
		At:        "2026-02-12T00:00:00Z",
	}, nil
}

func (s *nodeTelemetryRPCStub) GetInboundTraffic(ctx context.Context, req *nodev1.GetInboundTrafficRequest) (*nodev1.GetInboundTrafficResponse, error) {
	s.mu.Lock()
	s.resetSeen = req.GetReset_()
	s.mu.Unlock()
	return &nodev1.GetInboundTrafficResponse{
		Data: []*nodev1.InboundTraffic{{
			Tag:      "ss-in",
			User:     "alice",
			Uplink:   15,
			Downlink: 25,
			At:       "2026-02-12T00:00:00Z",
		}},
		Meta: &nodev1.InboundTrafficMeta{
			TrackedTags: 1,
			TcpConns:    1,
			UdpConns:    0,
		},
	}, nil
}

func TestNodeRPCTelemetryMonitor(t *testing.T) {
	ctx := context.Background()
	store := setupStore(t)

	stub := &nodeTelemetryRPCStub{}
	_, handler := nodev1connect.NewNodeControlServiceHandler(stub)
	mux := http.NewServeMux()
	mux.Handle("/rpc/", http.StripPrefix("/rpc", handler))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	parsedURL, err := url.Parse(srv.URL)
	require.NoError(t, err)
	host, portRaw, err := net.SplitHostPort(parsedURL.Host)
	require.NoError(t, err)
	port, err := strconv.Atoi(portRaw)
	require.NoError(t, err)

	group, err := store.CreateGroup(ctx, "g-rpc", "")
	require.NoError(t, err)
	user, err := store.CreateUser(ctx, "alice")
	require.NoError(t, err)
	require.NoError(t, store.ReplaceUserGroups(ctx, user.ID, []int64{group.ID}))

	nodeItem, err := store.CreateNode(ctx, db.NodeCreate{
		Name:          "node-rpc",
		APIAddress:    parsedURL.Scheme + "://" + host,
		APIPort:       port,
		SecretKey:     "secret",
		PublicAddress: host,
		GroupID:       &group.ID,
	})
	require.NoError(t, err)

	_, err = store.CreateInbound(ctx, db.InboundCreate{
		NodeID:     nodeItem.ID,
		Tag:        "ss-in",
		Protocol:   "shadowsocks",
		ListenPort: 8388,
		PublicPort: 8388,
		Settings:   []byte(`{"method":"2022-blake3-aes-128-gcm"}`),
	})
	require.NoError(t, err)

	client := node.NewClient(srv.Client())
	nodesMonitor := NewNodesMonitor(store, client)
	require.NoError(t, nodesMonitor.CheckOnce(ctx))

	gotNode, err := store.GetNodeByID(ctx, nodeItem.ID)
	require.NoError(t, err)
	require.Equal(t, "online", gotNode.Status)

	stub.mu.Lock()
	require.Equal(t, 1, stub.syncCalls)
	stub.mu.Unlock()

	trafficMonitor := NewTrafficMonitor(store, client)
	require.NoError(t, trafficMonitor.SampleOnce(ctx))

	gotUser, err := store.GetUserByID(ctx, user.ID)
	require.NoError(t, err)
	require.Equal(t, int64(40), gotUser.TrafficUsed)

	stub.mu.Lock()
	require.True(t, stub.resetSeen)
	stub.mu.Unlock()
}
