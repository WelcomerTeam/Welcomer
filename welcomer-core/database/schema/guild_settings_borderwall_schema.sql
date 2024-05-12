CREATE TABLE IF NOT EXISTS guild_settings_borderwall (
    guild_id bigint NOT NULL UNIQUE PRIMARY KEY,
    toggle_enabled boolean NOT NULL,
    toggle_send_dm boolean NOT NULL,
    channel bigint NOT NULL,
    message_verify jsonb NOT NULL,
    message_verified jsonb NOT NULL,
    roles_on_join bigint[] NOT NULL,
    roles_on_verify bigint[] NOT NULL
);

