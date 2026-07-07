-- name: CreateWelcomerTextGuildSettings :one
INSERT INTO guild_settings_welcomer_text (guild_id, toggle_enabled, channel, message_format, moderation_checkup_uuid)
    VALUES ($1, $2, $3, $4, $5)
RETURNING
    *;

-- name: CreateOrUpdateWelcomerTextGuildSettings :one
INSERT INTO guild_settings_welcomer_text (guild_id, toggle_enabled, channel, message_format, moderation_checkup_uuid)
    VALUES ($1, $2, $3, $4, $5)
ON CONFLICT(guild_id) DO UPDATE
    SET toggle_enabled = EXCLUDED.toggle_enabled,
        channel = EXCLUDED.channel,
        message_format = EXCLUDED.message_format,
        moderation_checkup_uuid = EXCLUDED.moderation_checkup_uuid
RETURNING
    *;

-- name: GetWelcomerTextGuildSettings :one
SELECT
    *
FROM
    guild_settings_welcomer_text
    LEFT JOIN moderation_checkup ON guild_settings_welcomer_text.moderation_checkup_uuid = moderation_checkup.checkup_uuid
WHERE
    guild_id = $1;

-- name: UpdateWelcomerTextGuildSettings :execrows
UPDATE
    guild_settings_welcomer_text
SET
    toggle_enabled = $2,
    channel = $3,
    message_format = $4,
    moderation_checkup_uuid = $5
WHERE
    guild_id = $1;

