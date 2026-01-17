-- name: FetchVoiceChannelEventsForAggregation :many
SELECT 
    guild_id,
    user_id,
    channel_id,
    event_type,
    occurred_at
FROM ingest_voice_channel_events
WHERE occurred_at > $1 AND occurred_at <= $2
ORDER BY user_id, occurred_at;

-- name: GetOpenVoiceChannelSession :one
SELECT 
    guild_id,
    user_id,
    channel_id,
    start_ts,
    last_seen_ts,
    closed_at
FROM guild_voice_channel_open_sessions
WHERE guild_id = $1 AND user_id = $2;

-- name: UpsertOpenVoiceChannelSession :exec
INSERT INTO guild_voice_channel_open_sessions (guild_id, user_id, channel_id, start_ts, last_seen_ts, closed_at)
VALUES ($1, $2, $3, $4, $5, NULL)
ON CONFLICT (guild_id, user_id) DO UPDATE SET
    channel_id = EXCLUDED.channel_id,
    last_seen_ts = EXCLUDED.last_seen_ts,
    closed_at = NULL;

-- name: CloseOpenVoiceChannelSession :exec
UPDATE guild_voice_channel_open_sessions
SET closed_at = $3
WHERE guild_id = $1 AND user_id = $2;

-- name: CreateVoiceChannelStat :exec
INSERT INTO guild_voice_channel_stats (guild_id, channel_id, user_id, start_ts, end_ts, total_time_ms)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: GetAllOpenVoiceChannelSessions :many
SELECT 
    guild_id,
    user_id,
    channel_id,
    start_ts,
    last_seen_ts,
    closed_at
FROM guild_voice_channel_open_sessions;
