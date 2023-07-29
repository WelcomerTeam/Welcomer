CREATE TABLE IF NOT EXISTS guild_settings_rules (
    guild_id bigint NOT NULL UNIQUE PRIMARY KEY,
    toggle_enabled boolean NOT NULL,
    toggle_dms_enabled boolean NOT NULL,
    rules text[],
    FOREIGN KEY (guild_id) REFERENCES guilds (guild_id) ON DELETE CASCADE ON UPDATE CASCADE
);

