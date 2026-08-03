CREATE TABLE IF NOT EXISTS guild_voice_channel_stats (
    guild_id BIGINT NOT NULL,
    channel_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    start_ts TIMESTAMP WITH TIME ZONE NOT NULL,
    end_ts TIMESTAMP WITH TIME ZONE NOT NULL,
    total_time_ms BIGINT NOT NULL,
    inferred BOOLEAN NOT NULL DEFAULT FALSE,
    PRIMARY KEY (start_ts, guild_id, channel_id, user_id)
) PARTITION BY RANGE (start_ts);

CREATE INDEX IF NOT EXISTS guild_voice_channel_stats_guild_id ON guild_voice_channel_stats (guild_id);
