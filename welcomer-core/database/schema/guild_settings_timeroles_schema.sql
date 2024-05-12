CREATE TABLE IF NOT EXISTS guild_settings_timeroles (
    guild_id bigint NOT NULL UNIQUE PRIMARY KEY,
    toggle_enabled boolean NOT NULL,
    timeroles jsonb NOT NULL
);