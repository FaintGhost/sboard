package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"sboard/panel/internal/api"
	"sboard/panel/internal/config"
)

func TestListEndpoints_PaginationLimitTooLarge(t *testing.T) {
	cfg := config.Config{JWTSecret: "secret"}
	store := setupStore(t)
	r := api.NewRouter(cfg, store)
	token := mustToken(cfg.JWTSecret)

	paths := []string{
		"/api/users?limit=501&offset=0",
		"/api/groups?limit=501&offset=0",
		"/api/nodes?limit=501&offset=0",
		"/api/inbounds?limit=501&offset=0",
		"/api/sync-jobs?limit=501&offset=0",
	}

	for _, path := range paths {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusBadRequest, w.Code, "path=%s", path)
		require.Contains(t, w.Body.String(), "invalid pagination", "path=%s", path)
	}
}

func TestListEndpoints_PaginationLimitBoundaryAccepted(t *testing.T) {
	cfg := config.Config{JWTSecret: "secret"}
	store := setupStore(t)
	r := api.NewRouter(cfg, store)
	token := mustToken(cfg.JWTSecret)

	paths := []string{
		"/api/users?limit=500&offset=0",
		"/api/groups?limit=500&offset=0",
		"/api/nodes?limit=500&offset=0",
		"/api/inbounds?limit=500&offset=0",
		"/api/sync-jobs?limit=500&offset=0",
	}

	for _, path := range paths {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		r.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code, "path=%s", path)
	}
}
