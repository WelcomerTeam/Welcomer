CREATE TABLE IF NOT EXISTS science_guild_events (
    guild_event_uuid uuid NOT NULL,
    guild_id bigint NOT NULL,
    user_id bigint,
    created_at timestamp NOT NULL,
    event_type integer NOT NULL,
    data json,
    PRIMARY KEY (guild_event_uuid, created_at)
) PARTITION BY RANGE (created_at);

ALTER TABLE science_guild_events ALTER COLUMN data SET STORAGE PLAIN;