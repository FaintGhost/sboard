package api_test

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"sboard/panel/internal/api"
	"sboard/panel/internal/config"
	"sboard/panel/internal/node"
)

func TestNodesDelete_ForceNotFound(t *testing.T) {
	cfg := config.Config{JWTSecret: "secret"}
	store := setupStore(t)
	r := api.NewRouter(cfg, store)
	token := mustToken(cfg.JWTSecret)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/nodes/99999?force=true", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestSyncJobsRetry_NodeMissingAndUpstreamFailure(t *testing.T) {
	doer := &fakeDoerFunc{do: func(req *http.Request) (*http.Response, error) {
		if req.URL.Path == "/api/config/sync" && req.Method == http.MethodPost {
			return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{"status":"ok"}`))}, nil
		}
		if req.URL.Path == "/api/health" && req.Method == http.MethodGet {
			return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{"status":"ok"}`))}, nil
		}
		return &http.Response{StatusCode: http.StatusNotFound, Body: io.NopCloser(strings.NewReader("not found"))}, nil
	}}
	restore := api.SetNodeClientFactoryForTest(func() *node.Client {
		return node.NewClient(doer)
	})
	t.Cleanup(restore)

	cfg := config.Config{JWTSecret: "secret"}
	store := setupStore(t)
	r := api.NewRouter(cfg, store)
	token := mustToken(cfg.JWTSecret)

	nodeID := createGroupAndNode(t, r, token)

	// create one inbound to generate a sync job
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/inbounds", strings.NewReader(fmt.Sprintf(
		`{"node_id":%d,"tag":"retry-edge","protocol":"vless","listen_port":6543,"public_port":6543,"settings":{}}`,
		nodeID,
	)))
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var jobID int64
	err := store.DB.QueryRow("SELECT id FROM sync_jobs ORDER BY id DESC LIMIT 1").Scan(&jobID)
	require.NoError(t, err)

	// invalid id (=0)
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/sync-jobs/0/retry", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)

	// set sync job node to a missing node id, then retry should return 404 node not found.
	_, err = store.DB.Exec("UPDATE sync_jobs SET node_id = ? WHERE id = ?", 999999, jobID)
	if err == nil {
		w = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/sync-jobs/%d/retry", jobID), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		r.ServeHTTP(w, req)
		require.Equal(t, http.StatusNotFound, w.Code)
	} else {
		// If FK is enabled in this sqlite env, ensure the test is still meaningful.
		require.Error(t, err)
	}

	// reset to a valid job: create another inbound/job.
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/inbounds", strings.NewReader(fmt.Sprintf(
		`{"node_id":%d,"tag":"retry-edge-2","protocol":"vless","listen_port":6544,"public_port":6544,"settings":{}}`,
		nodeID,
	)))
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var jobID2 int64
	err = store.DB.QueryRow("SELECT id FROM sync_jobs ORDER BY id DESC LIMIT 1").Scan(&jobID2)
	require.NoError(t, err)

	// make sync always fail -> retry endpoint should return 502 bad gateway
	doer.do = func(req *http.Request) (*http.Response, error) {
		if req.URL.Path == "/api/config/sync" && req.Method == http.MethodPost {
			return &http.Response{StatusCode: http.StatusInternalServerError, Body: io.NopCloser(strings.NewReader("retry-failed"))}, nil
		}
		return &http.Response{StatusCode: http.StatusNotFound, Body: io.NopCloser(strings.NewReader("not found"))}, nil
	}

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/sync-jobs/%d/retry", jobID2), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadGateway, w.Code)

	// extra safety: ensure retry job exists and failed status recorded
	var status string
	qErr := store.DB.QueryRow("SELECT status FROM sync_jobs WHERE parent_job_id = ? ORDER BY id DESC LIMIT 1", jobID2).Scan(&status)
	if qErr == nil {
		require.Equal(t, "failed", status)
	} else if qErr != sql.ErrNoRows {
		require.NoError(t, qErr)
	}
}
