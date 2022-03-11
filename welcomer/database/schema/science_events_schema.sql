CREATE TABLE IF NOT EXISTS science_events (
    event_uuid uuid NOT NULL UNIQUE PRIMARY KEY,
    created_at timestamp NOT NULL,
    event_type integer NOT NULL,
    data jsonb NULL
);

