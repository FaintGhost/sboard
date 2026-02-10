package api

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"sboard/panel/internal/db"
)

type syncJobListItemDTO struct {
	ID            int64   `json:"id"`
	NodeID        int64   `json:"node_id"`
	ParentJobID   *int64  `json:"parent_job_id,omitempty"`
	TriggerSource string  `json:"trigger_source"`
	Status        string  `json:"status"`
	InboundCount  int     `json:"inbound_count"`
	ActiveUsers   int     `json:"active_user_count"`
	PayloadHash   string  `json:"payload_hash"`
	AttemptCount  int     `json:"attempt_count"`
	DurationMS    int64   `json:"duration_ms"`
	ErrorSummary  string  `json:"error_summary"`
	CreatedAt     string  `json:"created_at"`
	StartedAt     *string `json:"started_at,omitempty"`
	FinishedAt    *string `json:"finished_at,omitempty"`
}

type syncJobDetailDTO struct {
	Job      syncJobListItemDTO   `json:"job"`
	Attempts []syncAttemptItemDTO `json:"attempts"`
}

type syncAttemptItemDTO struct {
	ID           int64   `json:"id"`
	AttemptNo    int     `json:"attempt_no"`
	Status       string  `json:"status"`
	HTTPStatus   int     `json:"http_status"`
	DurationMS   int64   `json:"duration_ms"`
	ErrorSummary string  `json:"error_summary"`
	BackoffMS    int64   `json:"backoff_ms"`
	StartedAt    string  `json:"started_at"`
	FinishedAt   *string `json:"finished_at,omitempty"`
}

func SyncJobsList(store *db.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !ensureStore(c, store) {
			return
		}

		limit, offset, err := parseLimitOffset(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pagination"})
			return
		}

		filter := db.SyncJobsListFilter{Limit: limit, Offset: offset}

		if rawNodeID := strings.TrimSpace(c.Query("node_id")); rawNodeID != "" {
			nodeID, err := parseID(rawNodeID)
			if err != nil || nodeID <= 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid node_id"})
				return
			}
			filter.NodeID = nodeID
		}

		if status := strings.TrimSpace(c.Query("status")); status != "" {
			switch status {
			case db.SyncJobStatusQueued, db.SyncJobStatusRunning, db.SyncJobStatusSuccess, db.SyncJobStatusFailed:
				filter.Status = status
			default:
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status"})
				return
			}
		}

		if trigger := strings.TrimSpace(c.Query("trigger_source")); trigger != "" {
			filter.TriggerSource = trigger
		}

		if fromRaw := strings.TrimSpace(c.Query("from")); fromRaw != "" {
			from, err := time.Parse(time.RFC3339, fromRaw)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid from"})
				return
			}
			filter.From = &from
		}

		if toRaw := strings.TrimSpace(c.Query("to")); toRaw != "" {
			to, err := time.Parse(time.RFC3339, toRaw)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid to"})
				return
			}
			filter.To = &to
		}

		items, err := store.ListSyncJobs(c.Request.Context(), filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "list sync jobs failed"})
			return
		}

		out := make([]syncJobListItemDTO, 0, len(items))
		for _, item := range items {
			out = append(out, toSyncJobListItemDTO(item))
		}
		c.JSON(http.StatusOK, gin.H{"data": out})
	}
}

func SyncJobsGet(store *db.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !ensureStore(c, store) {
			return
		}

		id, err := parseID(c.Param("id"))
		if err != nil || id <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		job, err := store.GetSyncJobByID(c.Request.Context(), id)
		if err != nil {
			if errors.Is(err, db.ErrNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "sync job not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "get sync job failed"})
			return
		}

		attempts, err := store.ListSyncAttemptsByJobID(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "list sync attempts failed"})
			return
		}

		outAttempts := make([]syncAttemptItemDTO, 0, len(attempts))
		for _, item := range attempts {
			outAttempts = append(outAttempts, toSyncAttemptDTO(item))
		}

		c.JSON(http.StatusOK, gin.H{"data": syncJobDetailDTO{
			Job:      toSyncJobListItemDTO(job),
			Attempts: outAttempts,
		}})
	}
}

func SyncJobsRetry(store *db.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !ensureStore(c, store) {
			return
		}

		id, err := parseID(c.Param("id"))
		if err != nil || id <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		parentJob, err := store.GetSyncJobByID(c.Request.Context(), id)
		if err != nil {
			if errors.Is(err, db.ErrNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "sync job not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "get sync job failed"})
			return
		}

		nodeItem, err := store.GetNodeByID(c.Request.Context(), parentJob.NodeID)
		if err != nil {
			if errors.Is(err, db.ErrNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "node not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "get node failed"})
			return
		}

		parentID := parentJob.ID
		result := trySyncNodeWithSourceAndParent(c.Request.Context(), store, nodeItem, triggerSourceRetry, &parentID)
		if result.Status != "ok" {
			c.JSON(http.StatusBadGateway, gin.H{"error": result.Error})
			return
		}

		latestJobs, err := store.ListSyncJobs(c.Request.Context(), db.SyncJobsListFilter{
			Limit:  1,
			Offset: 0,
			NodeID: parentJob.NodeID,
		})
		if err != nil || len(latestJobs) == 0 {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": toSyncJobListItemDTO(latestJobs[0])})
	}
}

func toSyncJobListItemDTO(item db.SyncJob) syncJobListItemDTO {
	out := syncJobListItemDTO{
		ID:            item.ID,
		NodeID:        item.NodeID,
		ParentJobID:   item.ParentJobID,
		TriggerSource: item.TriggerSource,
		Status:        item.Status,
		InboundCount:  item.InboundCount,
		ActiveUsers:   item.ActiveUserCount,
		PayloadHash:   item.PayloadHash,
		AttemptCount:  item.AttemptCount,
		DurationMS:    item.DurationMS,
		ErrorSummary:  item.ErrorSummary,
		CreatedAt:     formatTimeRFC3339OrEmpty(item.CreatedAt),
	}
	if item.StartedAt != nil {
		startedAt := formatTimeRFC3339OrEmpty(*item.StartedAt)
		out.StartedAt = &startedAt
	}
	if item.FinishedAt != nil {
		finishedAt := formatTimeRFC3339OrEmpty(*item.FinishedAt)
		out.FinishedAt = &finishedAt
	}
	return out
}

func toSyncAttemptDTO(item db.SyncAttempt) syncAttemptItemDTO {
	out := syncAttemptItemDTO{
		ID:           item.ID,
		AttemptNo:    item.AttemptNo,
		Status:       item.Status,
		HTTPStatus:   item.HTTPStatus,
		DurationMS:   item.DurationMS,
		ErrorSummary: item.ErrorSummary,
		BackoffMS:    item.BackoffMS,
		StartedAt:    formatTimeRFC3339OrEmpty(item.StartedAt),
	}
	if item.FinishedAt != nil {
		finishedAt := formatTimeRFC3339OrEmpty(*item.FinishedAt)
		out.FinishedAt = &finishedAt
	}
	return out
}
