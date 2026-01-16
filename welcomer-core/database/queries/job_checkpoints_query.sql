-- name: GetJobCheckpointByName :one
SELECT
    *
FROM
    job_checkpoints
WHERE
    job_name = $1 FOR UPDATE;

-- name: UpsertJobCheckpoint :exec
INSERT INTO job_checkpoints (job_name, last_processed_ts, updated_at)
VALUES ($1, $2, NOW())
ON CONFLICT (job_name) DO UPDATE SET
    last_processed_ts = EXCLUDED.last_processed_ts,
    updated_at = NOW();