package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"sboard/node/internal/stats"
)

type fakeInboundProvider struct {
	items       []stats.InboundTraffic
	calls       int
	latestReset bool
}

func (f *fakeInboundProvider) InboundTraffic(reset bool) []stats.InboundTraffic {
	f.calls++
	f.latestReset = reset
	return f.items
}

type fakeInboundProviderWithMeta struct {
	*fakeInboundProvider
	meta stats.InboundTrafficMeta
}

func (f *fakeInboundProviderWithMeta) InboundTrafficMeta() stats.InboundTrafficMeta {
	return f.meta
}

func TestStatsInboundsGetBranches(t *testing.T) {
	t.Run("unauthorized", func(t *testing.T) {
		r := newTestRouter(&fakeCore{}, &fakeInboundProvider{})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/stats/inbounds", nil)
		r.ServeHTTP(w, req)
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("provider nil", func(t *testing.T) {
		r := newTestRouter(&fakeCore{}, nil)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/stats/inbounds", nil)
		req.Header.Set("Authorization", "Bearer secret")
		r.ServeHTTP(w, req)
		require.Equal(t, http.StatusInternalServerError, w.Code)
		require.Contains(t, w.Body.String(), "stats not ready")
	})

	t.Run("reset query", func(t *testing.T) {
		provider := &fakeInboundProvider{
			items: []stats.InboundTraffic{{Tag: "ss-in", User: "alice", Uplink: 1, Downlink: 2}},
		}
		r := newTestRouter(&fakeCore{}, provider)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/stats/inbounds?reset=true", nil)
		req.Header.Set("Authorization", "Bearer secret")
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		require.Equal(t, 1, provider.calls)
		require.True(t, provider.latestReset)

		var body map[string]any
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		require.Equal(t, true, body["reset"])
	})

	t.Run("meta included", func(t *testing.T) {
		provider := &fakeInboundProviderWithMeta{
			fakeInboundProvider: &fakeInboundProvider{
				items: []stats.InboundTraffic{{Tag: "ss-in", Uplink: 3, Downlink: 4}},
			},
			meta: stats.InboundTrafficMeta{TrackedTags: 1, TCPConns: 2, UDPConns: 3},
		}
		r := newTestRouter(&fakeCore{}, provider)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/stats/inbounds?reset=1", nil)
		req.Header.Set("Authorization", "Bearer secret")
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		require.Equal(t, 1, provider.calls)
		require.True(t, provider.latestReset)
		require.Contains(t, w.Body.String(), "meta")
		require.Contains(t, w.Body.String(), "tracked_tags")
	})
}
