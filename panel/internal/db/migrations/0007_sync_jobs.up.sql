CREATE TABLE IF NOT EXISTS sync_jobs (
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

CREATE INDEX IF NOT EXISTS idx_sync_jobs_node_created_at ON sync_jobs(node_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_sync_jobs_status_created_at ON sync_jobs(status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_sync_jobs_trigger_created_at ON sync_jobs(trigger_source, created_at DESC);

CREATE TABLE IF NOT EXISTS sync_attempts (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  job_id INTEGER NOT NULL,
  attempt_no INTEGER NOT NULL,
  status TEXT NOT NULL,
  http_status INTEGER NOT NULL DEFAULT 0,
  duration_ms INTEGER NOT NULL DEFAULT 0,
  error_summary TEXT NOT NULL DEFAULT '',
  backoff_ms INTEGER NOT NULL DEFAULT 0,
  started_at TEXT NOT NULL,
  finished_at TEXT,
  FOREIGN KEY(job_id) REFERENCES sync_jobs(id) ON DELETE CASCADE,
  UNIQUE(job_id, attempt_no)
);

CREATE INDEX IF NOT EXISTS idx_sync_attempts_job_attempt ON sync_attempts(job_id, attempt_no ASC);
