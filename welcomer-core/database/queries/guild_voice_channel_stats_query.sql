-- name: CreateVoiceChannelStat :exec
INSERT INTO guild_voice_channel_stats (guild_id, channel_id, user_id, start_ts, end_ts, total_time_ms)
VALUES ($1, $2, $3, $4, $5, $6);
