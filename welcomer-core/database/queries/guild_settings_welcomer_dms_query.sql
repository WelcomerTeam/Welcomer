-- name: CreateWelcomerDMsGuildSettings :one
INSERT INTO guild_settings_welcomer_dms (guild_id, toggle_enabled, toggle_use_text_format, toggle_include_image, message_format, moderation_checkup_uuid)
    VALUES ($1, $2, $3, $4, $5, $6)
RETURNING
    *;

-- name: CreateOrUpdateWelcomerDMsGuildSettings :one
INSERT INTO guild_settings_welcomer_dms (guild_id, toggle_enabled, toggle_use_text_format, toggle_include_image, message_format, moderation_checkup_uuid)
    VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT(guild_id) DO UPDATE
    SET toggle_enabled = EXCLUDED.toggle_enabled, 
        toggle_use_text_format = EXCLUDED.toggle_use_text_format, 
        toggle_include_image = EXCLUDED.toggle_include_image, 
        message_format = EXCLUDED.message_format,
        moderation_checkup_uuid = EXCLUDED.moderation_checkup_uuid
RETURNING
    *;

-- name: GetWelcomerDMsGuildSettings :one
SELECT
    *
FROM
    guild_settings_welcomer_dms
    LEFT JOIN moderation_checkup ON guild_settings_welcomer_dms.moderation_checkup_uuid = moderation_checkup.checkup_uuid
WHERE
    guild_settings_welcomer_dms.guild_id = $1;

-- name: UpdateWelcomerDMsGuildSettings :execrows
UPDATE
    guild_settings_welcomer_dms
SET
    toggle_enabled = $2,
    toggle_use_text_format = $3,
    toggle_include_image = $4,
    message_format = $5,
    moderation_checkup_uuid = $6
WHERE
    guild_id = $1;

