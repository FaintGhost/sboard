package api_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"sboard/panel/internal/api"
	"sboard/panel/internal/config"
)

func TestSystemInfoGet(t *testing.T) {
	t.Setenv("PANEL_VERSION", "v0.2.0")
	t.Setenv("PANEL_COMMIT_ID", "abc1234")
	t.Setenv("SING_BOX_VERSION", "1.12.19")

	r := api.NewRouter(config.Config{JWTSecret: "secret"}, setupStore(t))
	token := mustToken("secret")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/system/info", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), "v0.2.0")
	require.Contains(t, w.Body.String(), "abc1234")
	require.Contains(t, w.Body.String(), "1.12.19")
}

func TestSystemInfoGet_DefaultNA(t *testing.T) {
	_ = os.Unsetenv("PANEL_VERSION")
	_ = os.Unsetenv("PANEL_COMMIT_ID")
	_ = os.Unsetenv("SING_BOX_VERSION")

	r := api.NewRouter(config.Config{JWTSecret: "secret"}, setupStore(t))
	token := mustToken("secret")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/system/info", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), "N/A")
}
