-- name: CreateGuildWelcomerImage :one
INSERT INTO guild_welcomer_images (image_uuid, guild_id, created_at, image_format)
    VALUES (uuid_generate_v4 (), $1, now(), $2)
RETURNING
    *;

-- name: GetGuildWelcomerImage :one
SELECT
    *
FROM
    guild_welcomer_images
WHERE
    image_uuid = $1;

-- name: DeleteGuildWelcomerImage :execrows
DELETE FROM guild_welcomer_images
WHERE image_uuid = $1;

