-- name: CreateWelcomerDMsGuildSettings :one
INSERT INTO guild_settings_welcomer_dms (guild_id)
    VALUES ($1)
RETURNING
    *;

-- name: GetWelcomerDMsGuildSettings :one
SELECT
    *
FROM
    guild_settings_welcomer_dms
WHERE
    guild_id = $1;

-- name: UpdateWelcomerDMsGuildSettings :execrows
UPDATE
    guild_settings_welcomer_dms
SET
    toggle_enabled = $2,
    toggle_use_text_format = $3,
    toggle_include_image = $4,
    message_format = $5
WHERE
    guild_id = $1;

