CREATE TABLE ingest_voice_channel_events (
    event_id BIGINT GENERATED ALWAYS AS IDENTITY,
    guild_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    channel_id BIGINT,
    event_type SMALLINT NOT NULL,
    occurred_at TIMESTAMP WITH TIME ZONE NOT NULL,
    PRIMARY KEY (event_id, occurred_at)
);
