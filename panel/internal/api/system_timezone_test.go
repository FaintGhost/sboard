package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"sboard/panel/internal/api"
	"sboard/panel/internal/config"
	"sboard/panel/internal/db"
)

type systemSettingsWithTimezoneResp struct {
	Data struct {
		SubscriptionBaseURL string `json:"subscription_base_url"`
		Timezone            string `json:"timezone"`
	} `json:"data"`
	Error string `json:"error"`
}

func TestSystemSettingsAPI_TimezoneDefaultAndUpdate(t *testing.T) {
	cfg := config.Config{JWTSecret: "secret"}
	store := setupStore(t)
	r := api.NewRouter(cfg, store)
	token := mustToken(cfg.JWTSecret)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/system/settings", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var got systemSettingsWithTimezoneResp
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	require.Equal(t, "", got.Data.SubscriptionBaseURL)
	require.Equal(t, "UTC", got.Data.Timezone)

	payload := []byte(`{"subscription_base_url":"https://203.0.113.10:8443","timezone":"Asia/Shanghai"}`)
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, "/api/system/settings", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	require.Equal(t, "https://203.0.113.10:8443", got.Data.SubscriptionBaseURL)
	require.Equal(t, "Asia/Shanghai", got.Data.Timezone)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/system/settings", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	require.Equal(t, "Asia/Shanghai", got.Data.Timezone)
}

func TestSystemSettingsAPI_TimezoneAffectsTrafficTimestamp(t *testing.T) {
	cfg := config.Config{JWTSecret: "secret"}
	store := setupStore(t)
	r := api.NewRouter(cfg, store)
	token := mustToken(cfg.JWTSecret)

	node, err := store.CreateNode(context.Background(), db.NodeCreate{
		Name:          "n1",
		APIAddress:    "127.0.0.1",
		APIPort:       3003,
		SecretKey:     "secret",
		PublicAddress: "example.com",
	})
	require.NoError(t, err)

	at := time.Date(2026, 2, 8, 8, 0, 0, 0, time.UTC)
	_, err = store.DB.Exec(
		"INSERT INTO traffic_stats (user_id, node_id, inbound_tag, upload, download, recorded_at) VALUES (NULL, ?, ?, ?, ?, ?)",
		node.ID,
		"ss-in",
		int64(1024),
		int64(2048),
		at.Format(time.RFC3339),
	)
	require.NoError(t, err)

	payload := []byte(`{"subscription_base_url":"","timezone":"Asia/Shanghai"}`)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/system/settings", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/nodes/%d/traffic?limit=1&offset=0", node.ID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		Data []struct {
			RecordedAt string `json:"recorded_at"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Len(t, resp.Data, 1)
	require.Equal(t, "2026-02-08T16:00:00+08:00", resp.Data[0].RecordedAt)
}

func TestSystemSettingsAPI_TimezoneValidation(t *testing.T) {
	cfg := config.Config{JWTSecret: "secret"}
	store := setupStore(t)
	r := api.NewRouter(cfg, store)
	token := mustToken(cfg.JWTSecret)

	payload := []byte(`{"subscription_base_url":"","timezone":"Mars/Base"}`)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/system/settings", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Contains(t, w.Body.String(), "invalid timezone")
}
