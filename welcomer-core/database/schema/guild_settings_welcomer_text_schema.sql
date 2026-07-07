CREATE TABLE IF NOT EXISTS guild_settings_welcomer_text (
    guild_id bigint NOT NULL UNIQUE PRIMARY KEY,
    toggle_enabled boolean NOT NULL,
    channel bigint NOT NULL,
    message_format jsonb NOT NULL,
    moderation_checkup_uuid uuid,
    FOREIGN KEY (moderation_checkup_uuid) REFERENCES moderation_checkup (checkup_uuid) ON DELETE SET NULL ON UPDATE CASCADE,
    FOREIGN KEY (guild_id) REFERENCES guilds (guild_id) ON DELETE CASCADE ON UPDATE CASCADE
);

