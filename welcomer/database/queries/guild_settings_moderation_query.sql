-- name: CreateModerationGuildSettings :one
INSERT INTO guild_settings_moderation (guild_id)
    VALUES ($1)
RETURNING
    *;

-- name: GetModerationGuildSettings :one
SELECT
    *
FROM
    guild_settings_moderation
WHERE
    guild_id = $1;

-- name: UpdateModerationGuildSettings :execrows
UPDATE
    guild_settings_moderation
SET
    toggle_reasons_required = $2
WHERE
    guild_id = $1;

