-- name: CreateWelcomerImages :one
INSERT INTO utils.images (utils.image_uuid, guild_id, created_at, image_type, data)
    VALUES ($1, $2, $3, $4, $5)
RETURNING
    *;

-- name: GetWelcomerImages :one
SELECT
    *
FROM
    utils.images
WHERE
    utils.image_uuid = $1;

-- name: GetWelcomerImagesByGuildId :many
SELECT
    *
FROM
    utils.images
WHERE
    guild_id = $1;

-- name: DeleteWelcomerImage :execrows
DELETE FROM utils.images
WHERE utils.image_uuid = $1;
