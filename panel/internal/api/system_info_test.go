package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"sboard/panel/internal/api"
	"sboard/panel/internal/buildinfo"
	"sboard/panel/internal/config"
)

func TestSystemInfoGet(t *testing.T) {
	buildinfo.PanelVersion = "v0.2.0"
	buildinfo.PanelCommitID = "abc1234"
	buildinfo.SingBoxVersion = "1.12.19"
	t.Cleanup(func() {
		buildinfo.PanelVersion = "N/A"
		buildinfo.PanelCommitID = "N/A"
		buildinfo.SingBoxVersion = "N/A"
	})

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
	buildinfo.PanelVersion = ""
	buildinfo.PanelCommitID = ""
	buildinfo.SingBoxVersion = ""
	t.Cleanup(func() {
		buildinfo.PanelVersion = "N/A"
		buildinfo.PanelCommitID = "N/A"
		buildinfo.SingBoxVersion = "N/A"
	})

	r := api.NewRouter(config.Config{JWTSecret: "secret"}, setupStore(t))
	token := mustToken("secret")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/system/info", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), "N/A")
}
