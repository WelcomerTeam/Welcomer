-- name: CreateManyIngestVoiceChannelEvents :copyfrom
INSERT INTO ingest_voice_channel_events (guild_id, user_id, channel_id, event_type, occurred_at)
    VALUES ($1, $2, $3, $4, $5);