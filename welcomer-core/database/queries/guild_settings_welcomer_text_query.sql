-- name: CreateWelcomerTextGuildSettings :one
INSERT INTO guild_settings_welcomer_text (guild_id, toggle_enabled, channel, message_format)
    VALUES ($1, $2, $3, $4)
RETURNING
    *;

-- name: CreateOrUpdateWelcomerTextGuildSettings :one
INSERT INTO guild_settings_welcomer_text (guild_id, toggle_enabled, channel, message_format)
    VALUES ($1, $2, $3, $4)
ON CONFLICT(guild_id) DO UPDATE
    SET toggle_enabled = EXCLUDED.toggle_enabled,
        channel = EXCLUDED.channel,
        message_format = EXCLUDED.message_format
RETURNING
    *;

-- name: GetWelcomerTextGuildSettings :one
SELECT
    *
FROM
    guild_settings_welcomer_text
WHERE
    guild_id = $1;

-- name: UpdateWelcomerTextGuildSettings :execrows
UPDATE
    guild_settings_welcomer_text
SET
    toggle_enabled = $2,
    channel = $3,
    message_format = $4
WHERE
    guild_id = $1;

