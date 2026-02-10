package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"sboard/panel/internal/api"
	"sboard/panel/internal/config"
	"sboard/panel/internal/db"
	"sboard/panel/internal/password"
)

func TestAdminBootstrapGet_DBErrorWhenStoreMissing(t *testing.T) {
	cfg := config.Config{JWTSecret: "secret"}
	r := api.NewRouter(cfg, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/admin/bootstrap", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)
	require.Contains(t, w.Body.String(), "db error")
}

func TestAdminBootstrapGet_NeedsSetupFalseWhenAdminExists(t *testing.T) {
	store := setupStore(t)
	hash, err := password.Hash("pass12345")
	require.NoError(t, err)
	created, err := db.AdminCreateIfNone(store, "admin", hash)
	require.NoError(t, err)
	require.True(t, created)

	cfg := config.Config{JWTSecret: "secret"}
	r := api.NewRouter(cfg, store)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/admin/bootstrap", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	data, ok := body["data"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, false, data["needs_setup"])
}

func TestAdminBootstrapPost_UsesHeaderTokenAndRejectsSecondBootstrap(t *testing.T) {
	store := setupStore(t)
	cfg := config.Config{JWTSecret: "secret", SetupToken: "setup-header-123"}
	r := api.NewRouter(cfg, store)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/admin/bootstrap", strings.NewReader(`{"username":"admin","password":"pass12345","confirm_password":"pass12345"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Setup-Token", "setup-header-123")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/admin/bootstrap", strings.NewReader(`{"username":"admin2","password":"pass12345","confirm_password":"pass12345"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Setup-Token", "setup-header-123")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusConflict, w.Code)
	require.Contains(t, w.Body.String(), "already initialized")
}
