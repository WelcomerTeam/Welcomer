CREATE TABLE IF NOT EXISTS guild_punishments (
    punishment_uuid uuid NOT NULL UNIQUE PRIMARY KEY,
    created_at timestamp NOT NULL,
    updated_at timestamp NOT NULL,
    guild_id bigint NOT NULL,
    punishment_type smallint NOT NULL,
    guild_moderation_event_uuid uuid NULL,
    user_id bigint NOT NULL,
    moderator_id bigint NOT NULL,
    data jsonb NULL
);

CREATE INDEX IF NOT EXISTS guild_punishments_guild_id_user_id ON guild_punishments (guild_id, user_id);

