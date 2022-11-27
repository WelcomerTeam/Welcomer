-- name: CreateGuild :one
INSERT INTO guilds (guild_id, created_at, updated_at, name, embed_colour, site_splash_url, site_staff_visible, site_guild_visible, site_allow_invites)
    VALUES ($1, now(), now(), $2, $3, $4, $5, $6, $7)
RETURNING
    *;

-- name: CreateOrUpdateGuild :one
INSERT INTO guilds (guild_id, created_at, updated_at, name, embed_colour, site_splash_url, site_staff_visible, site_guild_visible, site_allow_invites)
    VALUES ($1, now(), now(), $2, $3, $4, $5, $6, $7)
ON CONFLICT(guild_id) DO UPDATE
    SET name = EXCLUDED.name,
        embed_colour = EXCLUDED.embed_colour,
        site_splash_url = EXCLUDED.site_splash_url,
        site_staff_visible = EXCLUDED.site_staff_visible,
        site_guild_visible = EXCLUDED.site_guild_visible,
        site_allow_invites = EXCLUDED.site_allow_invites,
        updated_at = now()
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
    name = $2,
    embed_colour = $3,
    site_splash_url = $4,
    site_staff_visible = $5,
    site_guild_visible = $6,
    site_allow_invites = $7,
    updated_at = now()
WHERE
    guild_id = $1;

