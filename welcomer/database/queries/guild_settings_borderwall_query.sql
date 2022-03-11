-- name: CreateBorderwallGuildSettings :one
INSERT INTO guild_settings_borderwall (guild_id)
    VALUES ($1)
RETURNING
    *;

-- name: GetBorderwallGuildSettings :one
SELECT
    *
FROM
    guild_settings_borderwall
WHERE
    guild_id = $1;

-- name: UpdateBorderwallGuildSettings :execrows
UPDATE
    guild_settings_borderwall
SET
    toggle_enabled = $2,
    message_verify = $3,
    message_verified = $4,
    roles_on_join = $5,
    roles_on_verify = $6
WHERE
    guild_id = $1;

