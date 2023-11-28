-- name: CreateWelcomerImages :one
INSERT INTO welcomer_images (welcomer_image_uuid, guild_id, created_at, image_type, data)
    VALUES ($1, $2, $3, $4, $5)
RETURNING
    *;

-- name: GetWelcomerImages :one
SELECT
    *
FROM
    welcomer_images
WHERE
    welcomer_image_uuid = $1;

-- name: GetWelcomerImagesByGuildId :many
SELECT
    *
FROM
    welcomer_images
WHERE
    guild_id = $1;

-- name: DeleteWelcomerImage :execrows
DELETE FROM welcomer_images
WHERE welcomer_image_uuid = $1;
