-- name: CreateGuildPunishment :one
INSERT INTO guild_punishments (punishment_uuid, created_at, updated_at, guild_id, punishment_type, guild_moderation_event_uuid, user_id, moderator_id, data)
    VALUES (uuid_generate_v4 (), now(), now(), $1, $2, $3, $4, $5, $6)
RETURNING
    *;


-- name: GetGuildPunishment :one
SELECT
    *
FROM
    guild_punishments
WHERE
    punishment_uuid = $1;

-- name: GetGuildPunishmentsByGuildIDUserID :many
SELECT
    punishment_uuid,
    created_at,
    updated_at,
    guild_id,
    punishment_type,
    guild_moderation_event_uuid,
    user_id,
    moderator_id,
    data
FROM
    guild_punishments
WHERE
    guild_id = $1
    AND user_id = $2;

-- name: UpdateGuildPunishment :execrows
UPDATE
    guild_punishments
SET
    data = $2,
    updated_at = now()
WHERE
    punishment_uuid = $1;

