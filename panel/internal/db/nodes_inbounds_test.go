package db_test

import (
  "context"
  "encoding/json"
  "testing"

  "sboard/panel/internal/db"
  "github.com/stretchr/testify/require"
)

func TestNodeAndInboundCRUD(t *testing.T) {
  store := setupStore(t)
  ctx := context.Background()

  g, err := store.CreateGroup(ctx, "g1", "")
  require.NoError(t, err)

  node, err := store.CreateNode(ctx, db.NodeCreate{
    Name:          "n1",
    APIAddress:    "api.local",
    APIPort:       2222,
    SecretKey:     "secret",
    PublicAddress: "a.example.com",
    GroupID:       &g.ID,
  })
  require.NoError(t, err)
  require.NotZero(t, node.ID)
  require.NotNil(t, node.GroupID)

  inb, err := store.CreateInbound(ctx, db.InboundCreate{
    Tag:        "vless-in",
    NodeID:     node.ID,
    Protocol:   "vless",
    ListenPort: 443,
    PublicPort: 0,
    Settings:   json.RawMessage(`{}`),
  })
  require.NoError(t, err)
  require.Equal(t, "vless", inb.Protocol)

  listed, err := store.ListInbounds(ctx, 10, 0, node.ID)
  require.NoError(t, err)
  require.Len(t, listed, 1)

  err = store.DeleteNode(ctx, node.ID)
  require.Error(t, err)
  require.ErrorIs(t, err, db.ErrConflict)

  require.NoError(t, store.DeleteInbound(ctx, inb.ID))
  require.NoError(t, store.DeleteNode(ctx, node.ID))
}

