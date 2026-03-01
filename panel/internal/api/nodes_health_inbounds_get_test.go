package api_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"sboard/panel/internal/api"
	"sboard/panel/internal/config"
	"sboard/panel/internal/node"
	nodev1 "sboard/panel/internal/rpc/gen/sboard/node/v1"
)

func TestNodeHealth_StatusTransitions(t *testing.T) {
	cfg := config.Config{JWTSecret: "secret"}
	store := setupStore(t)

	doer := &fakeDoerFunc{do: func(req *http.Request) (*http.Response, error) {
		return serveNodeRPCRequest(req, nodeRPCServiceStub{}, nil)
	}}
	restore := api.SetNodeClientFactoryForTest(func() *node.Client {
		return node.NewClient(doer)
	})
	t.Cleanup(restore)

	r := api.NewRouter(cfg, store)
	token := mustToken(cfg.JWTSecret)
	nodeID := createGroupAndNode(t, r, token)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/nodes/%d/health", nodeID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	got, err := store.GetNodeByID(t.Context(), nodeID)
	require.NoError(t, err)
	require.Equal(t, "online", got.Status)
	require.NotNil(t, got.LastSeenAt)

	doer.do = func(req *http.Request) (*http.Response, error) {
		return serveNodeRPCRequest(req, nodeRPCServiceStub{
			healthFunc: func(context.Context, *nodev1.HealthRequest) (*nodev1.HealthResponse, error) {
				return nil, io.ErrUnexpectedEOF
			},
		}, nil)
	}

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/nodes/%d/health", nodeID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadGateway, w.Code)

	got, err = store.GetNodeByID(t.Context(), nodeID)
	require.NoError(t, err)
	require.Equal(t, "offline", got.Status)
}

func TestNodeHealth_InvalidIDAndNotFound(t *testing.T) {
	cfg := config.Config{JWTSecret: "secret"}
	store := setupStore(t)
	r := api.NewRouter(cfg, store)
	token := mustToken(cfg.JWTSecret)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/nodes/not-a-number/health", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/nodes/99999/health", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestInboundsGet_BasicAndErrors(t *testing.T) {
	cfg := config.Config{JWTSecret: "secret"}
	store := setupStore(t)

	restore := api.SetNodeClientFactoryForTest(func() *node.Client {
		return node.NewClient(&fakeDoerFunc{do: func(req *http.Request) (*http.Response, error) {
			return serveNodeRPCRequest(req, nodeRPCServiceStub{}, nil)
		}})
	})
	t.Cleanup(restore)

	r := api.NewRouter(cfg, store)
	token := mustToken(cfg.JWTSecret)

	nodeID := createGroupAndNode(t, r, token)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/inbounds", strings.NewReader(fmt.Sprintf(
		`{"node_id":%d,"tag":"vless-in","protocol":"vless","listen_port":443,"public_port":443,"settings":{}}`,
		nodeID,
	)))
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var created inboundResp
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &created))

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/inbounds/%d", created.Data.ID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var got inboundResp
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	require.Equal(t, created.Data.ID, got.Data.ID)
	require.Equal(t, "vless-in", got.Data.Tag)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/inbounds/not-a-number", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/inbounds/99999", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusNotFound, w.Code)
}

func createGroupAndNode(t *testing.T, r http.Handler, token string) int64 {
	t.Helper()

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/groups", strings.NewReader(`{"name":"g1","description":""}`))
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var g groupResp
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &g))

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/nodes", strings.NewReader(fmt.Sprintf(
		`{"name":"n1","api_address":"127.0.0.1","api_port":3000,"secret_key":"secret","public_address":"a.example.com","group_id":%d}`,
		g.Data.ID,
	)))
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var n nodeResp
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &n))
	return n.Data.ID
}
