package api_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"sboard/panel/internal/api"
	"sboard/panel/internal/config"
	"sboard/panel/internal/db"
)

func createNodeForAPITests(t *testing.T, store *db.Store) db.Node {
	t.Helper()
	node, err := store.CreateNode(t.Context(), db.NodeCreate{
		Name:          "node-api",
		APIAddress:    "127.0.0.1",
		APIPort:       3003,
		SecretKey:     "secret",
		PublicAddress: "example.com",
	})
	require.NoError(t, err)
	return node
}

func createInboundForAPITests(t *testing.T, store *db.Store, nodeID int64) db.Inbound {
	t.Helper()
	inbound, err := store.CreateInbound(t.Context(), db.InboundCreate{
		NodeID:     nodeID,
		Tag:        "ss-in",
		Protocol:   "shadowsocks",
		ListenPort: 8388,
		PublicPort: 8388,
		Settings:   json.RawMessage(`{"method":"aes-128-gcm","password":"pass12345"}`),
	})
	require.NoError(t, err)
	return inbound
}

func TestNodesUpdateValidationBranches(t *testing.T) {
	cfg := config.Config{JWTSecret: "secret"}
	store := setupStore(t)
	r := api.NewRouter(cfg, store)
	token := mustToken(cfg.JWTSecret)

	node := createNodeForAPITests(t, store)

	cases := []struct {
		name   string
		path   string
		body   string
		status int
	}{
		{name: "invalid id", path: "/api/nodes/abc", body: `{}`, status: http.StatusBadRequest},
		{name: "invalid body", path: fmt.Sprintf("/api/nodes/%d", node.ID), body: `{`, status: http.StatusBadRequest},
		{name: "invalid name", path: fmt.Sprintf("/api/nodes/%d", node.ID), body: `{"name":"   "}`, status: http.StatusBadRequest},
		{name: "invalid api address", path: fmt.Sprintf("/api/nodes/%d", node.ID), body: `{"api_address":"   "}`, status: http.StatusBadRequest},
		{name: "invalid api port", path: fmt.Sprintf("/api/nodes/%d", node.ID), body: `{"api_port":0}`, status: http.StatusBadRequest},
		{name: "invalid secret", path: fmt.Sprintf("/api/nodes/%d", node.ID), body: `{"secret_key":""}`, status: http.StatusBadRequest},
		{name: "invalid public address", path: fmt.Sprintf("/api/nodes/%d", node.ID), body: `{"public_address":""}`, status: http.StatusBadRequest},
		{name: "not found", path: "/api/nodes/999999", body: `{"name":"n2"}`, status: http.StatusNotFound},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPut, tc.path, strings.NewReader(tc.body))
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)
			require.Equal(t, tc.status, w.Code)
		})
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/nodes/%d", node.ID), strings.NewReader(`{"name":"node-updated","api_port":3009}`))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), "node-updated")
}

func TestGroupsGetAndUpdateValidationBranches(t *testing.T) {
	cfg := config.Config{JWTSecret: "secret"}
	store := setupStore(t)
	r := api.NewRouter(cfg, store)
	token := mustToken(cfg.JWTSecret)

	group, err := store.CreateGroup(t.Context(), "g-api", "desc")
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/groups/abc", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/groups/999999", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusNotFound, w.Code)

	cases := []struct {
		name   string
		path   string
		body   string
		status int
	}{
		{name: "update invalid id", path: "/api/groups/abc", body: `{}`, status: http.StatusBadRequest},
		{name: "update invalid body", path: fmt.Sprintf("/api/groups/%d", group.ID), body: `{`, status: http.StatusBadRequest},
		{name: "update invalid name", path: fmt.Sprintf("/api/groups/%d", group.ID), body: `{"name":"   "}`, status: http.StatusBadRequest},
		{name: "update not found", path: "/api/groups/999999", body: `{"name":"new-name"}`, status: http.StatusNotFound},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPut, tc.path, strings.NewReader(tc.body))
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)
			require.Equal(t, tc.status, w.Code)
		})
	}

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/groups/%d", group.ID), strings.NewReader(`{"name":"g-api-new"}`))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), "g-api-new")
}

func TestInboundsUpdateDeleteValidationBranches(t *testing.T) {
	cfg := config.Config{JWTSecret: "secret"}
	store := setupStore(t)
	r := api.NewRouter(cfg, store)
	token := mustToken(cfg.JWTSecret)

	node := createNodeForAPITests(t, store)
	inbound := createInboundForAPITests(t, store, node.ID)

	updateCases := []struct {
		name   string
		path   string
		body   string
		status int
	}{
		{name: "invalid id", path: "/api/inbounds/abc", body: `{}`, status: http.StatusBadRequest},
		{name: "not found", path: "/api/inbounds/999999", body: `{}`, status: http.StatusNotFound},
		{name: "invalid body", path: fmt.Sprintf("/api/inbounds/%d", inbound.ID), body: `{`, status: http.StatusBadRequest},
		{name: "invalid tag", path: fmt.Sprintf("/api/inbounds/%d", inbound.ID), body: `{"tag":"   "}`, status: http.StatusBadRequest},
		{name: "invalid protocol", path: fmt.Sprintf("/api/inbounds/%d", inbound.ID), body: `{"protocol":"   "}`, status: http.StatusBadRequest},
		{name: "invalid listen_port", path: fmt.Sprintf("/api/inbounds/%d", inbound.ID), body: `{"listen_port":0}`, status: http.StatusBadRequest},
		{name: "invalid public_port", path: fmt.Sprintf("/api/inbounds/%d", inbound.ID), body: `{"public_port":-1}`, status: http.StatusBadRequest},
		{name: "invalid settings", path: fmt.Sprintf("/api/inbounds/%d", inbound.ID), body: `{"settings":""}`, status: http.StatusBadRequest},
	}

	for _, tc := range updateCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPut, tc.path, strings.NewReader(tc.body))
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)
			require.Equal(t, tc.status, w.Code)
		})
	}

	_, err := store.DB.Exec("UPDATE inbounds SET node_id = ? WHERE id = ?", int64(999999), inbound.ID)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/inbounds/%d", inbound.ID), strings.NewReader(`{"tag":"ss-in-updated","settings":{"method":"aes-128-gcm","password":"pass12345"}}`))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), "ss-in-updated")
	require.Contains(t, w.Body.String(), "get node failed")

	deleteCases := []struct {
		name   string
		path   string
		status int
	}{
		{name: "delete invalid id", path: "/api/inbounds/abc", status: http.StatusBadRequest},
		{name: "delete not found", path: "/api/inbounds/999999", status: http.StatusNotFound},
	}
	for _, tc := range deleteCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, tc.path, nil)
			req.Header.Set("Authorization", "Bearer "+token)
			r.ServeHTTP(w, req)
			require.Equal(t, tc.status, w.Code)
		})
	}

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/inbounds/%d", inbound.ID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), "\"status\":\"ok\"")
	require.Contains(t, w.Body.String(), "get node failed")
}

func TestTrafficTimeseriesValidationAndSuccess(t *testing.T) {
	cfg := config.Config{JWTSecret: "secret"}
	store := setupStore(t)
	r := api.NewRouter(cfg, store)
	token := mustToken(cfg.JWTSecret)

	node := createNodeForAPITests(t, store)
	_, err := store.InsertInboundTrafficDelta(t.Context(), node.ID, "ss-in", 100, 200, time.Date(2026, 2, 10, 10, 0, 0, 0, time.UTC))
	require.NoError(t, err)

	cases := []struct {
		name   string
		url    string
		status int
	}{
		{name: "invalid window", url: "/api/traffic/timeseries?window=30s", status: http.StatusBadRequest},
		{name: "invalid bucket", url: "/api/traffic/timeseries?bucket=week", status: http.StatusBadRequest},
		{name: "invalid node_id", url: "/api/traffic/timeseries?node_id=abc", status: http.StatusBadRequest},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, tc.url, nil)
			req.Header.Set("Authorization", "Bearer "+token)
			r.ServeHTTP(w, req)
			require.Equal(t, tc.status, w.Code)
		})
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/traffic/timeseries?window=all&bucket=hour&node_id=%d", node.ID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), "bucket_start")
}
