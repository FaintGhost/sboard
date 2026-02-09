package db_test

import (
  "context"
  "errors"
  "testing"

  "sboard/panel/internal/db"
  "github.com/stretchr/testify/require"
)

func TestGroupCreateAndUnique(t *testing.T) {
  store := setupStore(t)
  g, err := store.CreateGroup(context.Background(), "g1", "desc")
  require.NoError(t, err)
  require.NotZero(t, g.ID)
  require.Equal(t, "g1", g.Name)

  _, err = store.CreateGroup(context.Background(), "g1", "")
  require.Error(t, err)
  require.True(t, errors.Is(err, db.ErrConflict))
}

func TestReplaceUserGroups(t *testing.T) {
  store := setupStore(t)
  ctx := context.Background()

  u, err := store.CreateUser(ctx, "alice")
  require.NoError(t, err)
  g1, err := store.CreateGroup(ctx, "g1", "")
  require.NoError(t, err)
  g2, err := store.CreateGroup(ctx, "g2", "")
  require.NoError(t, err)

  require.NoError(t, store.ReplaceUserGroups(ctx, u.ID, []int64{g2.ID, g1.ID, g2.ID}))
  ids, err := store.ListUserGroupIDs(ctx, u.ID)
  require.NoError(t, err)
  require.Equal(t, []int64{g1.ID, g2.ID}, ids)

  require.NoError(t, store.ReplaceUserGroups(ctx, u.ID, []int64{}))
  ids, err = store.ListUserGroupIDs(ctx, u.ID)
  require.NoError(t, err)
  require.Empty(t, ids)

  err = store.ReplaceUserGroups(ctx, u.ID, []int64{999999})
  require.Error(t, err)
  require.True(t, errors.Is(err, db.ErrNotFound))
}

func TestDeleteGroup_CleansUserGroupMembership(t *testing.T) {
  store := setupStore(t)
  ctx := context.Background()

  user, err := store.CreateUser(ctx, "alice-delete-group")
  require.NoError(t, err)

  group, err := store.CreateGroup(ctx, "g-delete", "")
  require.NoError(t, err)

  require.NoError(t, store.ReplaceUserGroups(ctx, user.ID, []int64{group.ID}))

  ids, err := store.ListUserGroupIDs(ctx, user.ID)
  require.NoError(t, err)
  require.Equal(t, []int64{group.ID}, ids)

  require.NoError(t, store.DeleteGroup(ctx, group.ID))

  ids, err = store.ListUserGroupIDs(ctx, user.ID)
  require.NoError(t, err)
  require.Empty(t, ids)
}
