package api_test

import (
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
)

func TestUsersCreate_ValidationAndConflict(t *testing.T) {
	cfg := config.Config{JWTSecret: "secret"}
	store := setupStore(t)
	r := api.NewRouter(cfg, store)
	token := mustToken(cfg.JWTSecret)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(`{`))
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(`{"username":"   "}`))
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(`{"username":"alice-create"}`))
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(`{"username":"alice-create"}`))
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusConflict, w.Code)
}

func TestGroupsCreate_ValidationAndConflict(t *testing.T) {
	cfg := config.Config{JWTSecret: "secret"}
	store := setupStore(t)
	r := api.NewRouter(cfg, store)
	token := mustToken(cfg.JWTSecret)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/groups", strings.NewReader(`{`))
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/groups", strings.NewReader(`{"name":"   ","description":""}`))
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/groups", strings.NewReader(`{"name":"g-create","description":""}`))
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/groups", strings.NewReader(`{"name":"g-create","description":""}`))
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusConflict, w.Code)
}

func TestNodesCreate_Validation(t *testing.T) {
	cfg := config.Config{JWTSecret: "secret"}
	store := setupStore(t)
	r := api.NewRouter(cfg, store)
	token := mustToken(cfg.JWTSecret)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/nodes", strings.NewReader(`{`))
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)

	badBodies := []string{
		`{"name":"","api_address":"127.0.0.1","api_port":3000,"secret_key":"secret","public_address":"a.example.com"}`,
		`{"name":"n1","api_address":"","api_port":3000,"secret_key":"secret","public_address":"a.example.com"}`,
		`{"name":"n1","api_address":"127.0.0.1","api_port":0,"secret_key":"secret","public_address":"a.example.com"}`,
		`{"name":"n1","api_address":"127.0.0.1","api_port":3000,"secret_key":"","public_address":"a.example.com"}`,
		`{"name":"n1","api_address":"127.0.0.1","api_port":3000,"secret_key":"secret","public_address":""}`,
	}
	for _, body := range badBodies {
		w = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodPost, "/api/nodes", strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		r.ServeHTTP(w, req)
		require.Equal(t, http.StatusBadRequest, w.Code, body)
	}
}

func TestInboundsCreate_ValidationNotFoundAndConflict(t *testing.T) {
	restore := api.SetNodeClientFactoryForTest(func() *node.Client {
		return node.NewClient(&fakeDoerFunc{do: func(req *http.Request) (*http.Response, error) {
			if req.URL.Path == "/api/config/sync" && req.Method == http.MethodPost {
				return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{"status":"ok"}`))}, nil
			}
			if req.URL.Path == "/api/health" && req.Method == http.MethodGet {
				return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{"status":"ok"}`))}, nil
			}
			return &http.Response{StatusCode: http.StatusNotFound, Body: io.NopCloser(strings.NewReader("not found"))}, nil
		}})
	})
	t.Cleanup(restore)

	cfg := config.Config{JWTSecret: "secret"}
	store := setupStore(t)
	r := api.NewRouter(cfg, store)
	token := mustToken(cfg.JWTSecret)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/inbounds", strings.NewReader(`{`))
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)

	badBodies := []string{
		`{"node_id":0,"tag":"vless-in","protocol":"vless","listen_port":443,"public_port":443,"settings":{}}`,
		`{"node_id":1,"tag":"","protocol":"vless","listen_port":443,"public_port":443,"settings":{}}`,
		`{"node_id":1,"tag":"vless-in","protocol":"","listen_port":443,"public_port":443,"settings":{}}`,
		`{"node_id":1,"tag":"vless-in","protocol":"vless","listen_port":0,"public_port":443,"settings":{}}`,
	}
	for _, body := range badBodies {
		w = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodPost, "/api/inbounds", strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		r.ServeHTTP(w, req)
		require.Equal(t, http.StatusBadRequest, w.Code, body)
	}

	// node not found
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/inbounds", strings.NewReader(`{"node_id":99999,"tag":"vless-in","protocol":"vless","listen_port":443,"public_port":443,"settings":{}}`))
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusNotFound, w.Code)

	nodeID := createGroupAndNode(t, r, token)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/inbounds", strings.NewReader(fmt.Sprintf(
		`{"node_id":%d,"tag":"dup-tag","protocol":"vless","listen_port":443,"public_port":443,"settings":{}}`,
		nodeID,
	)))
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/inbounds", strings.NewReader(fmt.Sprintf(
		`{"node_id":%d,"tag":"dup-tag","protocol":"vless","listen_port":8443,"public_port":8443,"settings":{}}`,
		nodeID,
	)))
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusConflict, w.Code)
}
