package api_test

import (
  "bytes"
  "encoding/json"
  "fmt"
  "io"
  "net/http"
  "net/http/httptest"
  "strings"
  "sync/atomic"
  "testing"

  "sboard/panel/internal/api"
  "sboard/panel/internal/config"
  "sboard/panel/internal/node"
  "github.com/stretchr/testify/require"
)

type syncDTO struct {
  Status string `json:"status"`
  Error  string `json:"error"`
}

type inboundRespWithSync struct {
  Data inboundDTO `json:"data"`
  Sync syncDTO    `json:"sync"`
}

func TestInboundsCreate_ValidatesShadowsocksMethodAndAutoSyncs(t *testing.T) {
  var syncCalls int32
  restore := api.SetNodeClientFactoryForTest(func() *node.Client {
    return node.NewClient(&fakeDoerFunc{do: func(req *http.Request) (*http.Response, error) {
      if req.URL.Path == "/api/config/sync" && req.Method == http.MethodPost {
        atomic.AddInt32(&syncCalls, 1)
        return &http.Response{
          StatusCode: http.StatusOK,
          Header:     http.Header{"Content-Type": []string{"application/json"}},
          Body:       io.NopCloser(strings.NewReader(`{"status":"ok"}`)),
        }, nil
      }
      if req.URL.Path == "/api/health" && req.Method == http.MethodGet {
        return &http.Response{
          StatusCode: http.StatusOK,
          Header:     http.Header{"Content-Type": []string{"application/json"}},
          Body:       io.NopCloser(strings.NewReader(`{"status":"ok"}`)),
        }, nil
      }
      return &http.Response{StatusCode: http.StatusNotFound, Body: io.NopCloser(strings.NewReader("not found"))}, nil
    }})
  })
  t.Cleanup(restore)

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
  req = httptest.NewRequest(http.MethodPost, "/api/nodes", bytes.NewReader([]byte(fmt.Sprintf(
    `{"name":"n1","api_address":"127.0.0.1","api_port":3000,"secret_key":"secret","public_address":"a.example.com","group_id":%d}`,
    g.Data.ID,
  ))))
  req.Header.Set("Authorization", "Bearer "+token)
  r.ServeHTTP(w, req)
  require.Equal(t, http.StatusCreated, w.Code)
  var n nodeResp
  require.NoError(t, json.Unmarshal(w.Body.Bytes(), &n))

  // Missing method should be rejected.
  w = httptest.NewRecorder()
  req = httptest.NewRequest(http.MethodPost, "/api/inbounds", strings.NewReader(fmt.Sprintf(
    `{"node_id":%d,"tag":"ss-in","protocol":"shadowsocks","listen_port":443,"public_port":443,"settings":{}}`,
    n.Data.ID,
  )))
  req.Header.Set("Authorization", "Bearer "+token)
  r.ServeHTTP(w, req)
  require.Equal(t, http.StatusBadRequest, w.Code)
  require.Equal(t, int32(0), atomic.LoadInt32(&syncCalls))

  // Valid method should be accepted and auto-synced once.
  w = httptest.NewRecorder()
  req = httptest.NewRequest(http.MethodPost, "/api/inbounds", strings.NewReader(fmt.Sprintf(
    `{"node_id":%d,"tag":"ss-in","protocol":"shadowsocks","listen_port":443,"public_port":443,"settings":{"method":"aes-128-gcm"}}`,
    n.Data.ID,
  )))
  req.Header.Set("Authorization", "Bearer "+token)
  r.ServeHTTP(w, req)
  require.Equal(t, http.StatusCreated, w.Code)
  var created inboundRespWithSync
  require.NoError(t, json.Unmarshal(w.Body.Bytes(), &created))
  require.Equal(t, "ok", created.Sync.Status)
  require.Equal(t, int32(1), atomic.LoadInt32(&syncCalls))
}

type fakeDoerFunc struct {
  do func(req *http.Request) (*http.Response, error)
}

func (d *fakeDoerFunc) Do(req *http.Request) (*http.Response, error) {
  return d.do(req)
}
