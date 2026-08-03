CREATE TABLE IF NOT EXISTS moderation_checkup_queue (
    checkup_queue_uuid uuid NOT NULL UNIQUE PRIMARY KEY,
    guild_id bigint NOT NULL,
    user_id bigint NOT NULL,
    data_type INT NOT NULL,
    created_at timestamp NOT NULL,
    value TEXT NOT NULL
);