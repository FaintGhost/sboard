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
