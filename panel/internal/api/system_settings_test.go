package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"sboard/panel/internal/api"
	"sboard/panel/internal/config"
)

type systemSettingsResp struct {
	Data struct {
		SubscriptionBaseURL string `json:"subscription_base_url"`
	} `json:"data"`
	Error string `json:"error"`
}

func TestSystemSettingsAPI_GetDefaultEmpty(t *testing.T) {
	cfg := config.Config{JWTSecret: "secret"}
	store := setupStore(t)
	r := api.NewRouter(cfg, store)
	token := mustToken(cfg.JWTSecret)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/system/settings", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp systemSettingsResp
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Equal(t, "", resp.Data.SubscriptionBaseURL)
}

func TestSystemSettingsAPI_PutAndClear(t *testing.T) {
	cfg := config.Config{JWTSecret: "secret"}
	store := setupStore(t)
	r := api.NewRouter(cfg, store)
	token := mustToken(cfg.JWTSecret)

	payload := []byte(`{"subscription_base_url":"https://203.0.113.10:8443/"}`)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/system/settings", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var updated systemSettingsResp
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &updated))
	require.Equal(t, "https://203.0.113.10:8443", updated.Data.SubscriptionBaseURL)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/system/settings", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &updated))
	require.Equal(t, "https://203.0.113.10:8443", updated.Data.SubscriptionBaseURL)

	payload = []byte(`{"subscription_base_url":"   "}`)
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, "/api/system/settings", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &updated))
	require.Equal(t, "", updated.Data.SubscriptionBaseURL)
}

func TestSystemSettingsAPI_ValidateURL(t *testing.T) {
	cfg := config.Config{JWTSecret: "secret"}
	store := setupStore(t)
	r := api.NewRouter(cfg, store)
	token := mustToken(cfg.JWTSecret)

	tests := []struct {
		name     string
		value    string
		contains string
	}{
		{
			name:     "missing scheme",
			value:    "203.0.113.10:8443",
			contains: "invalid subscription_base_url",
		},
		{
			name:     "invalid scheme",
			value:    "ftp://203.0.113.10:8443",
			contains: "must use http or https",
		},
		{
			name:     "domain not allowed",
			value:    "https://sub.example.com:443",
			contains: "must use a valid IP",
		},
		{
			name:     "missing port",
			value:    "https://203.0.113.10",
			contains: "must include port",
		},
		{
			name:     "invalid port",
			value:    "https://203.0.113.10:70000",
			contains: "invalid port",
		},
		{
			name:     "path not allowed",
			value:    "https://203.0.113.10:8443/panel",
			contains: "protocol + ip:port",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := []byte(`{"subscription_base_url":"` + tt.value + `"}`)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPut, "/api/system/settings", bytes.NewReader(payload))
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			require.Equal(t, http.StatusBadRequest, w.Code)
			require.Contains(t, w.Body.String(), tt.contains)
		})
	}
}
