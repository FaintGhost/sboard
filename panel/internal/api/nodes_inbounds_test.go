package api_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/require"
	"sboard/panel/internal/api"
	"sboard/panel/internal/config"
	"sboard/panel/internal/node"
	nodev1 "sboard/panel/internal/rpc/gen/sboard/node/v1"
)

type nodeDTO struct {
	ID            int64  `json:"id"`
	UUID          string `json:"uuid"`
	Name          string `json:"name"`
	APIAddress    string `json:"api_address"`
	APIPort       int    `json:"api_port"`
	SecretKey     string `json:"secret_key"`
	PublicAddress string `json:"public_address"`
	GroupID       *int64 `json:"group_id"`
	Status        string `json:"status"`
}

type nodeResp struct {
	Data nodeDTO `json:"data"`
}

type nodesListResp struct {
	Data []nodeDTO `json:"data"`
}

type inboundDTO struct {
	ID                int64           `json:"id"`
	UUID              string          `json:"uuid"`
	Tag               string          `json:"tag"`
	NodeID            int64           `json:"node_id"`
	Protocol          string          `json:"protocol"`
	ListenPort        int             `json:"listen_port"`
	PublicPort        int             `json:"public_port"`
	Settings          json.RawMessage `json:"settings"`
	TLSSettings       json.RawMessage `json:"tls_settings"`
	TransportSettings json.RawMessage `json:"transport_settings"`
}

type inboundResp struct {
	Data inboundDTO `json:"data"`
}

type inboundsListResp struct {
	Data []inboundDTO `json:"data"`
}

func TestNodesAndInboundsAPI_Basic(t *testing.T) {
	cfg := config.Config{JWTSecret: "secret"}
	store := setupStore(t)
	r := api.NewRouter(cfg, store)
	token := mustToken(cfg.JWTSecret)

	// Create group.
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/groups", strings.NewReader(`{"name":"g1","description":""}`))
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)
	var g groupResp
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &g))

	// Create node.
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/nodes", strings.NewReader(fmt.Sprintf(`{"name":"n1","api_address":"api.local","api_port":2222,"secret_key":"secret","public_address":"a.example.com","group_id":%d}`, g.Data.ID)))
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)
	var n nodeResp
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &n))
	require.Equal(t, "n1", n.Data.Name)
	require.NotNil(t, n.Data.GroupID)

	// List nodes.
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/nodes?limit=10&offset=0", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	var nodes nodesListResp
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &nodes))
	require.Len(t, nodes.Data, 1)

	// Create inbound.
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/inbounds", strings.NewReader(fmt.Sprintf(`{"node_id":%d,"tag":"vless-in","protocol":"vless","listen_port":443,"public_port":0,"settings":{}}`, n.Data.ID)))
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)
	var inb inboundResp
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &inb))
	require.Equal(t, "vless", inb.Data.Protocol)

	// List inbounds by node.
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/inbounds?limit=10&offset=0&node_id=%d", n.Data.ID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	var inbs inboundsListResp
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &inbs))
	require.Len(t, inbs.Data, 1)
}

type forceDeleteDoer struct{}

func (d *forceDeleteDoer) Do(req *http.Request) (*http.Response, error) {
	return serveNodeRPCRequest(req, nodeRPCServiceStub{
		syncConfigFunc: func(_ context.Context, reqBody *nodev1.SyncConfigRequest) (*nodev1.SyncConfigResponse, error) {
			if req.Header.Get("Authorization") != "Bearer secret" {
				return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("unauthorized"))
			}
			var body struct {
				Inbounds []map[string]any `json:"inbounds"`
			}
			if err := json.Unmarshal(reqBody.GetPayloadJson(), &body); err != nil {
				return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("bad json"))
			}
			if len(body.Inbounds) != 0 {
				return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("expected empty inbounds"))
			}
			return &nodev1.SyncConfigResponse{Status: "ok"}, nil
		},
	}, nil)
}

func TestNodesDelete_ForceDrainsNodeAndDeletesInbounds(t *testing.T) {
	restore := api.SetNodeClientFactoryForTest(func() *node.Client {
		return node.NewClient(&forceDeleteDoer{})
	})
	t.Cleanup(restore)

	cfg := config.Config{JWTSecret: "secret"}
	store := setupStore(t)
	r := api.NewRouter(cfg, store)
	token := mustToken(cfg.JWTSecret)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/groups", strings.NewReader(`{"name":"g1","description":""}`))
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)
	var g groupResp
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &g))

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/nodes", strings.NewReader(fmt.Sprintf(`{"name":"n1","api_address":"api.local","api_port":2222,"secret_key":"secret","public_address":"a.example.com","group_id":%d}`, g.Data.ID)))
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)
	var n nodeResp
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &n))

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/inbounds", strings.NewReader(fmt.Sprintf(`{"node_id":%d,"tag":"vless-in","protocol":"vless","listen_port":443,"public_port":0,"settings":{}}`, n.Data.ID)))
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/nodes/%d?force=true", n.Data.ID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/nodes/%d", n.Data.ID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusNotFound, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/inbounds?limit=10&offset=0&node_id=%d", n.Data.ID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	var inbs inboundsListResp
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &inbs))
	require.Len(t, inbs.Data, 0)
}
