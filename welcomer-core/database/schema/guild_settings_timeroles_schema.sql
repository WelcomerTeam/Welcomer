CREATE TYPE guild_settings_timeroles_role AS (
    role_id bigint
    seconds bigint
)

CREATE TABLE IF NOT EXISTS guild_settings_timeroles (
    guild_id bigint NOT NULL UNIQUE PRIMARY KEY,
    toggle_enabled boolean NOT NULL,
    timeroles guild_settings_timeroles_role[],
    FOREIGN KEY (guild_id) REFERENCES guilds (guild_id) ON DELETE CASCADE ON UPDATE CASCADE
);