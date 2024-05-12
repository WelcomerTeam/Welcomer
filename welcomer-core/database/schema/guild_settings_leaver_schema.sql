CREATE TABLE IF NOT EXISTS guild_settings_leaver (
    guild_id bigint NOT NULL UNIQUE PRIMARY KEY,
    toggle_enabled boolean NOT NULL,
    channel bigint NOT NULL,
    message_format jsonb NOT NULL
);

