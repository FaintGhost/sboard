package api_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"sboard/node/internal/api"
)

func TestLegacyNodeRESTEndpointsUnavailable(t *testing.T) {
	rpcHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	r := api.NewRouterWithRPC("secret", &fakeCore{}, nil, rpcHandler)

	for _, tc := range []struct {
		method string
		path   string
		body   string
	}{
		{method: http.MethodGet, path: "/api/health"},
		{method: http.MethodPost, path: "/api/config/sync", body: `{"inbounds":[]}`},
		{method: http.MethodGet, path: "/api/stats/traffic"},
		{method: http.MethodGet, path: "/api/stats/inbounds"},
	} {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
		r.ServeHTTP(w, req)
		require.Equal(t, http.StatusNotFound, w.Code, tc.path)
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/rpc/sboard.node.v1.NodeControlService/Health", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
}
