package api_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"sboard/panel/internal/api"
	"sboard/panel/internal/config"
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

type groupUsersListResp struct {
	Data []userDTO `json:"data"`
}

type groupUsersPutDTO struct {
	UserIDs []int64 `json:"user_ids"`
}

type groupUsersPutResp struct {
	Data groupUsersPutDTO `json:"data"`
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

func TestGroupUsersAPI_ListAndReplace(t *testing.T) {
	cfg := config.Config{JWTSecret: "secret"}
	store := setupStore(t)
	r := api.NewRouter(cfg, store)
	token := mustToken(cfg.JWTSecret)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/groups", strings.NewReader(`{"name":"g-members","description":""}`))
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var createdGroup groupResp
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &createdGroup))

	createUser := func(name string) userResp {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(fmt.Sprintf(`{"username":"%s"}`, name)))
		req.Header.Set("Authorization", "Bearer "+token)
		r.ServeHTTP(w, req)
		require.Equal(t, http.StatusCreated, w.Code)
		var u userResp
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &u))
		return u
	}

	alice := createUser("alice-members")
	bob := createUser("bob-members")
	carol := createUser("carol-members")

	// 先通过现有 user -> groups 绑定，验证 groups -> users 能正确读取。
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/users/%d/groups", alice.Data.ID), strings.NewReader(fmt.Sprintf(`{"group_ids":[%d]}`, createdGroup.Data.ID)))
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/groups/%d/users", createdGroup.Data.ID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var listed groupUsersListResp
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listed))
	require.Len(t, listed.Data, 1)
	require.Equal(t, alice.Data.ID, listed.Data[0].ID)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/groups/%d/users", createdGroup.Data.ID), strings.NewReader(fmt.Sprintf(`{"user_ids":[%d,%d]}`, bob.Data.ID, carol.Data.ID)))
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var replaced groupUsersPutResp
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &replaced))
	require.ElementsMatch(t, []int64{bob.Data.ID, carol.Data.ID}, replaced.Data.UserIDs)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/groups/%d/users", createdGroup.Data.ID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listed))
	require.Len(t, listed.Data, 2)
	gotIDs := []int64{listed.Data[0].ID, listed.Data[1].ID}
	sort.Slice(gotIDs, func(i, j int) bool { return gotIDs[i] < gotIDs[j] })
	require.Equal(t, []int64{bob.Data.ID, carol.Data.ID}, gotIDs)

	// alice 应已被移出该分组。
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/users/%d/groups", alice.Data.ID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	var userGroups userGroupsResp
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &userGroups))
	require.Len(t, userGroups.Data.GroupIDs, 0)
}

func TestGroupsAPI_DeleteGroupCleansUserMembership(t *testing.T) {
	cfg := config.Config{JWTSecret: "secret"}
	store := setupStore(t)
	r := api.NewRouter(cfg, store)
	token := mustToken(cfg.JWTSecret)

	// create group
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/groups", strings.NewReader(`{"name":"g-clean","description":""}`))
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var group groupResp
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &group))

	// create user
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(`{"username":"alice-clean"}`))
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var user userResp
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &user))

	// bind user to group
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/users/%d/groups", user.Data.ID), strings.NewReader(fmt.Sprintf(`{"group_ids":[%d]}`, group.Data.ID)))
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	// delete group
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/groups/%d", group.Data.ID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	// user groups should be empty
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/users/%d/groups", user.Data.ID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var groupsResp userGroupsResp
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &groupsResp))
	require.Empty(t, groupsResp.Data.GroupIDs)
}
