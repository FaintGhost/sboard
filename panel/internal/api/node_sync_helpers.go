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

func syncNodesByGroupIDs(ctx context.Context, store *db.Store, groupIDs []int64) {
  uniqueGroups := uniquePositiveInt64(groupIDs)
  if len(uniqueGroups) == 0 {
    return
  }

  groupSet := make(map[int64]struct{}, len(uniqueGroups))
  for _, id := range uniqueGroups {
    groupSet[id] = struct{}{}
  }

  const pageSize = 200
  offset := 0
  for {
    nodes, err := store.ListNodes(ctx, pageSize, offset)
    if err != nil {
      log.Printf("[sync] auto-sync list nodes failed: %v", err)
      return
    }
    for _, n := range nodes {
      if n.GroupID == nil {
        continue
      }
      if _, ok := groupSet[*n.GroupID]; !ok {
        continue
      }
      res := trySyncNode(ctx, store, n)
      if res.Status != "ok" {
        log.Printf("[sync] auto-sync node=%d failed: %s", n.ID, res.Error)
        continue
      }
      if err := store.MarkNodeOnline(ctx, n.ID, store.Now().UTC()); err != nil {
        log.Printf("[sync] auto-sync mark node=%d online failed: %v", n.ID, err)
      }
    }

    if len(nodes) < pageSize {
      return
    }
    offset += len(nodes)
  }
}

func syncNodesForUser(ctx context.Context, store *db.Store, userID int64) {
  groupIDs, err := store.ListUserGroupIDs(ctx, userID)
  if err != nil {
    log.Printf("[sync] auto-sync list user groups failed user=%d: %v", userID, err)
    return
  }
  syncNodesByGroupIDs(ctx, store, groupIDs)
}

func uniquePositiveInt64(items []int64) []int64 {
  out := make([]int64, 0, len(items))
  seen := make(map[int64]struct{}, len(items))
  for _, item := range items {
    if item <= 0 {
      continue
    }
    if _, ok := seen[item]; ok {
      continue
    }
    seen[item] = struct{}{}
    out = append(out, item)
  }
  return out
}
