CREATE TABLE IF NOT EXISTS science_guild_moderation_events (
    moderation_event_id uuid NOT NULL UNIQUE PRIMARY KEY,
    guild_id bigint NOT NULL,
    created_at timestamp NOT NULL,
    event_type integer NOT NULL,
    user_id bigint NOT NULL,
    moderator_id bigint NOT NULL,
    data jsonb NULL
);

