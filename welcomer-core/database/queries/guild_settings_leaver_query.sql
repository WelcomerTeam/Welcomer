-- name: CreateLeaverGuildSettings :one
INSERT INTO guild_settings_leaver (guild_id, toggle_enabled, channel, message_format, auto_delete_leaver_messages, leaver_message_lifetime)
    VALUES ($1, $2, $3, $4, $5, $6)
RETURNING
    *;

-- name: CreateOrUpdateLeaverGuildSettings :one
INSERT INTO guild_settings_leaver (guild_id, toggle_enabled, channel, message_format, auto_delete_leaver_messages, leaver_message_lifetime)
    VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT(guild_id) DO UPDATE
    SET toggle_enabled = EXCLUDED.toggle_enabled,
        channel = EXCLUDED.channel,
        message_format = EXCLUDED.message_format,
        auto_delete_leaver_messages = EXCLUDED.auto_delete_leaver_messages,
        leaver_message_lifetime = EXCLUDED.leaver_message_lifetime
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
    message_format = $4,
    auto_delete_leaver_messages = $5,
    leaver_message_lifetime = $6
WHERE
    guild_id = $1;

