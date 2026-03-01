package rpc

import (
	"context"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/require"
	"sboard/panel/internal/config"
	"sboard/panel/internal/db"
	panelv1 "sboard/panel/internal/rpc/gen/sboard/panel/v1"
	panelv1connect "sboard/panel/internal/rpc/gen/sboard/panel/v1/panelv1connect"
)

func setupApprovalTestServer(t *testing.T) (*httptest.Server, *db.Store) {
	t.Helper()
	store := setupRPCStore(t)
	cfg := config.Config{JWTSecret: testJWTSecret}
	handler := NewHandler(cfg, store)
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv, store
}

func approvalClient(t *testing.T, srv *httptest.Server) panelv1connect.NodeServiceClient {
	t.Helper()
	token := mustJWT(testJWTSecret)
	opts := connect.WithInterceptors(bearerInterceptor(token))
	return panelv1connect.NewNodeServiceClient(srv.Client(), srv.URL, opts)
}

func TestApproveNode_Success(t *testing.T) {
	srv, store := setupApprovalTestServer(t)
	ctx := context.Background()
	client := approvalClient(t, srv)

	// Create a pending node via store.
	pending, err := store.CreatePendingNode(ctx, db.PendingNodeParams{
		UUID:       "approve-success-uuid",
		SecretKey:  "secret-1",
		APIAddress: "127.0.0.1",
		APIPort:    8080,
	})
	require.NoError(t, err)
	require.Equal(t, "pending", pending.Status)

	groupID := int64(0)
	// Create a group so we can assign it.
	group, err := store.CreateGroup(ctx, "test-group", "")
	require.NoError(t, err)
	groupID = group.ID

	pubAddr := "10.0.0.1"
	resp, err := client.ApproveNode(ctx, &panelv1.ApproveNodeRequest{
		Id:            pending.ID,
		Name:          "my-node",
		GroupId:       &groupID,
		PublicAddress: &pubAddr,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Data)
	require.Equal(t, "offline", resp.Data.Status)
	require.Equal(t, "my-node", resp.Data.Name)
	require.NotNil(t, resp.Data.GroupId)
	require.Equal(t, groupID, *resp.Data.GroupId)
	require.Equal(t, "10.0.0.1", resp.Data.PublicAddress)
}

func TestApproveNode_NotPending(t *testing.T) {
	srv, store := setupApprovalTestServer(t)
	ctx := context.Background()
	client := approvalClient(t, srv)

	// Create a normal (offline) node via store.CreateNode.
	nodeItem, err := store.CreateNode(ctx, db.NodeCreate{
		Name:          "normal-node",
		APIAddress:    "127.0.0.1",
		APIPort:       9090,
		SecretKey:     "secret-2",
		PublicAddress: "10.0.0.2",
	})
	require.NoError(t, err)

	_, err = client.ApproveNode(ctx, &panelv1.ApproveNodeRequest{
		Id:   nodeItem.ID,
		Name: "renamed",
	})
	require.Error(t, err)

	var connErr *connect.Error
	require.ErrorAs(t, err, &connErr)
	require.Equal(t, connect.CodeInvalidArgument, connErr.Code())
}

func TestApproveNode_NotFound(t *testing.T) {
	srv, _ := setupApprovalTestServer(t)
	ctx := context.Background()
	client := approvalClient(t, srv)

	_, err := client.ApproveNode(ctx, &panelv1.ApproveNodeRequest{
		Id:   99999,
		Name: "ghost",
	})
	require.Error(t, err)

	var connErr *connect.Error
	require.ErrorAs(t, err, &connErr)
	require.Equal(t, connect.CodeNotFound, connErr.Code())
}

func TestApproveNode_EmptyName(t *testing.T) {
	srv, store := setupApprovalTestServer(t)
	ctx := context.Background()
	client := approvalClient(t, srv)

	pending, err := store.CreatePendingNode(ctx, db.PendingNodeParams{
		UUID:       "approve-empty-name-uuid",
		SecretKey:  "secret-3",
		APIAddress: "127.0.0.1",
		APIPort:    8081,
	})
	require.NoError(t, err)

	_, err = client.ApproveNode(ctx, &panelv1.ApproveNodeRequest{
		Id:   pending.ID,
		Name: "",
	})
	require.Error(t, err)

	var connErr *connect.Error
	require.ErrorAs(t, err, &connErr)
	require.Equal(t, connect.CodeInvalidArgument, connErr.Code())
}

func TestRejectNode_Success(t *testing.T) {
	srv, store := setupApprovalTestServer(t)
	ctx := context.Background()
	client := approvalClient(t, srv)

	pending, err := store.CreatePendingNode(ctx, db.PendingNodeParams{
		UUID:       "reject-success-uuid",
		SecretKey:  "secret-4",
		APIAddress: "127.0.0.1",
		APIPort:    8082,
	})
	require.NoError(t, err)
	require.Equal(t, "pending", pending.Status)

	resp, err := client.RejectNode(ctx, &panelv1.RejectNodeRequest{
		Id: pending.ID,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Verify the node was deleted.
	_, err = store.GetNodeByID(ctx, pending.ID)
	require.ErrorIs(t, err, db.ErrNotFound)
}

func TestRejectNode_NotFound(t *testing.T) {
	srv, _ := setupApprovalTestServer(t)
	ctx := context.Background()
	client := approvalClient(t, srv)

	_, err := client.RejectNode(ctx, &panelv1.RejectNodeRequest{
		Id: 99999,
	})
	require.Error(t, err)

	var connErr *connect.Error
	require.ErrorAs(t, err, &connErr)
	require.Equal(t, connect.CodeNotFound, connErr.Code())
}
