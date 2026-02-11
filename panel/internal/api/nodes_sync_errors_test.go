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

func TestNodeSync_InvalidIDNotFoundAndMissingGroup(t *testing.T) {
	cfg := config.Config{JWTSecret: "secret"}
	store := setupStore(t)
	r := api.NewRouter(cfg, store)
	token := mustToken(cfg.JWTSecret)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/nodes/not-a-number/sync", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/nodes/99999/sync", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusNotFound, w.Code)

	// Create node without group_id.
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/nodes", strings.NewReader(`{"name":"n1","api_address":"127.0.0.1","api_port":3000,"secret_key":"secret","public_address":"a.example.com"}`))
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var n nodeResp
	mustUnmarshal(t, w.Body.Bytes(), &n)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/nodes/%d/sync", n.Data.ID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestNodeSync_UpstreamFailureReturnsBadGateway(t *testing.T) {
	restore := api.SetNodeClientFactoryForTest(func() *node.Client {
		return node.NewClient(&fakeDoerFunc{do: func(req *http.Request) (*http.Response, error) {
			if req.URL.Path == "/api/config/sync" && req.Method == http.MethodPost {
				return &http.Response{StatusCode: http.StatusInternalServerError, Body: io.NopCloser(strings.NewReader("boom"))}, nil
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

	nodeID := createGroupAndNode(t, r, token)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/nodes/%d/sync", nodeID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadGateway, w.Code)
}

func mustUnmarshal(t *testing.T, b []byte, dst any) {
	t.Helper()
	require.NoError(t, json.Unmarshal(b, dst))
}
