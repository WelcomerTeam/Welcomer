-- name: CreateLeaverGuildSettings :one
INSERT INTO guild_settings_leaver (guild_id, toggle_enabled, channel, message_format)
    VALUES ($1, $2, $3, $4)
RETURNING
    *;

-- name: CreateOrUpdateLeaverGuildSettings :one
INSERT INTO guild_settings_leaver (guild_id, toggle_enabled, channel, message_format)
    VALUES ($1, $2, $3, $4)
ON CONFLICT(guild_id) DO UPDATE
    SET toggle_enabled = EXCLUDED.toggle_enabled,
        channel = EXCLUDED.channel,
        message_format = EXCLUDED.message_format
RETURNING
    *;

-- name: GetLeaverGuildSettings :one
SELECT
    *
FROM
    guild_settings_leaver
WHERE
    guild_id = $1;

-- name: UpdateLeaverGuildSettings :execrows
UPDATE
    guild_settings_leaver
SET
    toggle_enabled = $2,
    channel = $3,
    message_format = $4
WHERE
    guild_id = $1;

