package api_test

import (
  "encoding/json"
  "fmt"
  "net/http"
  "net/http/httptest"
  "strings"
  "testing"

  "sboard/panel/internal/api"
  "sboard/panel/internal/config"
  "github.com/stretchr/testify/require"
)

type groupDTO struct {
  ID          int64  `json:"id"`
  Name        string `json:"name"`
  Description string `json:"description"`
}

type groupResp struct {
  Data groupDTO `json:"data"`
}

type groupsListResp struct {
  Data []groupDTO `json:"data"`
}

type userGroupsDTO struct {
  GroupIDs []int64 `json:"group_ids"`
}

type userGroupsResp struct {
  Data userGroupsDTO `json:"data"`
}

func TestGroupsAPI_CRUDAndUserBind(t *testing.T) {
  cfg := config.Config{JWTSecret: "secret"}
  store := setupStore(t)
  r := api.NewRouter(cfg, store)
  token := mustToken(cfg.JWTSecret)

  w := httptest.NewRecorder()
  req := httptest.NewRequest(http.MethodPost, "/api/groups", strings.NewReader(`{"name":"g1","description":"d"}`))
  req.Header.Set("Authorization", "Bearer "+token)
  r.ServeHTTP(w, req)
  require.Equal(t, http.StatusCreated, w.Code)

  var created groupResp
  require.NoError(t, json.Unmarshal(w.Body.Bytes(), &created))
  require.Equal(t, "g1", created.Data.Name)

  w = httptest.NewRecorder()
  req = httptest.NewRequest(http.MethodGet, "/api/groups?limit=10&offset=0", nil)
  req.Header.Set("Authorization", "Bearer "+token)
  r.ServeHTTP(w, req)
  require.Equal(t, http.StatusOK, w.Code)

  var listed groupsListResp
  require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listed))
  require.Len(t, listed.Data, 1)

  // Create user and bind to group.
  w = httptest.NewRecorder()
  req = httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(`{"username":"alice"}`))
  req.Header.Set("Authorization", "Bearer "+token)
  r.ServeHTTP(w, req)
  require.Equal(t, http.StatusCreated, w.Code)

  var u userResp
  require.NoError(t, json.Unmarshal(w.Body.Bytes(), &u))

  w = httptest.NewRecorder()
  req = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/users/%d/groups", u.Data.ID), strings.NewReader(fmt.Sprintf(`{"group_ids":[%d]}`, created.Data.ID)))
  req.Header.Set("Authorization", "Bearer "+token)
  r.ServeHTTP(w, req)
  require.Equal(t, http.StatusOK, w.Code)

  var bound userGroupsResp
  require.NoError(t, json.Unmarshal(w.Body.Bytes(), &bound))
  require.Equal(t, []int64{created.Data.ID}, bound.Data.GroupIDs)

  w = httptest.NewRecorder()
  req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/users/%d/groups", u.Data.ID), nil)
  req.Header.Set("Authorization", "Bearer "+token)
  r.ServeHTTP(w, req)
  require.Equal(t, http.StatusOK, w.Code)
  require.NoError(t, json.Unmarshal(w.Body.Bytes(), &bound))
  require.Equal(t, []int64{created.Data.ID}, bound.Data.GroupIDs)
}

