package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"sboard/panel/internal/api"
	"sboard/panel/internal/config"
)

func TestTrafficEndpoints_StoreNotReady(t *testing.T) {
	cfg := config.Config{JWTSecret: "secret"}
	r := api.NewRouter(cfg, nil)
	token := mustToken(cfg.JWTSecret)

	paths := []string{
		"/api/nodes/1/traffic",
		"/api/traffic/nodes/summary",
		"/api/traffic/total/summary",
		"/api/traffic/timeseries",
	}

	for _, path := range paths {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusInternalServerError, w.Code, "path=%s", path)
		require.Contains(t, w.Body.String(), "store not ready", "path=%s", path)
	}
}
