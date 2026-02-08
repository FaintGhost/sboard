package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

const (
	SyncJobStatusQueued  = "queued"
	SyncJobStatusRunning = "running"
	SyncJobStatusSuccess = "success"
	SyncJobStatusFailed  = "failed"

	SyncAttemptStatusRunning = "running"
	SyncAttemptStatusSuccess = "success"
	SyncAttemptStatusFailed  = "failed"
)

type SyncJob struct {
	ID              int64
	NodeID          int64
	ParentJobID     *int64
	TriggerSource   string
	Status          string
	InboundCount    int
	ActiveUserCount int
	PayloadHash     string
	AttemptCount    int
	StartedAt       *time.Time
	FinishedAt      *time.Time
	DurationMS      int64
	ErrorSummary    string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type SyncAttempt struct {
	ID           int64
	JobID        int64
	AttemptNo    int
	Status       string
	HTTPStatus   int
	DurationMS   int64
	ErrorSummary string
	BackoffMS    int64
	StartedAt    time.Time
	FinishedAt   *time.Time
}

type SyncJobCreate struct {
	NodeID        int64
	ParentJobID   *int64
	TriggerSource string
}

type SyncJobsListFilter struct {
	Limit         int
	Offset        int
	NodeID        int64
	Status        string
	TriggerSource string
	From          *time.Time
	To            *time.Time
}

type SyncJobStartUpdate struct {
	InboundCount    int
	ActiveUserCount int
	PayloadHash     string
	StartedAt       time.Time
}

type SyncJobFinishUpdate struct {
	Status       string
	AttemptCount int
	ErrorSummary string
	FinishedAt   time.Time
	DurationMS   int64
}

type SyncAttemptCreate struct {
	JobID      int64
	AttemptNo  int
	Status     string
	HTTPStatus int
	BackoffMS  int64
	StartedAt  time.Time
}

type SyncAttemptFinishUpdate struct {
	Status       string
	HTTPStatus   int
	ErrorSummary string
	FinishedAt   time.Time
	DurationMS   int64
}

func (s *Store) CreateSyncJob(ctx context.Context, input SyncJobCreate) (SyncJob, error) {
	if input.NodeID <= 0 {
		return SyncJob{}, fmt.Errorf("invalid node id")
	}
	if input.TriggerSource == "" {
		return SyncJob{}, fmt.Errorf("invalid trigger source")
	}
	now := s.nowUTC()
	res, err := s.DB.ExecContext(
		ctx,
		`INSERT INTO sync_jobs (node_id, parent_job_id, trigger_source, status, created_at, updated_at)
     VALUES (?, ?, ?, ?, ?, ?)`,
		input.NodeID,
		nullInt64(input.ParentJobID),
		input.TriggerSource,
		SyncJobStatusQueued,
		now.Format(time.RFC3339),
		now.Format(time.RFC3339),
	)
	if err != nil {
		return SyncJob{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return SyncJob{}, err
	}
	return s.GetSyncJobByID(ctx, id)
}

func (s *Store) UpdateSyncJobStart(ctx context.Context, jobID int64, input SyncJobStartUpdate) error {
	if jobID <= 0 {
		return fmt.Errorf("invalid job id")
	}
	startedAt := input.StartedAt
	if startedAt.IsZero() {
		startedAt = s.nowUTC()
	}
	_, err := s.DB.ExecContext(
		ctx,
		`UPDATE sync_jobs
     SET status = ?, inbound_count = ?, active_user_count = ?, payload_hash = ?, started_at = ?, updated_at = ?
     WHERE id = ?`,
		SyncJobStatusRunning,
		input.InboundCount,
		input.ActiveUserCount,
		input.PayloadHash,
		startedAt.Format(time.RFC3339),
		s.nowUTC().Format(time.RFC3339),
		jobID,
	)
	return err
}

func (s *Store) UpdateSyncJobFinish(ctx context.Context, jobID int64, input SyncJobFinishUpdate) error {
	if jobID <= 0 {
		return fmt.Errorf("invalid job id")
	}
	if input.Status == "" {
		return fmt.Errorf("invalid status")
	}
	finishedAt := input.FinishedAt
	if finishedAt.IsZero() {
		finishedAt = s.nowUTC()
	}
	_, err := s.DB.ExecContext(
		ctx,
		`UPDATE sync_jobs
     SET status = ?, attempt_count = ?, error_summary = ?, finished_at = ?, duration_ms = ?, updated_at = ?
     WHERE id = ?`,
		input.Status,
		input.AttemptCount,
		input.ErrorSummary,
		finishedAt.Format(time.RFC3339),
		input.DurationMS,
		s.nowUTC().Format(time.RFC3339),
		jobID,
	)
	return err
}

func (s *Store) CreateSyncAttempt(ctx context.Context, input SyncAttemptCreate) (SyncAttempt, error) {
	if input.JobID <= 0 {
		return SyncAttempt{}, fmt.Errorf("invalid job id")
	}
	if input.AttemptNo <= 0 {
		return SyncAttempt{}, fmt.Errorf("invalid attempt no")
	}
	status := input.Status
	if status == "" {
		status = SyncAttemptStatusRunning
	}
	startedAt := input.StartedAt
	if startedAt.IsZero() {
		startedAt = s.nowUTC()
	}
	res, err := s.DB.ExecContext(
		ctx,
		`INSERT INTO sync_attempts
      (job_id, attempt_no, status, http_status, backoff_ms, started_at)
     VALUES (?, ?, ?, ?, ?, ?)`,
		input.JobID,
		input.AttemptNo,
		status,
		input.HTTPStatus,
		input.BackoffMS,
		startedAt.Format(time.RFC3339),
	)
	if err != nil {
		return SyncAttempt{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return SyncAttempt{}, err
	}
	return s.GetSyncAttemptByID(ctx, id)
}

func (s *Store) UpdateSyncAttemptFinish(ctx context.Context, attemptID int64, input SyncAttemptFinishUpdate) error {
	if attemptID <= 0 {
		return fmt.Errorf("invalid attempt id")
	}
	finishedAt := input.FinishedAt
	if finishedAt.IsZero() {
		finishedAt = s.nowUTC()
	}
	_, err := s.DB.ExecContext(
		ctx,
		`UPDATE sync_attempts
     SET status = ?, http_status = ?, error_summary = ?, finished_at = ?, duration_ms = ?
     WHERE id = ?`,
		input.Status,
		input.HTTPStatus,
		input.ErrorSummary,
		finishedAt.Format(time.RFC3339),
		input.DurationMS,
		attemptID,
	)
	return err
}

func (s *Store) GetSyncJobByID(ctx context.Context, id int64) (SyncJob, error) {
	row := s.DB.QueryRowContext(ctx, `
    SELECT id, node_id, parent_job_id, trigger_source, status, inbound_count, active_user_count, payload_hash,
      attempt_count, started_at, finished_at, duration_ms, error_summary, created_at, updated_at
    FROM sync_jobs WHERE id = ?
  `, id)
	return scanSyncJob(row)
}

func (s *Store) ListSyncJobs(ctx context.Context, filter SyncJobsListFilter) ([]SyncJob, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	wheres := make([]string, 0, 4)
	args := make([]any, 0, 8)
	if filter.NodeID > 0 {
		wheres = append(wheres, "node_id = ?")
		args = append(args, filter.NodeID)
	}
	if strings.TrimSpace(filter.Status) != "" {
		wheres = append(wheres, "status = ?")
		args = append(args, strings.TrimSpace(filter.Status))
	}
	if strings.TrimSpace(filter.TriggerSource) != "" {
		wheres = append(wheres, "trigger_source = ?")
		args = append(args, strings.TrimSpace(filter.TriggerSource))
	}
	if filter.From != nil {
		wheres = append(wheres, "created_at >= ?")
		args = append(args, filter.From.UTC().Format(time.RFC3339))
	}
	if filter.To != nil {
		wheres = append(wheres, "created_at < ?")
		args = append(args, filter.To.UTC().Format(time.RFC3339))
	}

	query := `
    SELECT id, node_id, parent_job_id, trigger_source, status, inbound_count, active_user_count, payload_hash,
      attempt_count, started_at, finished_at, duration_ms, error_summary, created_at, updated_at
    FROM sync_jobs`
	if len(wheres) > 0 {
		query += " WHERE " + strings.Join(wheres, " AND ")
	}
	query += " ORDER BY created_at DESC, id DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := s.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]SyncJob, 0)
	for rows.Next() {
		item, err := scanSyncJob(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (s *Store) GetSyncAttemptByID(ctx context.Context, id int64) (SyncAttempt, error) {
	row := s.DB.QueryRowContext(ctx, `
    SELECT id, job_id, attempt_no, status, http_status, duration_ms, error_summary, backoff_ms, started_at, finished_at
    FROM sync_attempts WHERE id = ?
  `, id)
	return scanSyncAttempt(row)
}

func (s *Store) ListSyncAttemptsByJobID(ctx context.Context, jobID int64) ([]SyncAttempt, error) {
	rows, err := s.DB.QueryContext(ctx, `
    SELECT id, job_id, attempt_no, status, http_status, duration_ms, error_summary, backoff_ms, started_at, finished_at
    FROM sync_attempts WHERE job_id = ? ORDER BY attempt_no ASC
  `, jobID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]SyncAttempt, 0)
	for rows.Next() {
		item, err := scanSyncAttempt(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (s *Store) PruneSyncJobsByNode(ctx context.Context, nodeID int64, keep int) error {
	if nodeID <= 0 {
		return nil
	}
	if keep <= 0 {
		keep = 1
	}
	_, err := s.DB.ExecContext(ctx, `
    DELETE FROM sync_jobs
    WHERE node_id = ? AND id NOT IN (
      SELECT id FROM sync_jobs WHERE node_id = ? ORDER BY created_at DESC LIMIT ?
    )
  `, nodeID, nodeID, keep)
	return err
}

func scanSyncJob(row rowScanner) (SyncJob, error) {
	var item SyncJob
	var parentJobID sql.NullInt64
	var startedAt sql.NullString
	var finishedAt sql.NullString
	var createdAt string
	var updatedAt string

	if err := row.Scan(
		&item.ID,
		&item.NodeID,
		&parentJobID,
		&item.TriggerSource,
		&item.Status,
		&item.InboundCount,
		&item.ActiveUserCount,
		&item.PayloadHash,
		&item.AttemptCount,
		&startedAt,
		&finishedAt,
		&item.DurationMS,
		&item.ErrorSummary,
		&createdAt,
		&updatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return SyncJob{}, ErrNotFound
		}
		return SyncJob{}, err
	}

	if parentJobID.Valid {
		item.ParentJobID = &parentJobID.Int64
	}

	if startedAt.Valid {
		if t, err := parseSQLiteTime(startedAt.String); err == nil {
			item.StartedAt = &t
		}
	}
	if finishedAt.Valid {
		if t, err := parseSQLiteTime(finishedAt.String); err == nil {
			item.FinishedAt = &t
		}
	}
	if t, err := parseSQLiteTime(createdAt); err == nil {
		item.CreatedAt = t
	}
	if t, err := parseSQLiteTime(updatedAt); err == nil {
		item.UpdatedAt = t
	}
	return item, nil
}

func scanSyncAttempt(row rowScanner) (SyncAttempt, error) {
	var item SyncAttempt
	var startedAt string
	var finishedAt sql.NullString
	if err := row.Scan(
		&item.ID,
		&item.JobID,
		&item.AttemptNo,
		&item.Status,
		&item.HTTPStatus,
		&item.DurationMS,
		&item.ErrorSummary,
		&item.BackoffMS,
		&startedAt,
		&finishedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return SyncAttempt{}, ErrNotFound
		}
		return SyncAttempt{}, err
	}

	if t, err := parseSQLiteTime(startedAt); err == nil {
		item.StartedAt = t
	}
	if finishedAt.Valid {
		if t, err := parseSQLiteTime(finishedAt.String); err == nil {
			item.FinishedAt = &t
		}
	}
	return item, nil
}
