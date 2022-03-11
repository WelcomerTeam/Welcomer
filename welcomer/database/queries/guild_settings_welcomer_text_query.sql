-- name: CreateWelcomerTextGuildSettings :one
INSERT INTO guild_settings_welcomer_text (guild_id)
    VALUES ($1)
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

