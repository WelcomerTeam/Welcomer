-- name: CreateWelcomerBackground :one
INSERT INTO guild_settings_welcomer_backgrounds (image_uuid, created_at, guild_id, filename, filesize, filetype)
    VALUES (uuid_generate_v4(), now(), $1, $2, $3, $4)
RETURNING
    *;

-- name: GetWelcomerBackground :one
SELECT
    *
FROM
    guild_settings_welcomer_backgrounds
WHERE
    image_uuid = $1;

-- name: GetWelcomerBackgroundByGuildID :many
SELECT
    *
FROM
    guild_settings_welcomer_backgrounds
WHERE
    guild_id = $1;

-- name: DeleteWelcomerBackground :execrows
DELETE FROM guild_settings_welcomer_backgrounds
WHERE image_uuid = $1;