CREATE TABLE IF NOT EXISTS job_checkpoints (
    job_name TEXT PRIMARY KEY,
    last_processed_ts TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
