CREATE TABLE IF NOT EXISTS guild_settings_welcomer_dms (
    guild_id bigint NOT NULL UNIQUE PRIMARY KEY,
    toggle_enabled boolean DEFAULT 'false',
    toggle_use_text_format boolean DEFAULT 'false',
    toggle_include_image boolean DEFAULT 'false',
    message_format jsonb NULL,
    FOREIGN KEY (guild_id) REFERENCES guilds (guild_id) ON DELETE CASCADE ON UPDATE CASCADE
);

