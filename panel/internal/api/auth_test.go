package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"sboard/panel/internal/api"
	"sboard/panel/internal/config"
)

func TestAdminLogin_NeedsSetupWhenNoAdmin(t *testing.T) {
	store := setupStore(t)
	cfg := config.Config{JWTSecret: "secret"}
	r := api.NewRouter(cfg, store)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/admin/login", strings.NewReader(`{"username":"admin","password":"pass"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusPreconditionRequired, w.Code)
}

func TestAdminBootstrapThenLogin(t *testing.T) {
	t.Setenv("PANEL_SETUP_TOKEN", "setup-123")
	store := setupStore(t)

	cfg := config.Load()
	cfg.JWTSecret = "secret"
	r := api.NewRouter(cfg, store)

	// wrong token
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/admin/bootstrap", strings.NewReader(`{"username":"admin","password":"pass","confirm_password":"pass","setup_token":"wrong"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusUnauthorized, w.Code)

	// correct token
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/admin/bootstrap", strings.NewReader(`{"username":"admin","password":"pass12345","confirm_password":"pass12345","setup_token":"setup-123"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	// login ok
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/admin/login", strings.NewReader(`{"username":"admin","password":"pass12345"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	// basic shape
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data, _ := resp["data"].(map[string]any)
	require.NotEmpty(t, data["token"])
}

func TestAdminProfile_GetAndUpdateCredentials(t *testing.T) {
	t.Setenv("PANEL_SETUP_TOKEN", "setup-123")
	store := setupStore(t)

	cfg := config.Load()
	cfg.JWTSecret = "secret"
	r := api.NewRouter(cfg, store)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/admin/bootstrap", strings.NewReader(`{"username":"admin","password":"pass12345","confirm_password":"pass12345","setup_token":"setup-123"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	token := mustToken(cfg.JWTSecret)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/admin/profile", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), `"username":"admin"`)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, "/api/admin/profile", strings.NewReader(`{"new_username":"root","old_password":"pass12345","new_password":"newpass12345","confirm_password":"newpass12345"}`))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), `"username":"root"`)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/admin/login", strings.NewReader(`{"username":"admin","password":"pass12345"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusUnauthorized, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/admin/login", strings.NewReader(`{"username":"root","password":"newpass12345"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestAdminProfile_UpdateRequiresOldPassword(t *testing.T) {
	t.Setenv("PANEL_SETUP_TOKEN", "setup-123")
	store := setupStore(t)

	cfg := config.Load()
	cfg.JWTSecret = "secret"
	r := api.NewRouter(cfg, store)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/admin/bootstrap", strings.NewReader(`{"username":"admin","password":"pass12345","confirm_password":"pass12345","setup_token":"setup-123"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	token := mustToken(cfg.JWTSecret)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, "/api/admin/profile", strings.NewReader(`{"new_username":"root","old_password":"wrong-pass","new_password":"newpass12345","confirm_password":"newpass12345"}`))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware(t *testing.T) {
	cfg := config.Config{JWTSecret: "secret"}
	r := api.NewRouter(cfg, nil)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCORSPreflight(t *testing.T) {
	// Set dummy env so config.Load doesn't complain in future refactors.
	os.Setenv("PANEL_SETUP_TOKEN", "setup-123")
	cfg := config.Config{JWTSecret: "secret"}
	r := api.NewRouter(cfg, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/api/admin/login", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "content-type")

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusNoContent, w.Code)
	require.Equal(t, "http://localhost:5173", w.Header().Get("Access-Control-Allow-Origin"))
	require.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "POST")
	require.Contains(t, strings.ToLower(w.Header().Get("Access-Control-Allow-Headers")), "content-type")
}
