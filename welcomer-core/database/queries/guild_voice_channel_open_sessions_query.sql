-- name: CreateGuildVoiceChannelOpenSession :exec
INSERT INTO guild_voice_channel_open_sessions (guild_id, user_id, channel_id, start_ts, last_seen_ts)
VALUES ($1, $2, $3, $4, $5);

-- name: DeleteAndGetGuildVoiceChannelOpenSession :one
DELETE FROM guild_voice_channel_open_sessions
WHERE guild_id = $1 AND user_id = $2
RETURNING guild_id, user_id, channel_id, start_ts, last_seen_ts;

-- name: UpdateGuildVoiceChannelOpenSessionLastSeen :exec
UPDATE guild_voice_channel_open_sessions
SET last_seen_ts = $3
WHERE guild_id = $1 AND user_id = $2;