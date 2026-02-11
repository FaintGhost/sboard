package api_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"sboard/panel/internal/api"
	"sboard/panel/internal/config"
	"sboard/panel/internal/db"
	"sboard/panel/internal/node"
)

func TestGroupsDelete_InvalidIDNotFoundAndConflict(t *testing.T) {
	cfg := config.Config{JWTSecret: "secret"}
	store := setupStore(t)
	r := api.NewRouter(cfg, store)
	token := mustToken(cfg.JWTSecret)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/groups/not-a-number", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, "/api/groups/99999", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusNotFound, w.Code)

	// Create group + node that references this group => delete should conflict.
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/groups", strings.NewReader(`{"name":"g-conflict","description":""}`))
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var g groupResp
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &g))

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/nodes", strings.NewReader(fmt.Sprintf(`{"name":"n1","api_address":"127.0.0.1","api_port":3000,"secret_key":"secret","public_address":"a.example.com","group_id":%d}`, g.Data.ID)))
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/groups/%d", g.Data.ID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusConflict, w.Code)
}

func TestUserGroups_GetPutErrorPaths(t *testing.T) {
	cfg := config.Config{JWTSecret: "secret"}
	store := setupStore(t)
	r := api.NewRouter(cfg, store)
	token := mustToken(cfg.JWTSecret)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/users/not-a-number/groups", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/users/99999/groups", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusNotFound, w.Code)

	// create user
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(`{"username":"alice-groups-edge"}`))
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)
	var user userResp
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &user))

	// invalid body
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/users/%d/groups", user.Data.ID), strings.NewReader(`{`))
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)

	// invalid group id (<=0) => bad request
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/users/%d/groups", user.Data.ID), strings.NewReader(`{"group_ids":[0]}`))
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)

	// not-found group => 404
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/users/%d/groups", user.Data.ID), strings.NewReader(`{"group_ids":[99999]}`))
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestSyncJobs_ListGetRetryErrorPathsAndFilters(t *testing.T) {
	restore := api.SetNodeClientFactoryForTest(func() *node.Client {
		return node.NewClient(&fakeDoerFunc{do: func(req *http.Request) (*http.Response, error) {
			if req.URL.Path == "/api/config/sync" && req.Method == http.MethodPost {
				return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{"status":"ok"}`))}, nil
			}
			if req.URL.Path == "/api/health" && req.Method == http.MethodGet {
				return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{"status":"ok"}`))}, nil
			}
			return &http.Response{StatusCode: http.StatusNotFound, Body: io.NopCloser(strings.NewReader("not found"))}, nil
		}})
	})
	t.Cleanup(restore)

	cfg := config.Config{JWTSecret: "secret"}
	store := setupStore(t)
	r := api.NewRouter(cfg, store)
	token := mustToken(cfg.JWTSecret)

	// invalid id paths
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/sync-jobs/not-a-number", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/sync-jobs/not-a-number/retry", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/sync-jobs/99999/retry", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusNotFound, w.Code)

	// Prepare one real sync job by creating group+node+inbound (auto sync)
	nodeID := createGroupAndNode(t, r, token)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/inbounds", strings.NewReader(fmt.Sprintf(
		`{"node_id":%d,"tag":"sync-jobs-edge","protocol":"vless","listen_port":7443,"public_port":7443,"settings":{}}`,
		nodeID,
	)))
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	// list validation paths
	badListCases := []string{
		"/api/sync-jobs?node_id=abc",
		"/api/sync-jobs?status=not-valid",
		"/api/sync-jobs?from=not-time",
		"/api/sync-jobs?to=not-time",
	}
	for _, path := range badListCases {
		w = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodGet, path, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		r.ServeHTTP(w, req)
		require.Equal(t, http.StatusBadRequest, w.Code, path)
	}

	// valid list filters
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/sync-jobs?limit=10&offset=0&node_id=%d&status=%s&trigger_source=auto_inbound_change", nodeID, db.SyncJobStatusSuccess), nil)
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

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/sync-jobs/99999", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusNotFound, w.Code)
}
