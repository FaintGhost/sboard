package api_test

import (
  "net/http"
  "net/http/httptest"
  "os"
  "path/filepath"
  "testing"

  "github.com/stretchr/testify/require"

  "sboard/panel/internal/api"
  "sboard/panel/internal/config"
)

func TestServeWebUI_IndexAndFallback(t *testing.T) {
  store := setupStore(t)

  dir := t.TempDir()
  webDir := filepath.Join(dir, "dist")
  require.NoError(t, os.MkdirAll(filepath.Join(webDir, "assets"), 0o755))
  require.NoError(t, os.WriteFile(filepath.Join(webDir, "index.html"), []byte("<html>ok</html>"), 0o644))

  cfg := config.Config{
    AdminUser: "admin",
    AdminPass: "pass",
    JWTSecret: "secret",
    ServeWeb:  true,
    WebDir:    webDir,
  }

  r := api.NewRouter(cfg, store)

  w := httptest.NewRecorder()
  req := httptest.NewRequest(http.MethodGet, "/", nil)
  r.ServeHTTP(w, req)
  require.Equal(t, http.StatusOK, w.Code)
  require.Contains(t, w.Body.String(), "ok")

  w = httptest.NewRecorder()
  req = httptest.NewRequest(http.MethodGet, "/users", nil)
  r.ServeHTTP(w, req)
  require.Equal(t, http.StatusOK, w.Code)
  require.Contains(t, w.Body.String(), "ok")
}

func TestServeWebUI_DoesNotCatchAPIRoutes(t *testing.T) {
  store := setupStore(t)

  dir := t.TempDir()
  webDir := filepath.Join(dir, "dist")
  require.NoError(t, os.MkdirAll(filepath.Join(webDir, "assets"), 0o755))
  require.NoError(t, os.WriteFile(filepath.Join(webDir, "index.html"), []byte("<html>ok</html>"), 0o644))

  cfg := config.Config{
    AdminUser: "admin",
    AdminPass: "pass",
    JWTSecret: "secret",
    ServeWeb:  true,
    WebDir:    webDir,
  }

  r := api.NewRouter(cfg, store)

  w := httptest.NewRecorder()
  req := httptest.NewRequest(http.MethodGet, "/api/unknown", nil)
  r.ServeHTTP(w, req)
  require.Equal(t, http.StatusNotFound, w.Code)
}

