CREATE TABLE IF NOT EXISTS guilds (
    guild_id bigint NOT NULL UNIQUE PRIMARY KEY,
    created_at timestamp NOT NULL,
    updated_at timestamp NOT NULL,
    name text NOT NULL,
    embed_colour integer NOT NULL DEFAULT '3553599',
    site_splash_url text NULL,
    site_staff_visible boolean NULL,
    site_guild_visible boolean NULL,
    site_allow_invites boolean NULL
);

