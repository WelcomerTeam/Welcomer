CREATE TABLE IF NOT EXISTS guild_message_counts_hour (
    hour_ts TIMESTAMPTZ NOT NULL,
    guild_id BIGINT NOT NULL,
    channel_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    message_count INTEGER NOT NULL,
    min_ts TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (hour_ts, guild_id, channel_id, user_id)
) PARTITION BY RANGE (hour_ts);

CREATE INDEX IF NOT EXISTS guild_message_counts_hour_guild_hour_user_idx ON guild_message_counts_hour (guild_id, hour_ts) INCLUDE (user_id);
