DROP INDEX IF EXISTS idx_sync_jobs_parent_job_id;

CREATE TABLE sync_jobs__old (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  node_id INTEGER NOT NULL,
  trigger_source TEXT NOT NULL,
  status TEXT NOT NULL,
  inbound_count INTEGER NOT NULL DEFAULT 0,
  active_user_count INTEGER NOT NULL DEFAULT 0,
  payload_hash TEXT NOT NULL DEFAULT '',
  attempt_count INTEGER NOT NULL DEFAULT 0,
  started_at TEXT,
  finished_at TEXT,
  duration_ms INTEGER NOT NULL DEFAULT 0,
  error_summary TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY(node_id) REFERENCES nodes(id) ON DELETE CASCADE
);

INSERT INTO sync_jobs__old (
  id, node_id, trigger_source, status, inbound_count, active_user_count,
  payload_hash, attempt_count, started_at, finished_at, duration_ms,
  error_summary, created_at, updated_at
)
SELECT
  id, node_id, trigger_source, status, inbound_count, active_user_count,
  payload_hash, attempt_count, started_at, finished_at, duration_ms,
  error_summary, created_at, updated_at
FROM sync_jobs;

DROP TABLE sync_jobs;
ALTER TABLE sync_jobs__old RENAME TO sync_jobs;

CREATE INDEX IF NOT EXISTS idx_sync_jobs_node_created_at ON sync_jobs(node_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_sync_jobs_status_created_at ON sync_jobs(status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_sync_jobs_trigger_created_at ON sync_jobs(trigger_source, created_at DESC);
