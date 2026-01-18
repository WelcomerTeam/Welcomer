CREATE TABLE IF NOT EXISTS guild_voice_channel_stats (
    stat_id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    guild_id BIGINT NOT NULL,
    channel_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    start_ts TIMESTAMP WITH TIME ZONE NOT NULL,
    end_ts TIMESTAMP WITH TIME ZONE NOT NULL,
    total_time_ms BIGINT NOT NULL,
    inferred BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX IF NOT EXISTS guild_voice_channel_stats_guild_id ON guild_voice_channel_stats (guild_id);
