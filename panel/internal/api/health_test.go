package api_test

import (
  "net/http"
  "net/http/httptest"
  "testing"

  "sboard/panel/internal/api"
  "sboard/panel/internal/config"
  "github.com/stretchr/testify/require"
)

func TestHealth(t *testing.T) {
  r := api.NewRouter(config.Config{}, nil)
  w := httptest.NewRecorder()
  req := httptest.NewRequest(http.MethodGet, "/api/health", nil)

  r.ServeHTTP(w, req)

  require.Equal(t, http.StatusOK, w.Code)
  require.Contains(t, w.Body.String(), "ok")
}
