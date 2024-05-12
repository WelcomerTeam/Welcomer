CREATE TABLE IF NOT EXISTS guild_settings_welcomer_dms (
    guild_id bigint NOT NULL UNIQUE PRIMARY KEY,
    toggle_enabled boolean NOT NULL,
    toggle_use_text_format boolean NOT NULL,
    toggle_include_image boolean NOT NULL,
    message_format jsonb NOT NULL
);

