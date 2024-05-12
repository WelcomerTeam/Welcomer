CREATE TABLE IF NOT EXISTS guild_settings_tempchannels (
    guild_id bigint NOT NULL UNIQUE PRIMARY KEY,
    toggle_enabled boolean NOT NULL,
    toggle_autopurge boolean NOT NULL,
    channel_lobby bigint NOT NULL,
    channel_category bigint NOT NULL,
    default_user_count integer NOT NULL
);

