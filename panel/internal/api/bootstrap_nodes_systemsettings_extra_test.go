package api_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"sboard/panel/internal/api"
	"sboard/panel/internal/config"
)

func TestAdminBootstrapPost_ValidationAndUnauthorizedCases(t *testing.T) {
	store := setupStore(t)

	// setup token missing in config => always unauthorized
	r := api.NewRouter(config.Config{JWTSecret: "secret"}, store)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/admin/bootstrap", bytes.NewReader([]byte(`{"username":"admin","password":"pass12345","confirm_password":"pass12345","setup_token":"x"}`)))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusUnauthorized, w.Code)

	// setup token configured
	r = api.NewRouter(config.Config{JWTSecret: "secret", SetupToken: "setup-123"}, store)

	// invalid body
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/admin/bootstrap", bytes.NewReader([]byte(`{`)))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)

	cases := []struct {
		name string
		body string
	}{
		{name: "missing username", body: `{"password":"pass12345","confirm_password":"pass12345","setup_token":"setup-123"}`},
		{name: "missing password", body: `{"username":"admin","confirm_password":"pass12345","setup_token":"setup-123"}`},
		{name: "password mismatch", body: `{"username":"admin","password":"pass12345","confirm_password":"different","setup_token":"setup-123"}`},
		{name: "password too short", body: `{"username":"admin","password":"short","confirm_password":"short","setup_token":"setup-123"}`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/api/admin/bootstrap", bytes.NewReader([]byte(tc.body)))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)
			require.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

func TestNodesGet_InvalidIDNotFoundAndSuccess(t *testing.T) {
	cfg := config.Config{JWTSecret: "secret"}
	store := setupStore(t)
	r := api.NewRouter(cfg, store)
	token := mustToken(cfg.JWTSecret)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/nodes/not-a-number", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/nodes/99999", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusNotFound, w.Code)

	nodeID := createGroupAndNode(t, r, token)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/nodes/%d", nodeID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Equal(t, nodeID, resp.Data.ID)
}

func TestSystemSettings_GetPut_StoreClosedReturnsInternalError(t *testing.T) {
	cfg := config.Config{JWTSecret: "secret"}
	store := setupStore(t)
	r := api.NewRouter(cfg, store)
	token := mustToken(cfg.JWTSecret)

	require.NoError(t, store.DB.Close())

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/system/settings", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusInternalServerError, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, "/api/system/settings", bytes.NewReader([]byte(`{"subscription_base_url":"","timezone":"UTC"}`)))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusInternalServerError, w.Code)
}
