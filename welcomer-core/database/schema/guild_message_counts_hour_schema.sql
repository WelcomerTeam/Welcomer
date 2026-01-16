CREATE TABLE IF NOT EXISTS guild_message_counts_hour (
    hour_ts TIMESTAMPTZ NOT NULL,
    guild_id BIGINT NOT NULL,
    channel_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    message_count INTEGER NOT NULL,
    PRIMARY KEY (hour_ts, guild_id, channel_id, user_id)
);

CREATE INDEX IF NOT EXISTS guild_message_counts_hour_guild_id ON guild_message_counts_hour (guild_id);
CREATE INDEX IF NOT EXISTS guild_message_counts_hour_channel_id ON guild_message_counts_hour (channel_id);