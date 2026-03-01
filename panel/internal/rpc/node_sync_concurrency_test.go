package rpc

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	nodev1 "sboard/panel/internal/rpc/gen/sboard/node/v1"
	nodev1connect "sboard/panel/internal/rpc/gen/sboard/node/v1/nodev1connect"
)

type nodeControlConcurrencyStub struct {
	active int32
	max    int32
	calls  int32
}

func (s *nodeControlConcurrencyStub) Health(ctx context.Context, req *nodev1.HealthRequest) (*nodev1.HealthResponse, error) {
	return &nodev1.HealthResponse{Status: "ok"}, nil
}

func (s *nodeControlConcurrencyStub) SyncConfig(ctx context.Context, req *nodev1.SyncConfigRequest) (*nodev1.SyncConfigResponse, error) {
	current := atomic.AddInt32(&s.active, 1)
	defer atomic.AddInt32(&s.active, -1)
	atomic.AddInt32(&s.calls, 1)

	for {
		maxSeen := atomic.LoadInt32(&s.max)
		if current <= maxSeen {
			break
		}
		if atomic.CompareAndSwapInt32(&s.max, maxSeen, current) {
			break
		}
	}

	time.Sleep(120 * time.Millisecond)
	return &nodev1.SyncConfigResponse{Status: "ok"}, nil
}

func (s *nodeControlConcurrencyStub) GetTraffic(ctx context.Context, req *nodev1.GetTrafficRequest) (*nodev1.GetTrafficResponse, error) {
	return &nodev1.GetTrafficResponse{}, nil
}

func (s *nodeControlConcurrencyStub) GetInboundTraffic(ctx context.Context, req *nodev1.GetInboundTrafficRequest) (*nodev1.GetInboundTrafficResponse, error) {
	return &nodev1.GetInboundTrafficResponse{}, nil
}

func TestNodeSyncRPCSerializesPerNode(t *testing.T) {
	ctx := context.Background()
	store := setupRPCStore(t)

	stub := &nodeControlConcurrencyStub{}
	_, handler := nodev1connect.NewNodeControlServiceHandler(stub)
	mux := http.NewServeMux()
	mux.Handle("/rpc/", http.StripPrefix("/rpc", handler))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	nodeItem := seedSyncNodeFixture(t, ctx, store, srv.URL, "secret")
	server := &Server{store: store}

	var wg sync.WaitGroup
	results := make(chan syncResult, 2)

	wg.Add(2)
	for i := 0; i < 2; i++ {
		go func() {
			defer wg.Done()
			results <- server.runNodeSync(ctx, nodeItem, rpcTriggerManualNodeSync, nil)
		}()
	}
	wg.Wait()
	close(results)

	for result := range results {
		require.Equal(t, "ok", result.Status)
	}

	require.Equal(t, int32(2), atomic.LoadInt32(&stub.calls))
	require.Equal(t, int32(1), atomic.LoadInt32(&stub.max))
}
