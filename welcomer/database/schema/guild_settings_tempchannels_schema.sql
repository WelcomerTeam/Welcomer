CREATE TABLE IF NOT EXISTS guild_settings_tempchannels (
    guild_id bigint NOT NULL UNIQUE PRIMARY KEY,
    toggle_enabled boolean DEFAULT 'false',
    toggle_autopurge boolean DEFAULT 'true',
    channel_lobby bigint NULL,
    channel_category bigint NULL,
    default_user_count integer NULL,
    FOREIGN KEY (guild_id) REFERENCES guilds (guild_id) ON DELETE CASCADE ON UPDATE CASCADE
);

