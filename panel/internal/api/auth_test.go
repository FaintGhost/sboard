package api_test

import (
  "net/http"
  "net/http/httptest"
  "strings"
  "testing"

  "sboard/panel/internal/api"
  "sboard/panel/internal/config"
  "github.com/stretchr/testify/require"
)

func TestAdminLogin(t *testing.T) {
  cfg := config.Config{AdminUser: "admin", AdminPass: "pass", JWTSecret: "secret"}
  r := api.NewRouter(cfg, nil)

  w := httptest.NewRecorder()
  req := httptest.NewRequest(http.MethodPost, "/api/admin/login", strings.NewReader(`{"username":"admin","password":"pass"}`))
  r.ServeHTTP(w, req)
  require.Equal(t, http.StatusOK, w.Code)

  w = httptest.NewRecorder()
  req = httptest.NewRequest(http.MethodPost, "/api/admin/login", strings.NewReader(`{"username":"admin","password":"wrong"}`))
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
  cfg := config.Config{AdminUser: "admin", AdminPass: "pass", JWTSecret: "secret"}
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
