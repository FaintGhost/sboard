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
	"sboard/panel/internal/node"
)

func TestNodesDelete_NonForce_InvalidNotFoundConflict(t *testing.T) {
	cfg := config.Config{JWTSecret: "secret"}
	store := setupStore(t)
	r := api.NewRouter(cfg, store)
	token := mustToken(cfg.JWTSecret)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/nodes/not-a-number", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, "/api/nodes/99999", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusNotFound, w.Code)

	nodeID := createGroupAndNode(t, r, token)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/inbounds", strings.NewReader(fmt.Sprintf(
		`{"node_id":%d,"tag":"conflict-delete","protocol":"vless","listen_port":6443,"public_port":6443,"settings":{}}`,
		nodeID,
	)))
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/nodes/%d", nodeID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusConflict, w.Code)
}

func TestNodesDelete_ForceDrainFailureAndNoInbounds(t *testing.T) {
	restore := api.SetNodeClientFactoryForTest(func() *node.Client {
		return node.NewClient(&fakeDoerFunc{do: func(req *http.Request) (*http.Response, error) {
			if req.URL.Path == "/api/config/sync" && req.Method == http.MethodPost {
				return &http.Response{StatusCode: http.StatusInternalServerError, Body: io.NopCloser(strings.NewReader("drain-failed"))}, nil
			}
			return &http.Response{StatusCode: http.StatusNotFound, Body: io.NopCloser(strings.NewReader("not found"))}, nil
		}})
	})
	t.Cleanup(restore)

	cfg := config.Config{JWTSecret: "secret"}
	store := setupStore(t)
	r := api.NewRouter(cfg, store)
	token := mustToken(cfg.JWTSecret)

	// Node with inbound: force drain should fail as upstream returns 500.
	nodeWithInbound := createGroupAndNode(t, r, token)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/inbounds", strings.NewReader(fmt.Sprintf(
		`{"node_id":%d,"tag":"force-fail","protocol":"vless","listen_port":7443,"public_port":7443,"settings":{}}`,
		nodeWithInbound,
	)))
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/nodes/%d?force=true", nodeWithInbound), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadGateway, w.Code)

	// Node without inbound: force delete should bypass drain and succeed.
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/groups", strings.NewReader(`{"name":"g2-force","description":""}`))
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)
	var g2 groupResp
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &g2))

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/nodes", strings.NewReader(fmt.Sprintf(`{"name":"n2-force","api_address":"127.0.0.1","api_port":3001,"secret_key":"secret","public_address":"a2.example.com","group_id":%d}`, g2.Data.ID)))
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)
	var n2 nodeResp
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &n2))
	nodeWithoutInbound := n2.Data.ID

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/nodes/%d?force=yes", nodeWithoutInbound), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/nodes/%d", nodeWithoutInbound), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestUsersUpdate_ErrorPathsAndConflict(t *testing.T) {
	cfg := config.Config{JWTSecret: "secret"}
	store := setupStore(t)
	r := api.NewRouter(cfg, store)
	token := mustToken(cfg.JWTSecret)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/users/not-a-number", strings.NewReader(`{"username":"a"}`))
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, "/api/users/99999", strings.NewReader(`{"username":"a"}`))
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusNotFound, w.Code)

	// create target user
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(`{"username":"alice-update"}`))
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)
	var alice userResp
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &alice))

	badCases := []string{
		`{`,
		`{"username":"   "}`,
		`{"status":"bad-status"}`,
		`{"expire_at":"not-time"}`,
		`{"traffic_limit":-1}`,
		`{"traffic_reset_day":32}`,
	}
	for _, body := range badCases {
		w = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/users/%d", alice.Data.ID), strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		r.ServeHTTP(w, req)
		require.Equal(t, http.StatusBadRequest, w.Code, body)
	}

	// conflict username
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(`{"username":"bob-update"}`))
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)
	var bob userResp
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &bob))

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/users/%d", bob.Data.ID), strings.NewReader(`{"username":"alice-update"}`))
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusConflict, w.Code)
}
