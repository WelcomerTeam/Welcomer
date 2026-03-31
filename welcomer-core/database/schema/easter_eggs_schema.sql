CREATE TABLE IF NOT EXISTS easter_eggs (
    easter_egg_uuid uuid NOT NULL UNIQUE PRIMARY KEY,
    guild_id BIGINT NOT NULL,
    channel_id BIGINT NOT NULL,
    message_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    claimed_egg TEXT NOT NULL,
    created_at timestamp NOT NULL,
    wm_user_id TEXT NOT NULL,
    UNIQUE (guild_id, channel_id, message_id)
);

CREATE INDEX IF NOT EXISTS easter_eggs_user_id ON easter_eggs (user_id);