-- name: CreateGuild :one
INSERT INTO guilds (guild_id, embed_colour, site_splash_url, site_staff_visible, site_guild_visible, site_allow_invites, member_count, number_locale, bucket_id)
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, trunc(random() * 10000)::int)
RETURNING
    *;

-- name: CreateOrUpdateGuild :one
INSERT INTO guilds (guild_id, embed_colour, site_splash_url, site_staff_visible, site_guild_visible, site_allow_invites, member_count, number_locale)
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT(guild_id) DO UPDATE
    SET embed_colour = EXCLUDED.embed_colour,
        site_splash_url = EXCLUDED.site_splash_url,
        site_staff_visible = EXCLUDED.site_staff_visible,
        site_guild_visible = EXCLUDED.site_guild_visible,
        site_allow_invites = EXCLUDED.site_allow_invites,
        member_count = EXCLUDED.member_count,
        number_locale = EXCLUDED.number_locale
RETURNING
    *;

-- name: GetGuild :one
SELECT
    *
FROM
    guilds
WHERE
    guild_id = $1;

-- name: UpdateGuild :one
UPDATE
    guilds
SET
    embed_colour = $2,
    site_splash_url = $3,
    site_staff_visible = $4,
    site_guild_visible = $5,
    site_allow_invites = $6,
    number_locale = $7
WHERE
    guild_id = $1
RETURNING
    *;

-- name: IncrementGuildMemberCount :one
UPDATE
    guilds
SET
    member_count = COALESCE(member_count, @guild_members_default) + @increment
WHERE
    guild_id = $1
RETURNING
    member_count;

-- name: SetGuildMemberCount :execrows
UPDATE
    guilds
SET
    member_count = $2
WHERE
    guild_id = $1;