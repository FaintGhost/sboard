package node

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	nodev1 "sboard/panel/internal/rpc/gen/sboard/node/v1"
)

func TestRPCClientTelemetry(t *testing.T) {
	var resetSeen bool
	srv := newNodeRPCServer(t, nodeControlServiceTestServer{
		getTraffic: func(ctx context.Context, req *nodev1.GetTrafficRequest) (*nodev1.GetTrafficResponse, error) {
			return &nodev1.GetTrafficResponse{
				Interface: "eth0",
				RxBytes:   321,
				TxBytes:   654,
				At:        "2026-02-11T12:34:56Z",
			}, nil
		},
		inboundTraffic: func(ctx context.Context, req *nodev1.GetInboundTrafficRequest) (*nodev1.GetInboundTrafficResponse, error) {
			resetSeen = req.GetReset_()
			return &nodev1.GetInboundTrafficResponse{
				Data: []*nodev1.InboundTraffic{{
					Tag:      "ss-in",
					User:     "alice",
					Uplink:   11,
					Downlink: 22,
					At:       "2026-02-11T12:34:56Z",
				}},
				Meta: &nodev1.InboundTrafficMeta{
					TrackedTags: 1,
					TcpConns:    2,
					UdpConns:    3,
				},
			}, nil
		},
	}, nil)
	defer srv.Close()

	client := NewClient(srv.Client())
	nodeItem := nodeFromServerURL(t, srv.URL)

	sample, err := client.Traffic(context.Background(), nodeItem)
	require.NoError(t, err)
	require.Equal(t, "eth0", sample.Interface)
	require.Equal(t, uint64(321), sample.RxBytes)
	require.Equal(t, uint64(654), sample.TxBytes)
	require.Equal(t, time.Date(2026, 2, 11, 12, 34, 56, 0, time.UTC), sample.At)

	rows, meta, err := client.InboundTrafficWithMeta(context.Background(), nodeItem, true)
	require.NoError(t, err)
	require.True(t, resetSeen)
	require.Len(t, rows, 1)
	require.Equal(t, "ss-in", rows[0].Tag)
	require.Equal(t, "alice", rows[0].User)
	require.Equal(t, int64(11), rows[0].Uplink)
	require.Equal(t, int64(22), rows[0].Downlink)
	require.NotNil(t, meta)
	require.Equal(t, 1, meta.TrackedTags)
	require.Equal(t, int64(2), meta.TCPConns)
	require.Equal(t, int64(3), meta.UDPConns)
}

func TestRPCClientConcurrentSyncOrdering(t *testing.T) {
	started := make(chan int, 2)
	releaseFirst := make(chan struct{})
	doneSecond := make(chan struct{})

	var (
		mu            sync.Mutex
		callOrder     []int
		inFlight      int
		maxConcurrent int
	)

	srv := newNodeRPCServer(t, nodeControlServiceTestServer{
		syncConfig: func(ctx context.Context, req *nodev1.SyncConfigRequest) (*nodev1.SyncConfigResponse, error) {
			order := 0
			if string(req.GetPayloadJson()) == `{"order":1}` {
				order = 1
			} else if string(req.GetPayloadJson()) == `{"order":2}` {
				order = 2
			} else {
				return nil, errors.New("unexpected payload")
			}

			mu.Lock()
			inFlight++
			if inFlight > maxConcurrent {
				maxConcurrent = inFlight
			}
			callOrder = append(callOrder, order)
			mu.Unlock()

			started <- order
			if order == 1 {
				<-releaseFirst
			} else {
				close(doneSecond)
			}

			mu.Lock()
			inFlight--
			mu.Unlock()
			return &nodev1.SyncConfigResponse{Status: "ok"}, nil
		},
	}, nil)
	defer srv.Close()

	client := NewClient(srv.Client())
	nodeItem := nodeFromServerURL(t, srv.URL)

	errs := make(chan error, 2)
	go func() {
		errs <- client.SyncConfig(context.Background(), nodeItem, map[string]any{"order": 1})
	}()
	<-started

	go func() {
		errs <- client.SyncConfig(context.Background(), nodeItem, map[string]any{"order": 2})
	}()

	select {
	case <-doneSecond:
		t.Fatal("second sync completed before first was released")
	case <-time.After(150 * time.Millisecond):
	}

	close(releaseFirst)
	<-started

	require.NoError(t, <-errs)
	require.NoError(t, <-errs)

	mu.Lock()
	defer mu.Unlock()
	require.Equal(t, []int{1, 2}, callOrder)
	require.Equal(t, 1, maxConcurrent)
}
