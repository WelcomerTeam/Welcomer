-- name: CreateModerationCheckupQueue :one
INSERT INTO moderation_checkup_queue (checkup_queue_uuid, guild_id, user_id, audit_type, created_at, value)
VALUES ($1, $2, $3, $4, now(), $5)
RETURNING checkup_queue_uuid, guild_id, user_id, audit_type, created_at, value;

-- name: FetchModerationCheckupQueue :many
SELECT checkup_queue_uuid, guild_id, user_id, audit_type, created_at, value
FROM moderation_checkup_queue
ORDER BY created_at ASC
LIMIT $1;

-- name: RemoveFromModerationCheckupQueue :exec
DELETE FROM moderation_checkup_queue
WHERE checkup_queue_uuid IN ($1);
