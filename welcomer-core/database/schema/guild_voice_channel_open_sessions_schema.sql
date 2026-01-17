CREATE TABLE IF NOT EXISTS guild_voice_channel_open_sessions (
    guild_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    channel_id BIGINT NOT NULL,
    start_ts TIMESTAMP WITH TIME ZONE NOT NULL,
    last_seen_ts TIMESTAMP WITH TIME ZONE NOT NULL,
    closed_at TIMESTAMP WITH TIME ZONE,
    PRIMARY KEY (guild_id, user_id)
);

CREATE INDEX IF NOT EXISTS guild_voice_channel_open_sessions_guild_id ON guild_voice_channel_open_sessions (guild_id);
CREATE INDEX IF NOT EXISTS guild_voice_channel_open_sessions_closed_at ON guild_voice_channel_open_sessions (closed_at);
