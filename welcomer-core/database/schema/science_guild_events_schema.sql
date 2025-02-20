CREATE TABLE IF NOT EXISTS science_guild_events (
    guild_event_uuid uuid NOT NULL UNIQUE PRIMARY KEY,
    guild_id bigint NOT NULL,
    user_id bigint,
    created_at timestamp NOT NULL,
    event_type integer NOT NULL,
    data json
);

ALTER TABLE science_guild_events ALTER COLUMN data SET STORAGE PLAIN;