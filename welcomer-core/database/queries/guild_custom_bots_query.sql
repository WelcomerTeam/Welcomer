-- name: CreateCustomBot :one
INSERT INTO custom_bots (custom_bot_uuid, guild_id, public_key, token, created_at, is_active, application_id, application_name, application_avatar)
VALUES ($1, $2, $3, $4, now(), $5, $6, $7, $8)
RETURNING
    *;

-- name: GetCustomBotsByGuildId :many
SELECT
    custom_bot_uuid,
    guild_id,
    public_key,
    created_at,
    is_active,
    application_id,
    application_name,
    application_avatar
FROM
    custom_bots
WHERE
    guild_id = $1;

-- name: UpdateCustomBotToken :one
UPDATE
    custom_bots
SET
    public_key = $2,
    token = $3,
    is_active = $4,
    application_id = $5,
    application_name = $6,
    application_avatar = $7
WHERE
    custom_bot_uuid = $1
RETURNING
    *;

-- name: UpdateCustomBot :one
UPDATE
    custom_bots
SET
    public_key = $2,
    is_active = $3,
    application_id = $4,
    application_name = $5,
    application_avatar = $6
WHERE
    custom_bot_uuid = $1
RETURNING
    *;

-- name: DeleteCustomBot :execrows
DELETE FROM
    custom_bots
WHERE
    custom_bot_uuid = $1;

-- name: GetCustomBotByIdWithToken :one
SELECT
    *
FROM
    custom_bots
WHERE
    custom_bot_uuid = $1;

-- name: GetAllCustomBotsWithToken :many
SELECT
    *
FROM
    custom_bots
WHERE
    is_active = true
    AND token IS NOT NULL
    AND token != '';