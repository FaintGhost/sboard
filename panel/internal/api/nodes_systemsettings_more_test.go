package api_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"sboard/panel/internal/api"
	"sboard/panel/internal/config"
)

func TestNodesDelete_NonForceSuccess(t *testing.T) {
	cfg := config.Config{JWTSecret: "secret"}
	store := setupStore(t)
	r := api.NewRouter(cfg, store)
	token := mustToken(cfg.JWTSecret)

	nodeID := createGroupAndNode(t, r, token)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/nodes/%d", nodeID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/nodes/%d", nodeID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestNodesList_InvalidPaginationAndStoreError(t *testing.T) {
	cfg := config.Config{JWTSecret: "secret"}
	store := setupStore(t)
	r := api.NewRouter(cfg, store)
	token := mustToken(cfg.JWTSecret)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/nodes?limit=-1&offset=0", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)

	require.NoError(t, store.DB.Close())

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/nodes?limit=10&offset=0", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestSystemSettingsPut_InvalidBodyAndUpsertFailBranch(t *testing.T) {
	cfg := config.Config{JWTSecret: "secret"}
	store := setupStore(t)
	r := api.NewRouter(cfg, store)
	token := mustToken(cfg.JWTSecret)

	// invalid body
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/system/settings", bytes.NewReader([]byte(`{`)))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)

	// close DB to force upsert failure on non-empty base url branch
	require.NoError(t, store.DB.Close())

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, "/api/system/settings", bytes.NewReader([]byte(`{"subscription_base_url":"https://203.0.113.10:8443","timezone":"UTC"}`)))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestSystemSettingsGet_InvalidSavedTimezoneFallsBackToCurrent(t *testing.T) {
	cfg := config.Config{JWTSecret: "secret"}
	store := setupStore(t)
	r := api.NewRouter(cfg, store)
	token := mustToken(cfg.JWTSecret)

	// First set a valid timezone via API so current runtime timezone is deterministic.
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/system/settings", strings.NewReader(`{"subscription_base_url":"","timezone":"Asia/Shanghai"}`))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	// Corrupt persisted timezone directly in db; GET should fallback to current runtime timezone.
	_, err := store.DB.Exec("UPDATE system_settings SET value = ? WHERE key = ?", "Mars/Base", "timezone")
	require.NoError(t, err)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/system/settings", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		Data struct {
			Timezone string `json:"timezone"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Equal(t, "Asia/Shanghai", resp.Data.Timezone)
}
