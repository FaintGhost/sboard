package api_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"sboard/panel/internal/api"
	"sboard/panel/internal/config"
)

func TestTrafficEndpoints_StoreNotReady(t *testing.T) {
	cfg := config.Config{JWTSecret: "secret"}
	r := api.NewRouter(cfg, nil)
	token := mustToken(cfg.JWTSecret)

	paths := []string{
		"/api/nodes/1/traffic",
		"/api/traffic/nodes/summary",
		"/api/traffic/total/summary",
		"/api/traffic/timeseries",
	}

	for _, path := range paths {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusInternalServerError, w.Code, "path=%s", path)
		require.Contains(t, w.Body.String(), "store not ready", "path=%s", path)
	}
}

func TestNodeTrafficList_PaginationValidation(t *testing.T) {
	cfg := config.Config{JWTSecret: "secret"}
	store := setupStore(t)
	r := api.NewRouter(cfg, store)
	token := mustToken(cfg.JWTSecret)

	node := createNodeForAPITests(t, store)

	badPaths := []string{
		fmt.Sprintf("/api/nodes/%d/traffic?limit=501&offset=0", node.ID),
		fmt.Sprintf("/api/nodes/%d/traffic?limit=-1&offset=0", node.ID),
		fmt.Sprintf("/api/nodes/%d/traffic?limit=abc&offset=0", node.ID),
		fmt.Sprintf("/api/nodes/%d/traffic?limit=10&offset=-1", node.ID),
	}

	for _, path := range badPaths {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusBadRequest, w.Code, "path=%s", path)
		require.Contains(t, w.Body.String(), "invalid pagination", "path=%s", path)
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/nodes/%d/traffic?limit=500&offset=0", node.ID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), `"data":[]`)
}
