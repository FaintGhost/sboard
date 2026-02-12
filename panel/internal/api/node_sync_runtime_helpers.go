package api

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"sboard/panel/internal/node"
)

var nodeSyncLocks = struct {
	mu    sync.Mutex
	locks map[int64]*sync.Mutex
}{locks: map[int64]*sync.Mutex{}}

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

func nodeLock(nodeID int64) *sync.Mutex {
	nodeSyncLocks.mu.Lock()
	defer nodeSyncLocks.mu.Unlock()
	lock, ok := nodeSyncLocks.locks[nodeID]
	if ok {
		return lock
	}
	lock = &sync.Mutex{}
	nodeSyncLocks.locks[nodeID] = lock
	return lock
}

func payloadHash(payload node.SyncPayload) string {
	type inboundSummary struct {
		Tag   string `json:"tag"`
		Type  string `json:"type"`
		Port  int    `json:"listen_port"`
		Users int    `json:"users_count"`
	}
	type payloadSummary struct {
		InboundCount int              `json:"inbound_count"`
		Inbounds     []inboundSummary `json:"inbounds"`
	}

	summary := payloadSummary{
		InboundCount: len(payload.Inbounds),
		Inbounds:     make([]inboundSummary, 0, len(payload.Inbounds)),
	}
	for _, inb := range payload.Inbounds {
		tag, _ := inb["tag"].(string)
		typ, _ := inb["type"].(string)
		port := intFromAny(inb["listen_port"])
		usersCount := usersCountFromAny(inb["users"])
		summary.Inbounds = append(summary.Inbounds, inboundSummary{
			Tag:   tag,
			Type:  typ,
			Port:  port,
			Users: usersCount,
		})
	}

	raw, err := json.Marshal(summary)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(raw)
	return fmt.Sprintf("%x", sum)
}

func usersCountFromAny(value any) int {
	switch v := value.(type) {
	case []map[string]any:
		return len(v)
	case []any:
		return len(v)
	default:
		return 0
	}
}

func intFromAny(value any) int {
	switch v := value.(type) {
	case int:
		return v
	case int32:
		return int(v)
	case int64:
		return int(v)
	case float64:
		return int(v)
	case float32:
		return int(v)
	default:
		return 0
	}
}

func parseSyncHTTPStatus(err error) int {
	if err == nil {
		return 200
	}
	msg := strings.TrimSpace(err.Error())
	const prefix = "node sync status "
	if !strings.HasPrefix(msg, prefix) {
		return 0
	}
	remain := strings.TrimPrefix(msg, prefix)
	idx := strings.Index(remain, ":")
	if idx <= 0 {
		return 0
	}
	codeStr := strings.TrimSpace(remain[:idx])
	code, err := strconv.Atoi(codeStr)
	if err != nil || code < 100 || code > 599 {
		return 0
	}
	return code
}

func normalizeSyncClientError(err error) string {
	if err == nil {
		return ""
	}
	msg := strings.TrimSpace(err.Error())
	if strings.Contains(msg, "node sync status ") {
		return truncateSyncError(msg)
	}
	return truncateSyncError("node sync request failed: " + msg)
}

func normalizeSyncError(errMsg string) string {
	errMsg = strings.TrimSpace(errMsg)
	if errMsg == "" {
		return "sync failed"
	}
	return truncateSyncError(errMsg)
}

func truncateSyncError(errMsg string) string {
	errMsg = strings.TrimSpace(errMsg)
	if len(errMsg) <= maxSyncErrorSummaryLn {
		return errMsg
	}
	return errMsg[:maxSyncErrorSummaryLn]
}

func sleepWithContext(ctx context.Context, delay time.Duration) error {
	if delay <= 0 {
		return nil
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func durationMSSince(startedAt time.Time, finishedAt time.Time) int64 {
	if startedAt.IsZero() || finishedAt.Before(startedAt) {
		return 0
	}
	return finishedAt.Sub(startedAt).Milliseconds()
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
