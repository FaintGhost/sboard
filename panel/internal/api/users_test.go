package api_test

import (
  "bytes"
  "encoding/json"
  "fmt"
  "io"
  "net/http"
  "net/http/httptest"
  "path/filepath"
  "runtime"
  "strings"
  "sync/atomic"
  "testing"
  "time"

  "sboard/panel/internal/api"
  "sboard/panel/internal/config"
  "sboard/panel/internal/db"
  "sboard/panel/internal/node"
  "github.com/golang-jwt/jwt/v5"
  "github.com/stretchr/testify/require"
)

type usersAPIFakeDoer struct {
  got int32
}

func (d *usersAPIFakeDoer) Do(req *http.Request) (*http.Response, error) {
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
  atomic.AddInt32(&d.got, 1)
  return &http.Response{
    StatusCode: http.StatusOK,
    Header:     http.Header{"Content-Type": []string{"application/json"}},
    Body:       io.NopCloser(strings.NewReader(`{"status":"ok"}`)),
  }, nil
}

type userDTO struct {
  ID              int64      `json:"id"`
  UUID            string     `json:"uuid"`
  Username        string     `json:"username"`
  TrafficLimit    int64      `json:"traffic_limit"`
  TrafficUsed     int64      `json:"traffic_used"`
  TrafficResetDay int        `json:"traffic_reset_day"`
  ExpireAt        *time.Time `json:"expire_at"`
  Status          string     `json:"status"`
}

type userResp struct {
  Data userDTO `json:"data"`
}

type listResp struct {
  Data []userDTO `json:"data"`
}

func setupStore(t *testing.T) *db.Store {
  t.Helper()
  dir := t.TempDir()
  dbPath := filepath.Join(dir, "test.db")
  database, err := db.Open(dbPath)
  require.NoError(t, err)
  t.Cleanup(func() { _ = database.Close() })

  _, file, _, ok := runtime.Caller(0)
  require.True(t, ok)
  migrationsDir := filepath.Join(filepath.Dir(file), "..", "db", "migrations")
  err = db.MigrateUp(database, migrationsDir)
  require.NoError(t, err)

  return db.NewStore(database)
}

func mustToken(secret string) string {
  now := time.Now()
  exp := now.Add(24 * time.Hour)
  claims := jwt.RegisteredClaims{
    Subject:   "admin",
    IssuedAt:  jwt.NewNumericDate(now),
    ExpiresAt: jwt.NewNumericDate(exp),
  }
  token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
  signed, err := token.SignedString([]byte(secret))
  if err != nil {
    panic(err)
  }
  return signed
}

func TestUsersAPI_CreateListAndDisable(t *testing.T) {
  cfg := config.Config{JWTSecret: "secret"}
  store := setupStore(t)
  r := api.NewRouter(cfg, store)
  token := mustToken(cfg.JWTSecret)

  w := httptest.NewRecorder()
  req := httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(`{"username":"alice"}`))
  req.Header.Set("Authorization", "Bearer "+token)
  r.ServeHTTP(w, req)
  require.Equal(t, http.StatusCreated, w.Code)

  var created userResp
  require.NoError(t, json.Unmarshal(w.Body.Bytes(), &created))
  require.Equal(t, "alice", created.Data.Username)
  require.Equal(t, "active", created.Data.Status)

  w = httptest.NewRecorder()
  req = httptest.NewRequest(http.MethodGet, "/api/users?limit=10&offset=0", nil)
  req.Header.Set("Authorization", "Bearer "+token)
  r.ServeHTTP(w, req)
  require.Equal(t, http.StatusOK, w.Code)

  var listed listResp
  require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listed))
  require.Len(t, listed.Data, 1)

  w = httptest.NewRecorder()
  req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/users/%d", created.Data.ID), nil)
  req.Header.Set("Authorization", "Bearer "+token)
  r.ServeHTTP(w, req)
  require.Equal(t, http.StatusOK, w.Code)

  var disabled userResp
  require.NoError(t, json.Unmarshal(w.Body.Bytes(), &disabled))
  require.Equal(t, "disabled", disabled.Data.Status)

  w = httptest.NewRecorder()
  req = httptest.NewRequest(http.MethodGet, "/api/users?limit=10&offset=0", nil)
  req.Header.Set("Authorization", "Bearer "+token)
  r.ServeHTTP(w, req)
  require.Equal(t, http.StatusOK, w.Code)

  require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listed))
  require.Len(t, listed.Data, 1)
  require.Equal(t, "disabled", listed.Data[0].Status)

  w = httptest.NewRecorder()
  req = httptest.NewRequest(http.MethodGet, "/api/users?status=disabled&limit=10&offset=0", nil)
  req.Header.Set("Authorization", "Bearer "+token)
  r.ServeHTTP(w, req)
  require.Equal(t, http.StatusOK, w.Code)

  require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listed))
  require.Len(t, listed.Data, 1)
}

func TestUsersAPI_GetAndUpdate(t *testing.T) {
  cfg := config.Config{JWTSecret: "secret"}
  store := setupStore(t)
  r := api.NewRouter(cfg, store)
  token := mustToken(cfg.JWTSecret)

  w := httptest.NewRecorder()
  req := httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(`{"username":"alice"}`))
  req.Header.Set("Authorization", "Bearer "+token)
  r.ServeHTTP(w, req)
  require.Equal(t, http.StatusCreated, w.Code)

  var created userResp
  require.NoError(t, json.Unmarshal(w.Body.Bytes(), &created))

  w = httptest.NewRecorder()
  req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/users/%d", created.Data.ID), nil)
  req.Header.Set("Authorization", "Bearer "+token)
  r.ServeHTTP(w, req)
  require.Equal(t, http.StatusOK, w.Code)

  w = httptest.NewRecorder()
  req = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/users/%d", created.Data.ID), strings.NewReader(`{"traffic_limit":1024,"traffic_reset_day":1,"status":"expired"}`))
  req.Header.Set("Authorization", "Bearer "+token)
  r.ServeHTTP(w, req)
  require.Equal(t, http.StatusOK, w.Code)

  var updated userResp
  require.NoError(t, json.Unmarshal(w.Body.Bytes(), &updated))
  require.Equal(t, int64(1024), updated.Data.TrafficLimit)
  require.Equal(t, 1, updated.Data.TrafficResetDay)
  require.Equal(t, "expired", updated.Data.Status)
}

func TestUsersAPI_EffectiveStatusAndFilters(t *testing.T) {
  cfg := config.Config{JWTSecret: "secret"}
  store := setupStore(t)
  r := api.NewRouter(cfg, store)
  token := mustToken(cfg.JWTSecret)

  activeUser, err := store.CreateUser(t.Context(), "active-user")
  require.NoError(t, err)

  expiredUser, err := store.CreateUser(t.Context(), "expired-user")
  require.NoError(t, err)
  expiredAt := time.Now().UTC().Add(-1 * time.Hour).Format(time.RFC3339)
  _, err = store.DB.Exec("UPDATE users SET expire_at = ? WHERE id = ?", expiredAt, expiredUser.ID)
  require.NoError(t, err)

  exceededUser, err := store.CreateUser(t.Context(), "exceeded-user")
  require.NoError(t, err)
  _, err = store.DB.Exec(
    "UPDATE users SET traffic_limit = ?, traffic_used = ? WHERE id = ?",
    int64(1024),
    int64(1024),
    exceededUser.ID,
  )
  require.NoError(t, err)

  disabledUser, err := store.CreateUser(t.Context(), "disabled-user")
  require.NoError(t, err)
  require.NoError(t, store.DisableUser(t.Context(), disabledUser.ID))

  // Verify effective status rendering in list.
  w := httptest.NewRecorder()
  req := httptest.NewRequest(http.MethodGet, "/api/users?limit=20&offset=0", nil)
  req.Header.Set("Authorization", "Bearer "+token)
  r.ServeHTTP(w, req)
  require.Equal(t, http.StatusOK, w.Code)

  var listed listResp
  require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listed))

  statusByName := map[string]string{}
  for _, item := range listed.Data {
    statusByName[item.Username] = item.Status
  }

  require.Equal(t, "active", statusByName[activeUser.Username])
  require.Equal(t, "expired", statusByName[expiredUser.Username])
  require.Equal(t, "traffic_exceeded", statusByName[exceededUser.Username])
  require.Equal(t, "disabled", statusByName[disabledUser.Username])

  // Verify status filters.
  assertSingleStatusFilteredUser := func(status string, expectedUsername string) {
    w := httptest.NewRecorder()
    req := httptest.NewRequest(http.MethodGet, "/api/users?limit=20&offset=0&status="+status, nil)
    req.Header.Set("Authorization", "Bearer "+token)
    r.ServeHTTP(w, req)
    require.Equal(t, http.StatusOK, w.Code)

    var resp listResp
    require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
    require.Len(t, resp.Data, 1)
    require.Equal(t, expectedUsername, resp.Data[0].Username)
    require.Equal(t, status, resp.Data[0].Status)
  }

  assertSingleStatusFilteredUser("active", activeUser.Username)
  assertSingleStatusFilteredUser("expired", expiredUser.Username)
  assertSingleStatusFilteredUser("traffic_exceeded", exceededUser.Username)
  assertSingleStatusFilteredUser("disabled", disabledUser.Username)
}

func TestUsersAPI_UpdateAndDisable_AutoSyncsNodesByUserGroups(t *testing.T) {
  doer := &usersAPIFakeDoer{}
  restore := api.SetNodeClientFactoryForTest(func() *node.Client {
    return node.NewClient(doer)
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
  req = httptest.NewRequest(
    http.MethodPost,
    "/api/nodes",
    bytes.NewReader([]byte(fmt.Sprintf(`{"name":"n1","api_address":"127.0.0.1","api_port":3000,"secret_key":"secret","public_address":"a.example.com","group_id":%d}`, g.Data.ID))),
  )
  req.Header.Set("Authorization", "Bearer "+token)
  r.ServeHTTP(w, req)
  require.Equal(t, http.StatusCreated, w.Code)
  var n nodeResp
  require.NoError(t, json.Unmarshal(w.Body.Bytes(), &n))

  w = httptest.NewRecorder()
  req = httptest.NewRequest(
    http.MethodPost,
    "/api/inbounds",
    strings.NewReader(fmt.Sprintf(`{"node_id":%d,"tag":"ss-in","protocol":"shadowsocks","listen_port":8388,"public_port":8388,"settings":{"method":"2022-blake3-aes-128-gcm"}}`, n.Data.ID)),
  )
  req.Header.Set("Authorization", "Bearer "+token)
  r.ServeHTTP(w, req)
  require.Equal(t, http.StatusCreated, w.Code)

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

  syncCallsAfterGroupBind := atomic.LoadInt32(&doer.got)
  require.GreaterOrEqual(t, syncCallsAfterGroupBind, int32(1))

  w = httptest.NewRecorder()
  req = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/users/%d", u.Data.ID), strings.NewReader(`{"traffic_limit":1024}`))
  req.Header.Set("Authorization", "Bearer "+token)
  r.ServeHTTP(w, req)
  require.Equal(t, http.StatusOK, w.Code)
  require.Equal(t, syncCallsAfterGroupBind+1, atomic.LoadInt32(&doer.got))

  w = httptest.NewRecorder()
  req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/users/%d", u.Data.ID), nil)
  req.Header.Set("Authorization", "Bearer "+token)
  r.ServeHTTP(w, req)
  require.Equal(t, http.StatusOK, w.Code)
  require.Equal(t, syncCallsAfterGroupBind+2, atomic.LoadInt32(&doer.got))
}
