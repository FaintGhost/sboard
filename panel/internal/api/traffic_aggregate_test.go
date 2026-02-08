package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"sboard/panel/internal/api"
	"sboard/panel/internal/config"
	"sboard/panel/internal/db"
)

func TestTrafficTotalSummary_WindowAll(t *testing.T) {
	cfg := config.Config{JWTSecret: "secret"}
	store := setupStore(t)
	r := api.NewRouter(cfg, store)
	token := mustToken(cfg.JWTSecret)

	n, err := store.CreateNode(t.Context(), db.NodeCreate{
		Name:          "n1",
		APIAddress:    "127.0.0.1",
		APIPort:       3003,
		SecretKey:     "secret",
		PublicAddress: "example.com",
	})
	require.NoError(t, err)

	now := time.Date(2026, 2, 8, 8, 0, 0, 0, time.UTC)
	_, err = store.DB.Exec(
		"INSERT INTO traffic_stats (user_id, node_id, inbound_tag, upload, download, recorded_at) VALUES (NULL, ?, ?, ?, ?, ?)",
		n.ID,
		"ss-in",
		int64(1024),
		int64(2048),
		now.Format(time.RFC3339),
	)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/traffic/total/summary?window=all", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		Data struct {
			Upload   int64 `json:"upload"`
			Download int64 `json:"download"`
			Samples  int64 `json:"samples"`
			Nodes    int64 `json:"nodes"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Equal(t, int64(1024), resp.Data.Upload)
	require.Equal(t, int64(2048), resp.Data.Download)
	require.Equal(t, int64(1), resp.Data.Samples)
	require.Equal(t, int64(1), resp.Data.Nodes)
}
