CREATE TABLE IF NOT EXISTS guild_settings_reaction_roles (
    reaction_role_id uuid NOT NULL UNIQUE PRIMARY KEY,
    guild_id bigint NOT NULL,
    toggle_enabled boolean NOT NULL,
    channel_id bigint NOT NULL,
    message_id bigint,
    is_system_message boolean NOT NULL,
    system_message_format jsonb,
    reaction_role_type int NOT NULL,
    roles jsonb NOT NULL,

    FOREIGN KEY (guild_id) REFERENCES guilds (guild_id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE INDEX IF NOT EXISTS guild_settings_reaction_roles_guild_id ON guild_settings_reaction_roles (guild_id)
