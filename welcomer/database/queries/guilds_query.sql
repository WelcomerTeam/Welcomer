-- name: CreateGuild :one
INSERT INTO guilds (guild_id, created_at, updated_at, name, embed_colour, site_splash_url, site_staff_visible, site_guild_visible, site_allow_invites)
    VALUES ($1, now(), now(), $2, '3553599', NULL, NULL, NULL, NULL)
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

