package api

import (
	"context"
	"log"
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
	triggerSourceRetry    = "manual_retry"
	triggerSourceInbound  = "auto_inbound_change"
	triggerSourceUser     = "auto_user_change"
	triggerSourceGroup    = "auto_group_change"
	maxSyncErrorSummaryLn = 1024
)

type syncResult struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

func trySyncNode(ctx context.Context, store *db.Store, n db.Node) syncResult {
	return trySyncNodeWithSource(ctx, store, n, triggerSourceManual)
}

func trySyncNodeWithSource(ctx context.Context, store *db.Store, n db.Node, triggerSource string) syncResult {
	return trySyncNodeWithSourceAndParent(ctx, store, n, triggerSource, nil)
}

func trySyncNodeWithSourceAndParent(ctx context.Context, store *db.Store, n db.Node, triggerSource string, parentJobID *int64) syncResult {
	lock := nodeLock(n.ID)
	lock.Lock()
	defer lock.Unlock()

	if triggerSource == "" {
		triggerSource = triggerSourceManual
	}

	job, err := store.CreateSyncJob(ctx, db.SyncJobCreate{
		NodeID:        n.ID,
		ParentJobID:   parentJobID,
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
