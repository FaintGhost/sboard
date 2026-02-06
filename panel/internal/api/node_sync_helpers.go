package api

import (
  "context"
  "log"

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
  // Log each inbound's key fields before sending to node (no secrets).
  for _, inb := range payload.Inbounds {
    pw, _ := inb["password"].(string)
    method, _ := inb["method"].(string)
    users, _ := inb["users"].([]map[string]any)
    log.Printf("[sync] node=%d inbound tag=%v type=%v method=%s password_len=%d users=%d",
      n.ID, inb["tag"], inb["type"], method, len(pw), len(users))
  }
  client := nodeClientFactory()
  if err := client.SyncConfig(ctx, n, payload); err != nil {
    return syncResult{Status: "error", Error: err.Error()}
  }
  return syncResult{Status: "ok"}
}

