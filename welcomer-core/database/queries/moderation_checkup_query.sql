-- name: CreateModerationCheckup :one
INSERT INTO moderation_checkup (checkup_uuid, guild_id, user_id, audit_type, started_at)
VALUES ($1, $2, $3, $4, now())
RETURNING checkup_uuid, guild_id, user_id, audit_type, started_at;

-- name: UpdateModerationCheckup :exec
UPDATE moderation_checkup
SET completed_at = now(),
    dom = $2,
    inv = $3,
    score_change = $4,
    score_safe = $5,
    score_question = $6,
    score_explicit = $7,
    is_blocked = $8
WHERE checkup_uuid = $1;