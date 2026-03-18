package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
	"sboard/panel/internal/api"
	"sboard/panel/internal/config"
	"sboard/panel/internal/node"
	nodev1 "sboard/panel/internal/rpc/gen/sboard/node/v1"
)

type syncJobsFlakyDoer struct {
	got      int32
	failLeft int32
}

func (d *syncJobsFlakyDoer) Do(req *http.Request) (*http.Response, error) {
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

func TestSyncJobsAPI_ListGetRetry(t *testing.T) {
	doer := &syncJobsFlakyDoer{}
	restore := api.SetNodeClientFactoryForTest(func() *node.Client {
		return node.NewClient(doer)
	})
	t.Cleanup(restore)

	cfg := config.Config{JWTSecret: "secret"}
	store := setupStore(t)
	r := api.NewRouter(cfg, store)
	token := mustToken(cfg.JWTSecret)

	// group
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/groups", strings.NewReader(`{"name":"g1","description":""}`))
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)
	var g groupResp
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &g))

	// node
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/nodes", bytes.NewReader([]byte(fmt.Sprintf(`{"name":"n1","api_address":"127.0.0.1","api_port":3000,"secret_key":"secret","public_address":"a.example.com","group_id":%d}`, g.Data.ID))))
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)
	var n nodeResp
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &n))

	// inbound (creates one auto sync job)
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/inbounds", strings.NewReader(fmt.Sprintf(`{"node_id":%d,"tag":"vless-in","protocol":"vless","listen_port":443,"public_port":443,"settings":{"users":[{"uuid":"a","flow":"xtls-rprx-vision"}]}}`, n.Data.ID)))
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	// list jobs
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/sync-jobs?limit=20&offset=0", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	var listResp struct {
		Data []struct {
			ID     int64  `json:"id"`
			NodeID int64  `json:"node_id"`
			Status string `json:"status"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listResp))
	require.NotEmpty(t, listResp.Data)

	jobID := listResp.Data[0].ID

	// get job detail
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/sync-jobs/%d", jobID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	// retry from existing job
	atomic.StoreInt32(&doer.failLeft, 2)
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/sync-jobs/%d/retry", jobID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var retryResp struct {
		Data struct {
			ID           int64  `json:"id"`
			ParentJobID  *int64 `json:"parent_job_id"`
			Status       string `json:"status"`
			AttemptCount int    `json:"attempt_count"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &retryResp))
	require.NotNil(t, retryResp.Data.ParentJobID)
	require.Equal(t, jobID, *retryResp.Data.ParentJobID)
	require.Equal(t, "success", retryResp.Data.Status)
	require.Equal(t, 3, retryResp.Data.AttemptCount)
}
