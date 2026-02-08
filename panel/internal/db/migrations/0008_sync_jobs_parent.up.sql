ALTER TABLE sync_jobs ADD COLUMN parent_job_id INTEGER REFERENCES sync_jobs(id);

CREATE INDEX IF NOT EXISTS idx_sync_jobs_parent_job_id ON sync_jobs(parent_job_id);
