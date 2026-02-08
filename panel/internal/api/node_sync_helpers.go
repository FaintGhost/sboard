package api

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strings"

	"sboard/panel/internal/db"
	"sboard/panel/internal/node"
)

const panelSyncDebugPayloadEnv = "PANEL_SYNC_DEBUG_PAYLOAD"

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
	if shouldDebugSyncPayload() {
		log.Printf("[sync] node=%d payload=%s", n.ID, syncPayloadDebugJSON(payload))
	}
	client := nodeClientFactory()
	if err := client.SyncConfig(ctx, n, payload); err != nil {
		msg := strings.TrimSpace(err.Error())
		if strings.Contains(msg, "node sync status ") {
			return syncResult{Status: "error", Error: msg}
		}
		return syncResult{Status: "error", Error: "node sync request failed: " + msg}
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
				if isNodeUnreachableSyncError(res.Error) {
					if err := store.MarkNodeOffline(ctx, n.ID); err != nil {
						log.Printf("[sync] auto-sync mark node=%d offline failed: %v", n.ID, err)
					}
				}
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

func isNodeUnreachableSyncError(errMsg string) bool {
	return strings.HasPrefix(strings.TrimSpace(errMsg), "node sync request failed:")
}

func shouldDebugSyncPayload() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(panelSyncDebugPayloadEnv)))
	switch v {
	case "1", "true", "yes", "y", "on":
		return true
	default:
		return false
	}
}

func syncPayloadDebugJSON(payload node.SyncPayload) string {
	raw, err := json.Marshal(payload)
	if err != nil {
		return `{"error":"marshal payload failed"}`
	}
	var data any
	if err := json.Unmarshal(raw, &data); err != nil {
		return `{"error":"decode payload failed"}`
	}
	sanitized := sanitizeSyncPayloadForLog(data, "")
	out, err := json.Marshal(sanitized)
	if err != nil {
		return `{"error":"encode sanitized payload failed"}`
	}
	if len(out) > 65535 {
		return string(out[:65535]) + "...(truncated)"
	}
	return string(out)
}

func sanitizeSyncPayloadForLog(value any, key string) any {
	key = strings.ToLower(strings.TrimSpace(key))
	switch v := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(v))
		for k, val := range v {
			out[k] = sanitizeSyncPayloadForLog(val, k)
		}
		return out
	case []any:
		out := make([]any, 0, len(v))
		for _, item := range v {
			out = append(out, sanitizeSyncPayloadForLog(item, key))
		}
		return out
	case string:
		if key == "password" || key == "uuid" {
			return maskSyncCredential(v)
		}
		return v
	default:
		return value
	}
}

func maskSyncCredential(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if len(value) <= 8 {
		return "***"
	}
	return value[:4] + "..." + value[len(value)-4:]
}
