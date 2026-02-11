package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"sboard/panel/internal/api"
	"sboard/panel/internal/config"
	"sboard/panel/internal/db"
)

func TestTrafficNodesSummary_WindowFilterAndAll(t *testing.T) {
	cfg := config.Config{JWTSecret: "secret"}
	store := setupStore(t)
	now := time.Date(2026, 2, 10, 12, 0, 0, 0, time.UTC)
	store.Now = func() time.Time { return now }

	r := api.NewRouter(cfg, store)
	token := mustToken(cfg.JWTSecret)

	n1, err := store.CreateNode(t.Context(), db.NodeCreate{
		Name:          "n1",
		APIAddress:    "127.0.0.1",
		APIPort:       3001,
		SecretKey:     "secret",
		PublicAddress: "example1.com",
	})
	require.NoError(t, err)

	n2, err := store.CreateNode(t.Context(), db.NodeCreate{
		Name:          "n2",
		APIAddress:    "127.0.0.1",
		APIPort:       3002,
		SecretKey:     "secret",
		PublicAddress: "example2.com",
	})
	require.NoError(t, err)

	_, err = store.DB.Exec(
		"INSERT INTO traffic_stats (user_id, node_id, inbound_tag, upload, download, recorded_at) VALUES (NULL, ?, ?, ?, ?, ?)",
		n1.ID, "in-a", int64(100), int64(200), now.Add(-1*time.Hour).Format(time.RFC3339),
	)
	require.NoError(t, err)
	_, err = store.DB.Exec(
		"INSERT INTO traffic_stats (user_id, node_id, inbound_tag, upload, download, recorded_at) VALUES (NULL, ?, ?, ?, ?, ?)",
		n2.ID, "in-b", int64(300), int64(400), now.Add(-48*time.Hour).Format(time.RFC3339),
	)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/traffic/nodes/summary?window=24h", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var resp24 struct {
		Data []struct {
			NodeID   int64 `json:"node_id"`
			Upload   int64 `json:"upload"`
			Download int64 `json:"download"`
			Samples  int64 `json:"samples"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp24))
	require.Len(t, resp24.Data, 2)

	var n1Found bool
	for _, item := range resp24.Data {
		if item.NodeID == n1.ID {
			n1Found = true
			require.Equal(t, int64(100), item.Upload)
			require.Equal(t, int64(200), item.Download)
			require.Equal(t, int64(1), item.Samples)
		}
		if item.NodeID == n2.ID {
			require.Equal(t, int64(0), item.Upload)
			require.Equal(t, int64(0), item.Download)
			require.Equal(t, int64(0), item.Samples)
		}
	}
	require.True(t, n1Found)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/traffic/nodes/summary?window=all", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var respAll struct {
		Data []struct {
			NodeID int64 `json:"node_id"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &respAll))
	require.Len(t, respAll.Data, 2)
}

func TestTrafficTimeseries_NodeFilterAndValidation(t *testing.T) {
	cfg := config.Config{JWTSecret: "secret"}
	store := setupStore(t)
	now := time.Date(2026, 2, 10, 12, 0, 0, 0, time.UTC)
	store.Now = func() time.Time { return now }

	r := api.NewRouter(cfg, store)
	token := mustToken(cfg.JWTSecret)

	n1, err := store.CreateNode(t.Context(), db.NodeCreate{
		Name:          "n1",
		APIAddress:    "127.0.0.1",
		APIPort:       3011,
		SecretKey:     "secret",
		PublicAddress: "example1.com",
	})
	require.NoError(t, err)

	n2, err := store.CreateNode(t.Context(), db.NodeCreate{
		Name:          "n2",
		APIAddress:    "127.0.0.1",
		APIPort:       3012,
		SecretKey:     "secret",
		PublicAddress: "example2.com",
	})
	require.NoError(t, err)

	_, err = store.DB.Exec(
		"INSERT INTO traffic_stats (user_id, node_id, inbound_tag, upload, download, recorded_at) VALUES (NULL, ?, ?, ?, ?, ?)",
		n1.ID, "in-a", int64(10), int64(20), now.Add(-10*time.Minute).Format(time.RFC3339),
	)
	require.NoError(t, err)
	_, err = store.DB.Exec(
		"INSERT INTO traffic_stats (user_id, node_id, inbound_tag, upload, download, recorded_at) VALUES (NULL, ?, ?, ?, ?, ?)",
		n2.ID, "in-b", int64(30), int64(40), now.Add(-9*time.Minute).Format(time.RFC3339),
	)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/traffic/timeseries?window=all&bucket=hour", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var total struct {
		Data []struct {
			Upload   int64 `json:"upload"`
			Download int64 `json:"download"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &total))
	require.Len(t, total.Data, 1)
	require.Equal(t, int64(40), total.Data[0].Upload)
	require.Equal(t, int64(60), total.Data[0].Download)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/traffic/timeseries?window=all&bucket=hour&node_id="+itoa(n1.ID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var filtered struct {
		Data []struct {
			Upload   int64 `json:"upload"`
			Download int64 `json:"download"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &filtered))
	require.Len(t, filtered.Data, 1)
	require.Equal(t, int64(10), filtered.Data[0].Upload)
	require.Equal(t, int64(20), filtered.Data[0].Download)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/traffic/timeseries?window=all&bucket=week", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/traffic/timeseries?window=all&bucket=hour&node_id=abc", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/traffic/nodes/summary?window=30s", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)
}

func itoa(v int64) string {
	return strconv.FormatInt(v, 10)
}
