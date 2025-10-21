CREATE TABLE IF NOT EXISTS guild_settings_leaver (
    guild_id bigint NOT NULL UNIQUE PRIMARY KEY,
    toggle_enabled boolean NOT NULL,
    channel bigint NOT NULL,
    message_format jsonb NOT NULL,
    auto_delete_leaver_messages boolean NOT NULL,
    leaver_message_lifetime integer NOT NULL,
    FOREIGN KEY (guild_id) REFERENCES guilds (guild_id) ON DELETE CASCADE ON UPDATE CASCADE
);

