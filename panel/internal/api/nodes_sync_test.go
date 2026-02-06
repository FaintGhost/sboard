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

type fakeDoer struct {
  got int32
}

func (d *fakeDoer) Do(req *http.Request) (*http.Response, error) {
  if req.URL.Path == "/api/health" && req.Method == http.MethodGet {
    return &http.Response{
      StatusCode: http.StatusOK,
      Header:     http.Header{"Content-Type": []string{"application/json"}},
      Body:       io.NopCloser(strings.NewReader(`{"status":"ok"}`)),
    }, nil
  }

  if req.URL.Path != "/api/config/sync" || req.Method != http.MethodPost {
    return &http.Response{StatusCode: http.StatusNotFound, Body: io.NopCloser(strings.NewReader("not found"))}, nil
  }
  if req.Header.Get("Authorization") != "Bearer secret" {
    return &http.Response{StatusCode: http.StatusUnauthorized, Body: io.NopCloser(strings.NewReader("unauthorized"))}, nil
  }
  b, _ := io.ReadAll(req.Body)
  var body struct {
    Inbounds []map[string]any `json:"inbounds"`
  }
  if err := json.Unmarshal(b, &body); err != nil {
    return &http.Response{StatusCode: http.StatusBadRequest, Body: io.NopCloser(strings.NewReader("bad json"))}, nil
  }
  if len(body.Inbounds) != 1 {
    return &http.Response{StatusCode: http.StatusBadRequest, Body: io.NopCloser(strings.NewReader("bad inbounds"))}, nil
  }
  inb := body.Inbounds[0]
  if inb["type"] != "vless" || inb["tag"] != "vless-in" {
    return &http.Response{StatusCode: http.StatusBadRequest, Body: io.NopCloser(strings.NewReader("bad inbound"))}, nil
  }
  if _, ok := inb["users"]; !ok {
    return &http.Response{StatusCode: http.StatusBadRequest, Body: io.NopCloser(strings.NewReader("missing users"))}, nil
  }
  atomic.AddInt32(&d.got, 1)
  return &http.Response{
    StatusCode: http.StatusOK,
    Header:     http.Header{"Content-Type": []string{"application/json"}},
    Body:       io.NopCloser(strings.NewReader(`{"status":"ok"}`)),
  }, nil
}

func TestNodeSync_PushesConfigToNode(t *testing.T) {
  doer := &fakeDoer{}
  restore := api.SetNodeClientFactoryForTest(func() *node.Client {
    return node.NewClient(doer)
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
  req = httptest.NewRequest(http.MethodPost, "/api/nodes", bytes.NewReader([]byte(fmt.Sprintf(`{"name":"n1","api_address":"127.0.0.1","api_port":3000,"secret_key":"secret","public_address":"a.example.com","group_id":%d}`, g.Data.ID))))
  req.Header.Set("Authorization", "Bearer "+token)
  r.ServeHTTP(w, req)
  require.Equal(t, http.StatusCreated, w.Code)
  var n nodeResp
  require.NoError(t, json.Unmarshal(w.Body.Bytes(), &n))

  // Create inbound.
  w = httptest.NewRecorder()
  req = httptest.NewRequest(http.MethodPost, "/api/inbounds", strings.NewReader(fmt.Sprintf(`{"node_id":%d,"tag":"vless-in","protocol":"vless","listen_port":443,"public_port":443,"settings":{}}`, n.Data.ID)))
  req.Header.Set("Authorization", "Bearer "+token)
  r.ServeHTTP(w, req)
  require.Equal(t, http.StatusCreated, w.Code)

  // Create user and bind to group.
  w = httptest.NewRecorder()
  req = httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(`{"username":"alice"}`))
  req.Header.Set("Authorization", "Bearer "+token)
  r.ServeHTTP(w, req)
  require.Equal(t, http.StatusCreated, w.Code)
  var u userResp
  require.NoError(t, json.Unmarshal(w.Body.Bytes(), &u))

  w = httptest.NewRecorder()
  req = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/users/%d/groups", u.Data.ID), strings.NewReader(fmt.Sprintf(`{"group_ids":[%d]}`, g.Data.ID)))
  req.Header.Set("Authorization", "Bearer "+token)
  r.ServeHTTP(w, req)
  require.Equal(t, http.StatusOK, w.Code)

  // Sync.
  w = httptest.NewRecorder()
  req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/nodes/%d/sync", n.Data.ID), nil)
  req.Header.Set("Authorization", "Bearer "+token)
  r.ServeHTTP(w, req)
  require.Equal(t, http.StatusOK, w.Code)
  // One sync is triggered automatically when creating inbounds, and another is triggered here.
  require.Equal(t, int32(2), atomic.LoadInt32(&doer.got))
}
