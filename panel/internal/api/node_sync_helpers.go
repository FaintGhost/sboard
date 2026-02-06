package api

import (
  "context"

  "sboard/panel/internal/db"
  "sboard/panel/internal/node"
)

type syncResult struct {
  Status string `json:"status"`
  Error  string `json:"error,omitempty"`
}

func trySyncNode(ctx context.Context, store *db.Store, n db.Node) syncResult {
  if n.GroupID == nil {
    return syncResult{Status: "error", Error: "node group_id not set"}
  }

  inbounds, err := store.ListInbounds(ctx, 10000, 0, n.ID)
  if err != nil {
    return syncResult{Status: "error", Error: "list inbounds failed"}
  }
  users, err := store.ListActiveUsersForGroup(ctx, *n.GroupID)
  if err != nil {
    return syncResult{Status: "error", Error: "list users failed"}
  }
  payload, err := node.BuildSyncPayload(n, inbounds, users)
  if err != nil {
    return syncResult{Status: "error", Error: err.Error()}
  }
  client := nodeClientFactory()
  if err := client.SyncConfig(ctx, n, payload); err != nil {
    return syncResult{Status: "error", Error: err.Error()}
  }
  return syncResult{Status: "ok"}
}

