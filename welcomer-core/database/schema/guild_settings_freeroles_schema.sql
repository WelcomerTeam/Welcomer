CREATE TABLE IF NOT EXISTS guild_settings_freeroles (
    guild_id bigint NOT NULL UNIQUE PRIMARY KEY,
    toggle_enabled boolean NOT NULL,
    roles bigint[] NOT NULL
);

