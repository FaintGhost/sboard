package db_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"sboard/panel/internal/db"
)

func TestGetNodeByUUID_Found(t *testing.T) {
	store := setupStore(t)
	ctx := context.Background()

	created, err := store.CreateNode(ctx, db.NodeCreate{
		Name:       "test-node",
		APIAddress: "10.0.0.1",
		APIPort:    3000,
		SecretKey:  "secret-1",
	})
	require.NoError(t, err)

	found, err := store.GetNodeByUUID(ctx, created.UUID)
	require.NoError(t, err)
	require.Equal(t, created.ID, found.ID)
	require.Equal(t, created.UUID, found.UUID)
	require.Equal(t, "test-node", found.Name)
}

func TestGetNodeByUUID_NotFound(t *testing.T) {
	store := setupStore(t)
	ctx := context.Background()

	_, err := store.GetNodeByUUID(ctx, "non-existent-uuid")
	require.ErrorIs(t, err, db.ErrNotFound)
}

func TestCreatePendingNode(t *testing.T) {
	store := setupStore(t)
	ctx := context.Background()

	node, err := store.CreatePendingNode(ctx, db.PendingNodeParams{
		UUID:       "node-uuid-1",
		SecretKey:  "secret-key-1",
		APIAddress: "10.0.0.5",
		APIPort:    8080,
	})
	require.NoError(t, err)
	require.NotZero(t, node.ID)
	require.Equal(t, "node-uuid-1", node.UUID)
	require.Equal(t, "", node.Name)
	require.Equal(t, "pending", node.Status)
	require.Equal(t, "secret-key-1", node.SecretKey)
	require.Equal(t, "10.0.0.5", node.APIAddress)
	require.Equal(t, 8080, node.APIPort)
}

func TestCreatePendingNode_DuplicateUUID(t *testing.T) {
	store := setupStore(t)
	ctx := context.Background()
	fixedNow := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)
	store.Now = func() time.Time { return fixedNow }

	first, err := store.CreatePendingNode(ctx, db.PendingNodeParams{
		UUID:       "dup-uuid",
		SecretKey:  "secret-1",
		APIAddress: "10.0.0.1",
		APIPort:    3000,
	})
	require.NoError(t, err)

	second, err := store.CreatePendingNode(ctx, db.PendingNodeParams{
		UUID:       "dup-uuid",
		SecretKey:  "secret-2",
		APIAddress: "10.0.0.2",
		APIPort:    3001,
	})
	require.NoError(t, err)
	require.Equal(t, first.ID, second.ID)
	require.NotNil(t, second.LastSeenAt)
	require.WithinDuration(t, fixedNow, *second.LastSeenAt, time.Second)
}

func TestApproveNode_Success(t *testing.T) {
	store := setupStore(t)
	ctx := context.Background()

	group, err := store.CreateGroup(ctx, "approve-group", "")
	require.NoError(t, err)

	pending, err := store.CreatePendingNode(ctx, db.PendingNodeParams{
		UUID:       "approve-uuid",
		SecretKey:  "secret-1",
		APIAddress: "10.0.0.1",
		APIPort:    3000,
	})
	require.NoError(t, err)
	require.Equal(t, "pending", pending.Status)

	approved, err := store.ApproveNode(ctx, pending.ID, db.ApproveNodeParams{
		Name:          "approved-node",
		GroupID:       &group.ID,
		PublicAddress: "public.example.com",
	})
	require.NoError(t, err)
	require.Equal(t, "offline", approved.Status)
	require.Equal(t, "approved-node", approved.Name)
	require.NotNil(t, approved.GroupID)
	require.Equal(t, group.ID, *approved.GroupID)
	require.Equal(t, "public.example.com", approved.PublicAddress)
}

func TestApproveNode_NotPending(t *testing.T) {
	store := setupStore(t)
	ctx := context.Background()

	// Create a normal (offline) node via the regular method
	node, err := store.CreateNode(ctx, db.NodeCreate{
		Name:       "offline-node",
		APIAddress: "10.0.0.1",
		APIPort:    3000,
		SecretKey:  "secret-1",
	})
	require.NoError(t, err)
	require.Equal(t, "offline", node.Status)

	_, err = store.ApproveNode(ctx, node.ID, db.ApproveNodeParams{
		Name: "should-fail",
	})
	require.Error(t, err)
	require.ErrorIs(t, err, db.ErrNotPending)
}

func TestUpdateNodeLastSeen(t *testing.T) {
	store := setupStore(t)
	ctx := context.Background()

	node, err := store.CreateNode(ctx, db.NodeCreate{
		Name:       "seen-node",
		APIAddress: "10.0.0.1",
		APIPort:    3000,
		SecretKey:  "secret-1",
	})
	require.NoError(t, err)

	seenAt := time.Date(2026, 3, 1, 15, 30, 0, 0, time.UTC)
	err = store.UpdateNodeLastSeen(ctx, node.UUID, seenAt)
	require.NoError(t, err)

	updated, err := store.GetNodeByUUID(ctx, node.UUID)
	require.NoError(t, err)
	require.NotNil(t, updated.LastSeenAt)
	require.WithinDuration(t, seenAt, *updated.LastSeenAt, time.Second)
}
