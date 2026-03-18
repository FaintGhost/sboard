package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/require"
	"sboard/panel/internal/api"
	"sboard/panel/internal/config"
	"sboard/panel/internal/node"
	nodev1 "sboard/panel/internal/rpc/gen/sboard/node/v1"
)

type fakeDoer struct {
	got int32
}

func (d *fakeDoer) Do(req *http.Request) (*http.Response, error) {
	return serveNodeRPCRequest(req, nodeRPCServiceStub{
		syncConfigFunc: func(_ context.Context, reqBody *nodev1.SyncConfigRequest) (*nodev1.SyncConfigResponse, error) {
			if req.Header.Get("Authorization") != "Bearer secret" {
				return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("unauthorized"))
			}

			var body struct {
				Inbounds []map[string]any `json:"inbounds"`
			}
			if err := json.Unmarshal(reqBody.PayloadJson, &body); err != nil {
				return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("bad json"))
			}
			if len(body.Inbounds) != 1 {
				return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("bad inbounds"))
			}
			inb := body.Inbounds[0]
			if inb["type"] != "vless" || inb["tag"] != "vless-in" {
				return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("bad inbound"))
			}
			if _, ok := inb["users"]; !ok {
				return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("missing users"))
			}
			atomic.AddInt32(&d.got, 1)
			return &nodev1.SyncConfigResponse{Status: "ok"}, nil
		},
	}, nil)
}

type flakySyncDoer struct {
	got      int32
	failLeft int32
}

func (d *flakySyncDoer) Do(req *http.Request) (*http.Response, error) {
	return serveNodeRPCRequest(req, nodeRPCServiceStub{
		syncConfigFunc: func(context.Context, *nodev1.SyncConfigRequest) (*nodev1.SyncConfigResponse, error) {
			atomic.AddInt32(&d.got, 1)
			if atomic.LoadInt32(&d.failLeft) > 0 {
				atomic.AddInt32(&d.failLeft, -1)
				return nil, io.ErrUnexpectedEOF
			}
			return &nodev1.SyncConfigResponse{Status: "ok"}, nil
		},
	}, nil)
}

type blockingSyncDoer struct {
	mu            sync.Mutex
	inFlight      int
	maxInFlight   int
	totalRequests int
	delay         time.Duration
}

func newBlockingSyncDoer(delay time.Duration) *blockingSyncDoer {
	return &blockingSyncDoer{
		delay: delay,
	}
}

func (d *blockingSyncDoer) Do(req *http.Request) (*http.Response, error) {
	return serveNodeRPCRequest(req, nodeRPCServiceStub{
		syncConfigFunc: func(context.Context, *nodev1.SyncConfigRequest) (*nodev1.SyncConfigResponse, error) {
			d.mu.Lock()
			d.totalRequests++
			d.inFlight++
			if d.inFlight > d.maxInFlight {
				d.maxInFlight = d.inFlight
			}
			d.mu.Unlock()

			time.Sleep(d.delay)

			d.mu.Lock()
			d.inFlight--
			d.mu.Unlock()

			return &nodev1.SyncConfigResponse{Status: "ok"}, nil
		},
	}, nil)
}

func (d *blockingSyncDoer) maxConcurrent() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.maxInFlight
}

func TestNodeSync_PushesConfigToNode(t *testing.T) {
	doer := &fakeDoer{}
	restore := api.SetNodeClientFactoryForTest(func() *node.Client {
		return node.NewClient(doer)
	})
	t.Cleanup(restore)

	cfg := config.Config{JWTSecret: "secret"}
	store := setupStore(t)
	r := api.NewRouter(cfg, store)
	token := mustToken(cfg.JWTSecret)

	// Create group.
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/groups", strings.NewReader(`{"name":"g1","description":""}`))
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)
	var g groupResp
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &g))

	// Create node.
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/nodes", bytes.NewReader([]byte(fmt.Sprintf(`{"name":"n1","api_address":"127.0.0.1","api_port":3000,"secret_key":"secret","public_address":"a.example.com","group_id":%d}`, g.Data.ID))))
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)
	var n nodeResp
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &n))

	// Create inbound.
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/inbounds", strings.NewReader(fmt.Sprintf(`{"node_id":%d,"tag":"vless-in","protocol":"vless","listen_port":443,"public_port":443,"settings":{"users":[{"uuid":"a","flow":"xtls-rprx-vision"}]}}`, n.Data.ID)))
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	// Create user and bind to group.
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(`{"username":"alice"}`))
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)
	var u userResp
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &u))

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/users/%d/groups", u.Data.ID), strings.NewReader(fmt.Sprintf(`{"group_ids":[%d]}`, g.Data.ID)))
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	// Sync.
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/nodes/%d/sync", n.Data.ID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	// One sync is triggered on inbound create, one on user-group bind, and one here.
	require.Equal(t, int32(3), atomic.LoadInt32(&doer.got))
}

func TestNodeSync_RetriesAndRecordsSyncJob(t *testing.T) {
	doer := &flakySyncDoer{}
	restore := api.SetNodeClientFactoryForTest(func() *node.Client {
		return node.NewClient(doer)
	})
	t.Cleanup(restore)

	cfg := config.Config{JWTSecret: "secret"}
	store := setupStore(t)
	r := api.NewRouter(cfg, store)
	token := mustToken(cfg.JWTSecret)

	// Create group.
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/groups", strings.NewReader(`{"name":"g1","description":""}`))
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)
	var g groupResp
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &g))

	// Create node.
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/nodes", bytes.NewReader([]byte(fmt.Sprintf(`{"name":"n1","api_address":"127.0.0.1","api_port":3000,"secret_key":"secret","public_address":"a.example.com","group_id":%d}`, g.Data.ID))))
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)
	var n nodeResp
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &n))

	// Create inbound.
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/inbounds", strings.NewReader(fmt.Sprintf(`{"node_id":%d,"tag":"vless-in","protocol":"vless","listen_port":443,"public_port":443,"settings":{"users":[{"uuid":"a","flow":"xtls-rprx-vision"}]}}`, n.Data.ID)))
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	atomic.StoreInt32(&doer.got, 0)
	atomic.StoreInt32(&doer.failLeft, 2)

	// Call sync endpoint.
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/nodes/%d/sync", n.Data.ID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, int32(3), atomic.LoadInt32(&doer.got))

	var jobCount int
	err := store.DB.QueryRow("SELECT COUNT(1) FROM sync_jobs WHERE node_id = ?", n.Data.ID).Scan(&jobCount)
	require.NoError(t, err)
	require.GreaterOrEqual(t, jobCount, 1)

	var lastJobID int64
	var status string
	var attemptCount int
	err = store.DB.QueryRow(
		"SELECT id, status, attempt_count FROM sync_jobs WHERE node_id = ? ORDER BY id DESC LIMIT 1",
		n.Data.ID,
	).Scan(&lastJobID, &status, &attemptCount)
	require.NoError(t, err)
	require.Equal(t, "success", status)
	require.Equal(t, 3, attemptCount)

	var attempts int
	err = store.DB.QueryRow("SELECT COUNT(1) FROM sync_attempts WHERE job_id = ?", lastJobID).Scan(&attempts)
	require.NoError(t, err)
	require.Equal(t, 3, attempts)
}

func TestNodeSync_SerializesPerNode(t *testing.T) {
	doer := newBlockingSyncDoer(120 * time.Millisecond)
	restore := api.SetNodeClientFactoryForTest(func() *node.Client {
		return node.NewClient(doer)
	})
	t.Cleanup(restore)

	cfg := config.Config{JWTSecret: "secret"}
	store := setupStore(t)
	r := api.NewRouter(cfg, store)
	token := mustToken(cfg.JWTSecret)

	// Create group.
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/groups", strings.NewReader(`{"name":"g1","description":""}`))
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)
	var g groupResp
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &g))

	// Create node.
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/nodes", bytes.NewReader([]byte(fmt.Sprintf(`{"name":"n1","api_address":"127.0.0.1","api_port":3000,"secret_key":"secret","public_address":"a.example.com","group_id":%d}`, g.Data.ID))))
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)
	var n nodeResp
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &n))

	// Create inbound.
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/inbounds", strings.NewReader(fmt.Sprintf(`{"node_id":%d,"tag":"vless-in","protocol":"vless","listen_port":443,"public_port":443,"settings":{"users":[{"uuid":"a","flow":"xtls-rprx-vision"}]}}`, n.Data.ID)))
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	start := make(chan struct{})
	results := make(chan int, 2)
	runSync := func() {
		<-start
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/nodes/%d/sync", n.Data.ID), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		r.ServeHTTP(w, req)
		results <- w.Code
	}

	go runSync()
	go runSync()
	close(start)

	code1 := <-results
	code2 := <-results
	require.Equal(t, http.StatusOK, code1)
	require.Equal(t, http.StatusOK, code2)
	require.Equal(t, 1, doer.maxConcurrent())
}
