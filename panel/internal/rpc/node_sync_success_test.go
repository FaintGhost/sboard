package rpc

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"sboard/panel/internal/db"
	nodev1 "sboard/panel/internal/rpc/gen/sboard/node/v1"
	nodev1connect "sboard/panel/internal/rpc/gen/sboard/node/v1/nodev1connect"
)

type nodeControlServiceStub struct {
	mu      sync.Mutex
	payload []byte
	calls   int
}

func (s *nodeControlServiceStub) Health(ctx context.Context, req *nodev1.HealthRequest) (*nodev1.HealthResponse, error) {
	return &nodev1.HealthResponse{Status: "ok"}, nil
}

func (s *nodeControlServiceStub) SyncConfig(ctx context.Context, req *nodev1.SyncConfigRequest) (*nodev1.SyncConfigResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls++
	s.payload = append([]byte(nil), req.GetPayloadJson()...)
	return &nodev1.SyncConfigResponse{Status: "ok"}, nil
}

func (s *nodeControlServiceStub) GetTraffic(ctx context.Context, req *nodev1.GetTrafficRequest) (*nodev1.GetTrafficResponse, error) {
	return &nodev1.GetTrafficResponse{}, nil
}

func (s *nodeControlServiceStub) GetInboundTraffic(ctx context.Context, req *nodev1.GetInboundTrafficRequest) (*nodev1.GetInboundTrafficResponse, error) {
	return &nodev1.GetInboundTrafficResponse{}, nil
}

func setupRPCStore(t *testing.T) *db.Store {
	t.Helper()

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	database, err := db.Open(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { _ = database.Close() })

	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)
	migrationsDir := filepath.Join(filepath.Dir(file), "..", "db", "migrations")
	require.NoError(t, db.MigrateUp(database, migrationsDir))
	return db.NewStore(database)
}

func TestNodeSyncRPCSuccess(t *testing.T) {
	ctx := context.Background()
	store := setupRPCStore(t)

	stub := &nodeControlServiceStub{}
	path, handler := nodev1connect.NewNodeControlServiceHandler(stub)
	_ = path

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

	group, err := store.CreateGroup(ctx, "g-1", "")
	require.NoError(t, err)

	user, err := store.CreateUser(ctx, "alice")
	require.NoError(t, err)
	require.NoError(t, store.ReplaceGroupUsers(ctx, group.ID, []int64{user.ID}))

	groupID := group.ID
	nodeItem, err := store.CreateNode(ctx, db.NodeCreate{
		Name:          "node-a",
		APIAddress:    parsedURL.Scheme + "://" + host,
		APIPort:       port,
		SecretKey:     "secret",
		PublicAddress: host,
		GroupID:       &groupID,
	})
	require.NoError(t, err)

	_, err = store.CreateInbound(ctx, db.InboundCreate{
		NodeID:     nodeItem.ID,
		Tag:        "vless-in",
		Protocol:   "vless",
		ListenPort: 443,
		PublicPort: 443,
		Settings:   json.RawMessage(`{}`),
	})
	require.NoError(t, err)

	server := &Server{store: store}
	result := server.runNodeSync(ctx, nodeItem, rpcTriggerManualNodeSync, nil)
	require.Equal(t, "ok", result.Status)

	stub.mu.Lock()
	defer stub.mu.Unlock()
	require.Equal(t, 1, stub.calls)
	require.NotEmpty(t, stub.payload)
}
