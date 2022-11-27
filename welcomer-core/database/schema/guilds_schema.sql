CREATE TABLE IF NOT EXISTS guilds (
    guild_id bigint NOT NULL UNIQUE PRIMARY KEY,
    created_at timestamp NOT NULL,
    updated_at timestamp NOT NULL,
    name text NOT NULL,
    embed_colour integer NOT NULL,
    site_splash_url text NOT NULL,
    site_staff_visible boolean NOT NULL,
    site_guild_visible boolean NOT NULL,
    site_allow_invites boolean NOT NULL
);

