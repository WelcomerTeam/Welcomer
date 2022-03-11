CREATE TABLE IF NOT EXISTS guild_settings_moderation (
    guild_id bigint NOT NULL UNIQUE PRIMARY KEY,
    toggle_reasons_required boolean DEFAULT 'false',
    FOREIGN KEY (guild_id) REFERENCES guilds (guild_id) ON DELETE CASCADE ON UPDATE CASCADE
);

