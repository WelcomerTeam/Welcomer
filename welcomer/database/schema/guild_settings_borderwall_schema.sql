CREATE TABLE IF NOT EXISTS guild_settings_borderwall (
    guild_id bigint NOT NULL UNIQUE PRIMARY KEY,
    toggle_enabled boolean DEFAULT 'false',
    message_verify jsonb NULL,
    message_verified jsonb NULL,
    roles_on_join bigint[],
    roles_on_verify bigint[],
    FOREIGN KEY (guild_id) REFERENCES guilds (guild_id) ON DELETE CASCADE ON UPDATE CASCADE
);

