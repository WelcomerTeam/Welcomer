-- name: CreateWelcomerImagesGuildSettings :one
INSERT INTO guild_settings_Welcomer_images (guild_id, toggle_enabled, toggle_image_border, background_name, colour_text, colour_text_border, colour_image_border, colour_profile_border, image_alignment, image_theme, image_message, image_profile_border_type)
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING
    *;

-- name: CreateOrUpdateWelcomerImagesGuildSettings :one
INSERT INTO guild_settings_Welcomer_images (guild_id, toggle_enabled, toggle_image_border, background_name, colour_text, colour_text_border, colour_image_border, colour_profile_border, image_alignment, image_theme, image_message, image_profile_border_type)
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
ON CONFLICT(guild_id) DO UPDATE
    SET toggle_enabled = EXCLUDED.toggle_enabled,
        toggle_image_border = EXCLUDED.toggle_image_border,
        background_name = EXCLUDED.background_name,
        colour_text = EXCLUDED.colour_text,
        colour_text_border = EXCLUDED.colour_text_border,
        colour_image_border = EXCLUDED.colour_image_border,
        colour_profile_border = EXCLUDED.colour_profile_border,
        image_alignment = EXCLUDED.image_alignment,
        image_theme = EXCLUDED.image_theme,
        image_message = EXCLUDED.image_message,
        image_profile_border_type = EXCLUDED.image_profile_border_type
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
