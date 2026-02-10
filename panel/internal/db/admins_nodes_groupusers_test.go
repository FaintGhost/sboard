package db_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"sboard/panel/internal/db"
)

func TestAdminLifecycle(t *testing.T) {
	store := setupStore(t)

	_, err := db.AdminCount(nil)
	require.ErrorContains(t, err, "nil store")

	created, err := db.AdminCreateIfNone(store, " admin ", "hash-1")
	require.NoError(t, err)
	require.True(t, created)

	created, err = db.AdminCreateIfNone(store, "admin2", "hash-2")
	require.NoError(t, err)
	require.False(t, created)

	count, err := db.AdminCount(store)
	require.NoError(t, err)
	require.Equal(t, 1, count)

	_, ok, err := db.AdminGetByID(store, 0)
	require.NoError(t, err)
	require.False(t, ok)

	admin, ok, err := db.AdminGetByUsername(store, "admin")
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, "admin", admin.Username)

	first, ok, err := db.AdminGetFirst(store)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, admin.ID, first.ID)

	err = db.AdminUpdateCredentials(store, admin.ID, "", "hash")
	require.ErrorContains(t, err, "missing username")
	err = db.AdminUpdateCredentials(store, admin.ID, "admin", "")
	require.ErrorContains(t, err, "missing password_hash")
	err = db.AdminUpdateCredentials(store, 999999, "admin", "hash")
	require.ErrorIs(t, err, db.ErrNotFound)

	err = db.AdminUpdateCredentials(store, admin.ID, "new-admin", "hash-new")
	require.NoError(t, err)

	updated, ok, err := db.AdminGetByID(store, admin.ID)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, "new-admin", updated.Username)
	require.Equal(t, "hash-new", updated.PasswordHash)

	_, err = store.DB.Exec("INSERT INTO admins (username, password_hash) VALUES (?, ?)", "taken", "hash")
	require.NoError(t, err)
	err = db.AdminUpdateCredentials(store, admin.ID, "taken", "other")
	require.True(t, errors.Is(err, db.ErrConflict))
}

func TestGroupsAndGroupUsers(t *testing.T) {
	store := setupStore(t)
	ctx := context.Background()

	group1, err := store.CreateGroup(ctx, "g1", "desc1")
	require.NoError(t, err)
	group2, err := store.CreateGroup(ctx, "g2", "desc2")
	require.NoError(t, err)

	user1, err := store.CreateUser(ctx, "u1")
	require.NoError(t, err)
	user2, err := store.CreateUser(ctx, "u2")
	require.NoError(t, err)

	err = store.ReplaceGroupUsers(ctx, group1.ID, []int64{user2.ID, user1.ID, user2.ID})
	require.NoError(t, err)

	users, err := store.ListGroupUsers(ctx, group1.ID)
	require.NoError(t, err)
	require.Len(t, users, 2)
	require.Equal(t, user1.ID, users[0].ID)
	require.Equal(t, user2.ID, users[1].ID)

	err = store.ReplaceGroupUsers(ctx, group1.ID, []int64{0})
	require.ErrorContains(t, err, "invalid user id")
	err = store.ReplaceGroupUsers(ctx, group1.ID, []int64{999999})
	require.ErrorIs(t, err, db.ErrNotFound)
	err = store.ReplaceGroupUsers(ctx, 999999, []int64{user1.ID})
	require.ErrorIs(t, err, db.ErrNotFound)

	_, err = store.ListGroupUsers(ctx, 0)
	require.ErrorContains(t, err, "invalid group id")
	_, err = store.ListGroupUsers(ctx, 999999)
	require.ErrorIs(t, err, db.ErrNotFound)

	groups, err := store.ListGroups(ctx, 10, 0)
	require.NoError(t, err)
	require.Len(t, groups, 2)
	require.Equal(t, group2.ID, groups[0].ID)
	require.Equal(t, int64(0), groups[0].MemberCount)
	require.Equal(t, group1.ID, groups[1].ID)
	require.Equal(t, int64(2), groups[1].MemberCount)

	newName := "g1-new"
	newDesc := "desc-new"
	updated, err := store.UpdateGroup(ctx, group1.ID, db.GroupUpdate{Name: &newName, Description: &newDesc})
	require.NoError(t, err)
	require.Equal(t, newName, updated.Name)
	require.Equal(t, newDesc, updated.Description)

	unchanged, err := store.UpdateGroup(ctx, group1.ID, db.GroupUpdate{})
	require.NoError(t, err)
	require.Equal(t, updated.ID, unchanged.ID)

	_, err = store.UpdateGroup(ctx, 999999, db.GroupUpdate{Name: &newName})
	require.ErrorIs(t, err, db.ErrNotFound)
}

func TestUserGroupsBatchAndUserDelete(t *testing.T) {
	store := setupStore(t)
	ctx := context.Background()

	u1, err := store.CreateUser(ctx, "u1-batch")
	require.NoError(t, err)
	u2, err := store.CreateUser(ctx, "u2-batch")
	require.NoError(t, err)
	u3, err := store.CreateUser(ctx, "u3-batch")
	require.NoError(t, err)

	g1, err := store.CreateGroup(ctx, "g1-batch", "")
	require.NoError(t, err)
	g2, err := store.CreateGroup(ctx, "g2-batch", "")
	require.NoError(t, err)

	err = store.ReplaceUserGroups(ctx, u1.ID, []int64{g2.ID, g1.ID})
	require.NoError(t, err)
	err = store.ReplaceUserGroups(ctx, u2.ID, []int64{g2.ID})
	require.NoError(t, err)

	batch, err := store.ListUserGroupIDsBatch(ctx, []int64{u1.ID, u2.ID, u3.ID})
	require.NoError(t, err)
	require.Equal(t, []int64{g1.ID, g2.ID}, batch[u1.ID])
	require.Equal(t, []int64{g2.ID}, batch[u2.ID])
	require.Empty(t, batch[u3.ID])

	empty, err := store.ListUserGroupIDsBatch(ctx, nil)
	require.NoError(t, err)
	require.Empty(t, empty)

	gotByUUID, err := store.GetUserByUUID(ctx, u1.UUID)
	require.NoError(t, err)
	require.Equal(t, u1.ID, gotByUUID.ID)

	err = store.DeleteUser(ctx, u1.ID)
	require.NoError(t, err)
	_, err = store.GetUserByUUID(ctx, u1.UUID)
	require.ErrorIs(t, err, db.ErrNotFound)
	err = store.DeleteUser(ctx, u1.ID)
	require.ErrorIs(t, err, db.ErrNotFound)
}

func TestNodeListUpdateAndStatus(t *testing.T) {
	store := setupStore(t)
	ctx := context.Background()
	fixedNow := time.Date(2026, 2, 10, 13, 30, 0, 0, time.UTC)
	store.Now = func() time.Time { return fixedNow }

	group, err := store.CreateGroup(ctx, "node-g", "")
	require.NoError(t, err)

	node, err := store.CreateNode(ctx, db.NodeCreate{
		Name:          "node-1",
		APIAddress:    "10.0.0.1",
		APIPort:       3000,
		SecretKey:     "secret-1",
		PublicAddress: "public-1",
		GroupID:       &group.ID,
	})
	require.NoError(t, err)

	nodes, err := store.ListNodes(ctx, 10, 0)
	require.NoError(t, err)
	require.Len(t, nodes, 1)
	require.Equal(t, node.ID, nodes[0].ID)

	newName := "node-1-new"
	newAddr := "10.0.0.2"
	newPort := 3100
	newSecret := "secret-2"
	newPublic := "public-2"
	updated, err := store.UpdateNode(ctx, node.ID, db.NodeUpdate{
		Name:          &newName,
		APIAddress:    &newAddr,
		APIPort:       &newPort,
		SecretKey:     &newSecret,
		PublicAddress: &newPublic,
		GroupIDSet:    true,
	})
	require.NoError(t, err)
	require.Equal(t, newName, updated.Name)
	require.Equal(t, newAddr, updated.APIAddress)
	require.Equal(t, newPort, updated.APIPort)
	require.Equal(t, newSecret, updated.SecretKey)
	require.Equal(t, newPublic, updated.PublicAddress)
	require.Nil(t, updated.GroupID)

	noChange, err := store.UpdateNode(ctx, node.ID, db.NodeUpdate{})
	require.NoError(t, err)
	require.Equal(t, updated.ID, noChange.ID)

	err = store.MarkNodeOnline(ctx, 0, time.Time{})
	require.ErrorContains(t, err, "invalid id")
	err = store.MarkNodeOnline(ctx, 999999, fixedNow)
	require.ErrorIs(t, err, db.ErrNotFound)
	err = store.MarkNodeOnline(ctx, node.ID, time.Time{})
	require.NoError(t, err)

	online, err := store.GetNodeByID(ctx, node.ID)
	require.NoError(t, err)
	require.Equal(t, "online", online.Status)
	require.NotNil(t, online.LastSeenAt)
	require.WithinDuration(t, fixedNow, *online.LastSeenAt, time.Second)

	err = store.MarkNodeOffline(ctx, 0)
	require.ErrorContains(t, err, "invalid id")
	err = store.MarkNodeOffline(ctx, 999999)
	require.ErrorIs(t, err, db.ErrNotFound)
	err = store.MarkNodeOffline(ctx, node.ID)
	require.NoError(t, err)

	offline, err := store.GetNodeByID(ctx, node.ID)
	require.NoError(t, err)
	require.Equal(t, "offline", offline.Status)

	_, err = store.UpdateNode(ctx, 999999, db.NodeUpdate{Name: &newName})
	require.ErrorIs(t, err, db.ErrNotFound)
}
