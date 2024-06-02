CREATE TABLE IF NOT EXISTS science_command_usages (
    command_uuid uuid NOT NULL UNIQUE PRIMARY KEY,
    created_at timestamp NOT NULL,
    updated_at timestamp NOT NULL,
    guild_id bigint NOT NULL,
    user_id bigint NOT NULL,
    channel_id bigint NULL,
    command text NOT NULL,
    errored boolean NOT NULL,
    execution_time_ms bigint NOT NULL,
    FOREIGN KEY (guild_id) REFERENCES guilds (guild_id)
);

CREATE INDEX IF NOT EXISTS science_command_usages_user_id ON science_command_usages (user_id);

CREATE INDEX IF NOT EXISTS science_command_usages_guild_id ON science_command_usages (guild_id);

CREATE INDEX IF NOT EXISTS science_command_usages_command ON science_command_usages (command);

