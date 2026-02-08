package api

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"sboard/panel/internal/db"
	"sboard/panel/internal/node"
)

const panelSyncDebugPayloadEnv = "PANEL_SYNC_DEBUG_PAYLOAD"

const (
	syncMaxAttempts       = 3
	syncRetryBaseDelay    = 200 * time.Millisecond
	syncJobKeepPerNode    = 500
	triggerSourceManual   = "manual_node_sync"
	triggerSourceInbound  = "auto_inbound_change"
	triggerSourceUser     = "auto_user_change"
	triggerSourceGroup    = "auto_group_change"
	maxSyncErrorSummaryLn = 1024
)

var nodeSyncLocks = struct {
	mu    sync.Mutex
	locks map[int64]*sync.Mutex
}{locks: map[int64]*sync.Mutex{}}

type syncResult struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

func trySyncNode(ctx context.Context, store *db.Store, n db.Node) syncResult {
	return trySyncNodeWithSource(ctx, store, n, triggerSourceManual)
}

func trySyncNodeWithSource(ctx context.Context, store *db.Store, n db.Node, triggerSource string) syncResult {
	lock := nodeLock(n.ID)
	lock.Lock()
	defer lock.Unlock()

	if triggerSource == "" {
		triggerSource = triggerSourceManual
	}

	job, err := store.CreateSyncJob(ctx, db.SyncJobCreate{
		NodeID:        n.ID,
		TriggerSource: triggerSource,
	})
	if err != nil {
		return syncResult{Status: "error", Error: "create sync job failed"}
	}
	defer func() {
		if err := store.PruneSyncJobsByNode(ctx, n.ID, syncJobKeepPerNode); err != nil {
			log.Printf("[sync] prune sync jobs failed node=%d: %v", n.ID, err)
		}
	}()

	return runSyncJob(ctx, store, n, job)
}

func runSyncJob(ctx context.Context, store *db.Store, n db.Node, job db.SyncJob) syncResult {
	jobStart := store.Now().UTC()
	finishWithNoAttempt := func(errMsg string) syncResult {
		errMsg = normalizeSyncError(errMsg)
		_ = store.UpdateSyncJobFinish(ctx, job.ID, db.SyncJobFinishUpdate{
			Status:       db.SyncJobStatusFailed,
			AttemptCount: 0,
			ErrorSummary: errMsg,
			FinishedAt:   store.Now().UTC(),
			DurationMS:   durationMSSince(jobStart, store.Now().UTC()),
		})
		return syncResult{Status: "error", Error: errMsg}
	}

	if n.GroupID == nil {
		return finishWithNoAttempt("node group_id not set")
	}

	inbounds, err := store.ListInbounds(ctx, 10000, 0, n.ID)
	if err != nil {
		return finishWithNoAttempt("list inbounds failed")
	}
	users, err := store.ListActiveUsersForGroup(ctx, *n.GroupID)
	if err != nil {
		return finishWithNoAttempt("list users failed")
	}
	payload, err := node.BuildSyncPayload(n, inbounds, users)
	if err != nil {
		return finishWithNoAttempt(err.Error())
	}
	if err := store.UpdateSyncJobStart(ctx, job.ID, db.SyncJobStartUpdate{
		InboundCount:    len(inbounds),
		ActiveUserCount: len(users),
		PayloadHash:     payloadHash(payload),
		StartedAt:       jobStart,
	}); err != nil {
		log.Printf("[sync] update sync job start failed job=%d node=%d: %v", job.ID, n.ID, err)
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

	attemptCount := 0
	lastErrMsg := ""
	for attemptNo := 1; attemptNo <= syncMaxAttempts; attemptNo++ {
		attemptCount = attemptNo
		backoff := int64(0)
		if attemptNo > 1 {
			backoff = int64(syncRetryBaseDelay.Milliseconds()) << (attemptNo - 2)
			if err := sleepWithContext(ctx, time.Duration(backoff)*time.Millisecond); err != nil {
				lastErrMsg = normalizeSyncError(err.Error())
				break
			}
		}

		attemptStartedAt := store.Now().UTC()
		attempt, err := store.CreateSyncAttempt(ctx, db.SyncAttemptCreate{
			JobID:     job.ID,
			AttemptNo: attemptNo,
			Status:    db.SyncAttemptStatusRunning,
			BackoffMS: backoff,
			StartedAt: attemptStartedAt,
		})
		if err != nil {
			log.Printf("[sync] create attempt failed job=%d attempt=%d: %v", job.ID, attemptNo, err)
		}

		err = client.SyncConfig(ctx, n, payload)
		attemptFinishedAt := store.Now().UTC()
		if err == nil {
			if attempt.ID > 0 {
				if uerr := store.UpdateSyncAttemptFinish(ctx, attempt.ID, db.SyncAttemptFinishUpdate{
					Status:     db.SyncAttemptStatusSuccess,
					HTTPStatus: 200,
					FinishedAt: attemptFinishedAt,
					DurationMS: durationMSSince(attemptStartedAt, attemptFinishedAt),
				}); uerr != nil {
					log.Printf("[sync] finish attempt success update failed job=%d attempt=%d: %v", job.ID, attemptNo, uerr)
				}
			}
			if uerr := store.UpdateSyncJobFinish(ctx, job.ID, db.SyncJobFinishUpdate{
				Status:       db.SyncJobStatusSuccess,
				AttemptCount: attemptNo,
				FinishedAt:   attemptFinishedAt,
				DurationMS:   durationMSSince(jobStart, attemptFinishedAt),
			}); uerr != nil {
				log.Printf("[sync] finish job success update failed job=%d: %v", job.ID, uerr)
			}
			return syncResult{Status: "ok"}
		}

		msg := normalizeSyncClientError(err)
		lastErrMsg = msg
		httpStatus := parseSyncHTTPStatus(err)
		if attempt.ID > 0 {
			if uerr := store.UpdateSyncAttemptFinish(ctx, attempt.ID, db.SyncAttemptFinishUpdate{
				Status:       db.SyncAttemptStatusFailed,
				HTTPStatus:   httpStatus,
				ErrorSummary: msg,
				FinishedAt:   attemptFinishedAt,
				DurationMS:   durationMSSince(attemptStartedAt, attemptFinishedAt),
			}); uerr != nil {
				log.Printf("[sync] finish attempt failed update failed job=%d attempt=%d: %v", job.ID, attemptNo, uerr)
			}
		}
	}

	if err := store.UpdateSyncJobFinish(ctx, job.ID, db.SyncJobFinishUpdate{
		Status:       db.SyncJobStatusFailed,
		AttemptCount: attemptCount,
		ErrorSummary: lastErrMsg,
		FinishedAt:   store.Now().UTC(),
		DurationMS:   durationMSSince(jobStart, store.Now().UTC()),
	}); err != nil {
		log.Printf("[sync] finish job failed update failed job=%d: %v", job.ID, err)
	}
	return syncResult{Status: "error", Error: lastErrMsg}
}

func syncNodesByGroupIDs(ctx context.Context, store *db.Store, groupIDs []int64) {
	syncNodesByGroupIDsWithSource(ctx, store, groupIDs, triggerSourceGroup)
}

func syncNodesByGroupIDsWithSource(ctx context.Context, store *db.Store, groupIDs []int64, triggerSource string) {
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
			res := trySyncNodeWithSource(ctx, store, n, triggerSource)
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
	syncNodesByGroupIDsWithSource(ctx, store, groupIDs, triggerSourceUser)
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
