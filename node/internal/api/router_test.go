package api_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"sboard/node/internal/api"
)

type fakeCore struct{ err error }

func (f *fakeCore) ApplyConfig(ctx *gin.Context, body []byte) error { return f.err }

func TestHealth(t *testing.T) {
	r := api.NewRouter("secret", nil, nil)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestConfigSyncAuth(t *testing.T) {
	r := api.NewRouter("secret", &fakeCore{}, nil)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/config/sync", strings.NewReader(`{"inbounds":[]}`))
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusUnauthorized, w.Code)
}
