-- name: CreateModerationCheckupQueue :one
INSERT INTO moderation_checkup_queue (checkup_queue_uuid, guild_id, user_id, data_type, created_at, value)
VALUES (uuid_generate_v7(), $1, $2, $3, now(), $4)
RETURNING checkup_queue_uuid, guild_id, user_id, data_type, created_at, value;

-- name: FetchModerationCheckupQueue :many
SELECT checkup_queue_uuid, guild_id, user_id, data_type, created_at, value
FROM moderation_checkup_queue
ORDER BY created_at ASC
LIMIT $1;

-- name: RemoveFromModerationCheckupQueue :exec
DELETE FROM moderation_checkup_queue
WHERE checkup_queue_uuid IN ($1);
