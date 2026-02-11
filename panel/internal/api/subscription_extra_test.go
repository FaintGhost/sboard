package api_test

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"sboard/panel/internal/api"
	"sboard/panel/internal/config"
)

func TestSubscription_InvalidFormatAndNotFound(t *testing.T) {
	store := setupSubscriptionStore(t)
	userUUID := seedSubscriptionData(t, store)
	r := api.NewRouter(config.Config{}, store)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/sub/"+userUUID+"?format=badfmt", nil)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/sub/not-exists-user-uuid", nil)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestSubscription_RejectsDisabledUser(t *testing.T) {
	store := setupSubscriptionStore(t)
	userUUID := seedSubscriptionData(t, store)

	_, err := store.DB.Exec("UPDATE users SET status = 'disabled' WHERE uuid = ?", userUUID)
	require.NoError(t, err)

	r := api.NewRouter(config.Config{}, store)
	req := httptest.NewRequest(http.MethodGet, "/api/sub/"+userUUID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestSubscription_DefaultByUAAndExplicitV2Ray(t *testing.T) {
	store := setupSubscriptionStore(t)
	userUUID := seedSubscriptionData(t, store)
	r := api.NewRouter(config.Config{}, store)

	// UA is sing-box family => should return json (singbox)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/sub/"+userUUID, nil)
	req.Header.Set("User-Agent", "sing-box 1.11")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Header().Get("Content-Type"), "application/json")

	// Explicit v2ray format should return base64 plain text
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/sub/"+userUUID+"?format=v2ray", nil)
	req.Header.Set("User-Agent", "sing-box 1.11")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Header().Get("Content-Type"), "text/plain")

	decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(w.Body.String()))
	require.NoError(t, err)
	require.Contains(t, string(decoded), "a.example.com")
}
