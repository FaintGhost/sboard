DROP INDEX IF EXISTS idx_sync_attempts_job_attempt;
DROP TABLE IF EXISTS sync_attempts;

DROP INDEX IF EXISTS idx_sync_jobs_trigger_created_at;
DROP INDEX IF EXISTS idx_sync_jobs_status_created_at;
DROP INDEX IF EXISTS idx_sync_jobs_node_created_at;
DROP TABLE IF EXISTS sync_jobs;
