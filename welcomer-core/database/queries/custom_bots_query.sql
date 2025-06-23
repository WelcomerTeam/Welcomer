-- name: CreateCustomBot :one
INSERT INTO custom_bots (custom_bot_uuid, guild_id, token, created_at, is_active, application_id, application_name, application_avatar)
VALUES ($1, $2, $3, now(), $4, $5, $6, $7)
RETURNING
    *;

-- name: GetCustomBotByGuildId :one
SELECT
    *
FROM
    custom_bots
WHERE
    guild_id = $1;

-- name: GetCustomBotByApplicationId :one
SELECT
    *
FROM
    custom_bots
WHERE
    application_id = $1;

-- name: UpdateCustomBotToken :one
UPDATE
    custom_bots
SET
    token = $2,
    is_active = $3,
    application_id = $4,
    application_name = $5,
    application_avatar = $6
WHERE
    custom_bot_uuid = $1
RETURNING
    *;

-- name: UpdateCustomBotApplication :one
UPDATE
    custom_bots
SET
    application_id = $2,
    application_name = $3,
    application_avatar = $4
WHERE
    custom_bot_uuid = $1
RETURNING
    *;

-- name: DeleteCustomBot :execrows
DELETE FROM
    custom_bots
WHERE
    custom_bot_uuid = $1;