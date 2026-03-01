package monitor

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"sboard/panel/internal/db"
)

// mockNodeClient records which node IDs received Health calls.
type mockNodeClient struct {
	healthCalls []int64 // records node IDs that received Health calls
	healthErr   error   // configurable error to return from Health
}

func (m *mockNodeClient) Health(ctx context.Context, node db.Node) error {
	m.healthCalls = append(m.healthCalls, node.ID)
	return m.healthErr
}

func (m *mockNodeClient) SyncConfig(ctx context.Context, node db.Node, payload any) error {
	return nil
}

func TestNodesMonitor_SkipsPendingNodes(t *testing.T) {
	ctx := context.Background()
	store := setupStore(t)

	// Create an offline node (normal, fully configured).
	nodeA, err := store.CreateNode(ctx, db.NodeCreate{
		Name:          "Node-A",
		APIAddress:    "127.0.0.1",
		APIPort:       3001,
		SecretKey:     "key-a",
		PublicAddress: "a.example.com",
	})
	require.NoError(t, err)
	require.Equal(t, "offline", nodeA.Status)

	// Create a pending node (via heartbeat registration, no name/group).
	nodeB, err := store.CreatePendingNode(ctx, db.PendingNodeParams{
		UUID:       "pending-uuid-001",
		APIAddress: "10.0.0.1",
		APIPort:    3002,
		SecretKey:  "key-b",
	})
	require.NoError(t, err)
	require.Equal(t, "pending", nodeB.Status)

	client := &mockNodeClient{}
	m := NewNodesMonitor(store, client)

	require.NoError(t, m.CheckOnce(ctx))

	// Only Node-A should have been polled.
	require.Equal(t, []int64{nodeA.ID}, client.healthCalls,
		"CheckOnce should only call Health on non-pending nodes")

	// Node-B should remain pending.
	gotB, err := store.GetNodeByID(ctx, nodeB.ID)
	require.NoError(t, err)
	require.Equal(t, "pending", gotB.Status,
		"pending node status must not be changed by CheckOnce")
}

func TestNodesMonitor_ApprovedNodeGetsPolled(t *testing.T) {
	ctx := context.Background()
	store := setupStore(t)

	// Start as a pending node.
	pending, err := store.CreatePendingNode(ctx, db.PendingNodeParams{
		UUID:       "approved-uuid-001",
		APIAddress: "10.0.0.2",
		APIPort:    3003,
		SecretKey:  "key-c",
	})
	require.NoError(t, err)
	require.Equal(t, "pending", pending.Status)

	// Approve the node — this transitions it to "offline".
	approved, err := store.ApproveNode(ctx, pending.ID, db.ApproveNodeParams{
		Name:          "Node-C",
		PublicAddress: "c.example.com",
	})
	require.NoError(t, err)
	require.Equal(t, "offline", approved.Status)

	client := &mockNodeClient{}
	m := NewNodesMonitor(store, client)

	require.NoError(t, m.CheckOnce(ctx))

	// The approved (now offline) node should be polled.
	require.Equal(t, []int64{approved.ID}, client.healthCalls,
		"approved node should be polled by CheckOnce")
}
