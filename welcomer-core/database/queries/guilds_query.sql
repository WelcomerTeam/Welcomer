-- name: CreateGuild :one
INSERT INTO guilds (guild_id, embed_colour, site_splash_url, site_staff_visible, site_guild_visible, site_allow_invites)
    VALUES ($1, $2, $3, $4, $5, $6)
RETURNING
    *;

-- name: CreateOrUpdateGuild :one
INSERT INTO guilds (guild_id, embed_colour, site_splash_url, site_staff_visible, site_guild_visible, site_allow_invites)
    VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT(guild_id) DO UPDATE
    SET embed_colour = EXCLUDED.embed_colour,
        site_splash_url = EXCLUDED.site_splash_url,
        site_staff_visible = EXCLUDED.site_staff_visible,
        site_guild_visible = EXCLUDED.site_guild_visible,
        site_allow_invites = EXCLUDED.site_allow_invites
RETURNING
    *;

-- name: GetGuild :one
SELECT
    *
FROM
    guilds
WHERE
    guild_id = $1;

-- name: UpdateGuild :execrows
UPDATE
    guilds
SET
    embed_colour = $2,
    site_splash_url = $3,
    site_staff_visible = $4,
    site_guild_visible = $5,
    site_allow_invites = $6
WHERE
    guild_id = $1;

