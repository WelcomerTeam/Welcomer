-- name: CreateManyIngestMessageEvents :copyfrom
INSERT INTO ingest_message_events (message_id, guild_id, channel_id, user_id, event_type, occurred_at)
    VALUES ($1, $2, $3, $4, $5, $6);