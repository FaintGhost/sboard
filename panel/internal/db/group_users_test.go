package db_test

import (
  "context"
  "testing"

  "github.com/stretchr/testify/require"
)

func TestListActiveUsersForGroupSkipsTrafficExceeded(t *testing.T) {
  store := setupStore(t)
  ctx := context.Background()

  groupID := insertGroup(t, store, "vip")

  okUser, err := store.CreateUser(ctx, "ok-user")
  require.NoError(t, err)
  exceededUser, err := store.CreateUser(ctx, "exceeded-user")
  require.NoError(t, err)

  _, err = store.DB.Exec("INSERT INTO user_groups (user_id, group_id) VALUES (?, ?)", okUser.ID, groupID)
  require.NoError(t, err)
  _, err = store.DB.Exec("INSERT INTO user_groups (user_id, group_id) VALUES (?, ?)", exceededUser.ID, groupID)
  require.NoError(t, err)

  _, err = store.DB.Exec(
    "UPDATE users SET traffic_limit = ?, traffic_used = ? WHERE id = ?",
    int64(1024),
    int64(1024),
    exceededUser.ID,
  )
  require.NoError(t, err)

  users, err := store.ListActiveUsersForGroup(ctx, groupID)
  require.NoError(t, err)
  require.Len(t, users, 1)
  require.Equal(t, okUser.ID, users[0].ID)
}

