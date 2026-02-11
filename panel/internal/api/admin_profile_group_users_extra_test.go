package api_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"sboard/panel/internal/api"
	"sboard/panel/internal/config"
	"sboard/panel/internal/password"
)

func TestAdminProfile_GetPut_ErrorPaths(t *testing.T) {
	cfg := config.Config{JWTSecret: "secret", SetupToken: "setup-123"}
	store := setupStore(t)
	r := api.NewRouter(cfg, store)
	token := mustToken(cfg.JWTSecret)

	// GET before setup
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/admin/profile", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusPreconditionRequired, w.Code)

	// invalid body (bind error)
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, "/api/admin/profile", strings.NewReader(`{`))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)

	// needs setup (valid body but no admin yet)
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, "/api/admin/profile", strings.NewReader(`{"old_password":"pass12345"}`))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusPreconditionRequired, w.Code)

	// bootstrap admin
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/admin/bootstrap", strings.NewReader(`{"username":"admin","password":"pass12345","confirm_password":"pass12345","setup_token":"setup-123"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	// missing old_password
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, "/api/admin/profile", strings.NewReader(`{"new_username":"root"}`))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)

	// passwords not match
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, "/api/admin/profile", strings.NewReader(`{"old_password":"pass12345","new_password":"newpass12345","confirm_password":"different"}`))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)

	// password too short
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, "/api/admin/profile", strings.NewReader(`{"old_password":"pass12345","new_password":"short","confirm_password":"short"}`))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)

	// no changes
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, "/api/admin/profile", strings.NewReader(`{"old_password":"pass12345"}`))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)

	// create second admin directly, then update username conflict
	h, err := password.Hash("otherpass123")
	require.NoError(t, err)
	_, err = store.DB.Exec("INSERT INTO admins (username, password_hash) VALUES (?, ?)", "other-admin", h)
	require.NoError(t, err)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, "/api/admin/profile", strings.NewReader(`{"new_username":"other-admin","old_password":"pass12345"}`))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusConflict, w.Code)
}

func TestGroupUsers_ListReplace_ErrorPathsAndDedupSort(t *testing.T) {
	cfg := config.Config{JWTSecret: "secret"}
	store := setupStore(t)
	r := api.NewRouter(cfg, store)
	token := mustToken(cfg.JWTSecret)

	// list invalid id
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/groups/not-a-number/users", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)

	// list not found
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/groups/99999/users", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusNotFound, w.Code)

	// create group + users
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/groups", strings.NewReader(`{"name":"g-users-edge","description":""}`))
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)
	var g groupResp
	mustUnmarshalLocal(t, w.Body.Bytes(), &g)

	createUser := func(name string) int64 {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(fmt.Sprintf(`{"username":"%s"}`, name)))
		req.Header.Set("Authorization", "Bearer "+token)
		r.ServeHTTP(w, req)
		require.Equal(t, http.StatusCreated, w.Code)
		var u userResp
		mustUnmarshalLocal(t, w.Body.Bytes(), &u)
		return u.Data.ID
	}
	u1 := createUser("u1-edge")
	u2 := createUser("u2-edge")

	// replace invalid id
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, "/api/groups/not-a-number/users", strings.NewReader(`{"user_ids":[1]}`))
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)

	// replace invalid body
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/groups/%d/users", g.Data.ID), strings.NewReader(`{`))
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)

	// replace invalid user id
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/groups/%d/users", g.Data.ID), strings.NewReader(`{"user_ids":[0]}`))
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)

	// replace user not found
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/groups/%d/users", g.Data.ID), strings.NewReader(`{"user_ids":[99999]}`))
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusNotFound, w.Code)

	// replace group not found
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, "/api/groups/99999/users", strings.NewReader(fmt.Sprintf(`{"user_ids":[%d]}`, u1)))
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusNotFound, w.Code)

	// success: should dedup + sort response ids
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/groups/%d/users", g.Data.ID), strings.NewReader(fmt.Sprintf(`{"user_ids":[%d,%d,%d]}`, u2, u1, u2)))
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		Data struct {
			UserIDs []int64 `json:"user_ids"`
		} `json:"data"`
	}
	mustUnmarshalLocal(t, w.Body.Bytes(), &resp)
	require.Equal(t, []int64{u1, u2}, resp.Data.UserIDs)
}

func mustUnmarshalLocal(t *testing.T, b []byte, v any) {
	t.Helper()
	require.NoError(t, json.Unmarshal(b, v))
}
