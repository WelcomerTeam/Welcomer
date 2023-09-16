-- name: CreateBorderwallGuildSettings :one
INSERT INTO guild_settings_borderwall (guild_id, toggle_enabled, toggle_send_dm, channel, message_verify, message_verified, roles_on_join, roles_on_verify)
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING
    *;

-- name: CreateOrUpdateBorderwallGuildSettings :one
INSERT INTO guild_settings_borderwall (guild_id, toggle_enabled, toggle_send_dm, channel, message_verify, message_verified, roles_on_join, roles_on_verify)
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT(guild_id) DO UPDATE
    SET toggle_enabled = EXCLUDED.toggle_enabled, 
        toggle_send_dm = EXCLUDED.toggle_send_dm, 
        channel = EXCLUDED.channel, 
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
    toggle_send_dm = $3,
    channel = $4,
    message_verify = $5,
    message_verified = $6,
    roles_on_join = $7,
    roles_on_verify = $8
WHERE
    guild_id = $1;

