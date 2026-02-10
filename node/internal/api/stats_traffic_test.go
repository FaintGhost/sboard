package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"sboard/node/internal/stats"
)

func pickUsableInterface(t *testing.T) string {
	t.Helper()
	if _, _, err := stats.ReadNetDev("lo"); err == nil {
		return "lo"
	}
	iface, err := stats.DetectDefaultInterface()
	if err == nil {
		return iface
	}
	return ""
}

func TestStatsTrafficGetBranches(t *testing.T) {
	t.Run("unauthorized", func(t *testing.T) {
		r := newTestRouter(&fakeCore{}, nil)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/stats/traffic", nil)
		r.ServeHTTP(w, req)
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("env fallback error", func(t *testing.T) {
		t.Setenv("NODE_TRAFFIC_INTERFACE", "definitely-not-existing-iface")
		r := newTestRouter(&fakeCore{}, nil)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/stats/traffic", nil)
		req.Header.Set("Authorization", "Bearer secret")
		r.ServeHTTP(w, req)
		require.Equal(t, http.StatusInternalServerError, w.Code)
		require.Contains(t, w.Body.String(), "definitely-not-existing-iface")
	})

	t.Run("query override env", func(t *testing.T) {
		t.Setenv("NODE_TRAFFIC_INTERFACE", "definitely-not-existing-iface")
		r := newTestRouter(&fakeCore{}, nil)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/stats/traffic?interface=missing-iface-2", nil)
		req.Header.Set("Authorization", "Bearer secret")
		r.ServeHTTP(w, req)
		require.Equal(t, http.StatusInternalServerError, w.Code)
		require.Contains(t, w.Body.String(), "missing-iface-2")
	})

	t.Run("success", func(t *testing.T) {
		iface := pickUsableInterface(t)
		if iface == "" {
			t.Skip("no usable interface found for stats test")
		}

		r := newTestRouter(&fakeCore{}, nil)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/stats/traffic?interface="+iface, nil)
		req.Header.Set("Authorization", "Bearer secret")
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		var sample stats.TrafficSample
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &sample))
		require.Equal(t, iface, sample.Interface)
	})
}
