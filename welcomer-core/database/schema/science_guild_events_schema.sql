CREATE TABLE IF NOT EXISTS science_guild_events (
    guild_event_uuid uuid NOT NULL UNIQUE PRIMARY KEY,
    guild_id bigint NOT NULL,
    created_at timestamp NOT NULL,
    event_type integer NOT NULL,
    data jsonb NOT NULL,
    FOREIGN KEY (guild_id) REFERENCES guilds (guild_id)
);

