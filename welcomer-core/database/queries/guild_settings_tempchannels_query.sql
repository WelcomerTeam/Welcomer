-- name: CreateTempChannelsGuildSettings :one
INSERT INTO guild_settings_tempchannels (guild_id, toggle_enabled, toggle_autopurge, channel_lobby, channel_category, default_user_count)
    VALUES ($1, $2, $3, $4, $5, $6)
RETURNING
    *;

-- name: CreateOrUpdateTempChannelsGuildSettings :one
INSERT INTO guild_settings_tempchannels (guild_id, toggle_enabled, toggle_autopurge, channel_lobby, channel_category, default_user_count)
    VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT(guild_id) DO UPDATE
    SET toggle_enabled = EXCLUDED.toggle_enabled,
        toggle_autopurge = EXCLUDED.toggle_autopurge,
        channel_lobby = EXCLUDED.channel_lobby,
        channel_category = EXCLUDED.channel_category,
        default_user_count = EXCLUDED.default_user_count
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

