-- name: CreateBorderwallGuildSettings :one
INSERT INTO guild_settings_borderwall (guild_id, toggle_enabled, message_verify, message_verified, roles_on_join, roles_on_verify)
    VALUES ($1, $2, $3, $4, $5, $6)
RETURNING
    *;

-- name: CreateOrUpdateBorderwallGuildSettings :one
INSERT INTO guild_settings_borderwall (guild_id, toggle_enabled, message_verify, message_verified, roles_on_join, roles_on_verify)
    VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT(guild_id) DO UPDATE
    SET toggle_enabled = EXCLUDED.toggle_enabled,
        message_verify = EXCLUDED.message_verify,
        message_verified = EXCLUDED.message_verified,
        roles_on_join = EXCLUDED.roles_on_join,
        roles_on_verify = EXCLUDED.roles_on_verify
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

