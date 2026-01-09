CREATE TABLE IF NOT EXISTS guilds (
    guild_id bigint NOT NULL UNIQUE PRIMARY KEY,
    embed_colour integer NOT NULL,
    site_splash_url text NOT NULL,
    site_staff_visible boolean NOT NULL,
    site_guild_visible boolean NOT NULL,
    site_allow_invites boolean NOT NULL,
    member_count integer NOT NULL,
    number_locale integer,
    bucket_id smallint NOT NULL,
    bio text NOT NULL DEFAULT ''
);
