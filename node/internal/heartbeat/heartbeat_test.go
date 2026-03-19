package heartbeat_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sboard/node/internal/heartbeat"
)

// requestRecord captures the fields of a single heartbeat request.
type requestRecord struct {
	UUID      string `json:"uuid"`
	SecretKey string `json:"secretKey"`
	Version   string `json:"version"`
	APIAddr   string `json:"apiAddr"`
}

func TestHeartbeat_SendsOnTick(t *testing.T) {
	received := make(chan requestRecord, 10)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "/rpc/sboard.panel.v1.NodeRegistrationService/Heartbeat", r.URL.Path)

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		var rec requestRecord
		require.NoError(t, json.Unmarshal(body, &rec))
		received <- rec

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status": "NODE_HEARTBEAT_STATUS_RECOGNIZED",
		})
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := heartbeat.Config{
		PanelURL:  srv.URL,
		NodeUUID:  "uuid-test-1",
		SecretKey: "secret-test-1",
		APIAddr:   ":3000",
		Version:   "1.0.0",
		Interval:  50 * time.Millisecond,
	}

	done := make(chan struct{})
	go func() {
		heartbeat.Run(ctx, cfg, srv.Client())
		close(done)
	}()

	// Wait for at least 2 heartbeats (immediate + 1 tick).
	var records []requestRecord
	timeout := time.After(2 * time.Second)
	for len(records) < 2 {
		select {
		case rec := <-received:
			records = append(records, rec)
		case <-timeout:
			t.Fatalf("timed out waiting for heartbeats; got %d", len(records))
		}
	}

	cancel()
	<-done

	// Verify the first heartbeat request fields.
	rec := records[0]
	assert.Equal(t, "uuid-test-1", rec.UUID)
	assert.Equal(t, "secret-test-1", rec.SecretKey)
	assert.Equal(t, ":3000", rec.APIAddr)
	assert.Equal(t, "1.0.0", rec.Version)
}

func TestHeartbeat_UsesExistingRPCPrefix(t *testing.T) {
	received := make(chan struct{}, 1)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/rpc/sboard.panel.v1.NodeRegistrationService/Heartbeat", r.URL.Path)
		received <- struct{}{}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status": "NODE_HEARTBEAT_STATUS_RECOGNIZED",
		})
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan struct{})
	go func() {
		heartbeat.Run(ctx, heartbeat.Config{
			PanelURL:  srv.URL + "/rpc",
			NodeUUID:  "uuid-test-rpc",
			SecretKey: "secret-test-rpc",
			Interval:  50 * time.Millisecond,
		}, srv.Client())
		close(done)
	}()

	select {
	case <-received:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for heartbeat request")
	}

	cancel()
	<-done
}

func TestHeartbeat_DisabledWhenNoPanelURL(t *testing.T) {
	cfg := heartbeat.Config{
		PanelURL: "", // disabled
		Interval: 50 * time.Millisecond,
	}

	done := make(chan struct{})
	go func() {
		heartbeat.Run(context.Background(), cfg, nil)
		close(done)
	}()

	select {
	case <-done:
		// Run returned immediately — expected.
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Run did not return immediately when PanelURL is empty")
	}
}

func TestHeartbeat_GracefulShutdown(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status": "NODE_HEARTBEAT_STATUS_RECOGNIZED",
		})
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())

	cfg := heartbeat.Config{
		PanelURL:  srv.URL,
		NodeUUID:  "uuid-shutdown",
		SecretKey: "key",
		Interval:  50 * time.Millisecond,
	}

	done := make(chan struct{})
	go func() {
		heartbeat.Run(ctx, cfg, srv.Client())
		close(done)
	}()

	// Let at least one heartbeat fire, then cancel.
	time.Sleep(80 * time.Millisecond)
	cancel()

	select {
	case <-done:
		// Goroutine exited promptly — expected.
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not exit after context cancellation")
	}
}

func TestHeartbeat_ContinuesOnError(t *testing.T) {
	var count atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := count.Add(1)
		if n <= 2 {
			// First two requests return 500.
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// Subsequent requests succeed.
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status": "NODE_HEARTBEAT_STATUS_RECOGNIZED",
		})
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := heartbeat.Config{
		PanelURL:  srv.URL,
		NodeUUID:  "uuid-err",
		SecretKey: "key",
		Interval:  50 * time.Millisecond,
	}

	done := make(chan struct{})
	go func() {
		heartbeat.Run(ctx, cfg, srv.Client())
		close(done)
	}()

	// Wait until at least 3 requests have been made (2 errors + 1 success).
	deadline := time.After(3 * time.Second)
	for {
		if count.Load() >= 3 {
			break
		}
		select {
		case <-deadline:
			t.Fatalf("timed out; only %d requests sent", count.Load())
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}

	cancel()
	<-done

	assert.GreaterOrEqual(t, count.Load(), int32(3), "heartbeat should have continued past errors")
}
