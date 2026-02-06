package api_test

import (
  "bytes"
  "log"
  "net/http"
  "net/http/httptest"
  "testing"

  "sboard/panel/internal/api"
  "sboard/panel/internal/config"
  "github.com/stretchr/testify/require"
)

func TestRequestLoggerPrintsAPI(t *testing.T) {
  cfg := config.Config{LogRequests: true, JWTSecret: "secret"}
  r := api.NewRouter(cfg, nil)

  var buf bytes.Buffer
  oldOut := log.Writer()
  oldFlags := log.Flags()
  log.SetOutput(&buf)
  log.SetFlags(0)
  t.Cleanup(func() {
    log.SetOutput(oldOut)
    log.SetFlags(oldFlags)
  })

  w := httptest.NewRecorder()
  req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
  req.Header.Set("Authorization", "Bearer secret-token-should-not-be-logged")
  r.ServeHTTP(w, req)
  require.Equal(t, http.StatusOK, w.Code)

  got := buf.String()
  require.Contains(t, got, "GET /api/health")
  require.Contains(t, got, "-> 200")
  require.NotContains(t, got, "secret-token-should-not-be-logged")
}

