-- name: CreateTempChannelsGuildSettings :one
INSERT INTO guild_settings_tempchannels (guild_id)
    VALUES ($1)
RETURNING
    *;

-- name: GetTempChannelsGuildSettings :one
SELECT
    *
FROM
    guild_settings_tempchannels
WHERE
    guild_id = $1;

-- name: UpdateTempChannelsGuildSettings :execrows
UPDATE
    guild_settings_tempchannels
SET
    toggle_enabled = $2,
    toggle_autopurge = $3,
    channel_lobby = $4,
    channel_category = $5,
    default_user_count = $6
WHERE
    guild_id = $1;

