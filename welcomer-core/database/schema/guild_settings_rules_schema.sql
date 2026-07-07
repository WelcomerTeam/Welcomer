CREATE TABLE IF NOT EXISTS guild_settings_rules (
    guild_id bigint NOT NULL UNIQUE PRIMARY KEY,
    toggle_enabled boolean NOT NULL,
    toggle_dms_enabled boolean NOT NULL,
    rules text[],
    moderation_checkup_uuid uuid,
    FOREIGN KEY (moderation_checkup_uuid) REFERENCES moderation_checkup (checkup_uuid) ON DELETE SET NULL ON UPDATE CASCADE,
    FOREIGN KEY (guild_id) REFERENCES guilds (guild_id) ON DELETE CASCADE ON UPDATE CASCADE
);

