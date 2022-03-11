-- name: CreateWelcomerImagesGuildSettings :one
INSERT INTO guild_settings_Welcomer_images (guild_id)
    VALUES ($1)
RETURNING
    *;

-- name: GetWelcomerImagesGuildSettings :one
SELECT
    *
FROM
    guild_settings_Welcomer_images
WHERE
    guild_id = $1;

-- name: UpdateWelcomerImagesGuildSettings :execrows
UPDATE
    guild_settings_Welcomer_images
SET
    toggle_enabled = $2,
    toggle_image_border = $3,
    background_name = $4,
    colour_text = $5,
    colour_text_border = $6,
    colour_image_border = $7,
    colour_profile_border = $8,
    image_alignment = $9,
    image_theme = $10,
    image_message = $11,
    image_profile_border_type = $12
WHERE
    guild_id = $1;

