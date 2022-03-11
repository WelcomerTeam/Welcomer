CREATE TABLE IF NOT EXISTS guild_settings_welcomer_text (
    guild_id bigint NOT NULL UNIQUE PRIMARY KEY,
    toggle_enabled boolean DEFAULT 'false',
    channel bigint NULL,
    message_format jsonb NULL,
    FOREIGN KEY (guild_id) REFERENCES guilds (guild_id) ON DELETE CASCADE ON UPDATE CASCADE
);

