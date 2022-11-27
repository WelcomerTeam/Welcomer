-- name: CreateScienceGuildEvent :one
INSERT INTO science_guild_events (guild_event_uuid, guild_id, created_at, event_type, data)
    VALUES (uuid_generate_v4(), $1, now(), $2, $3)
RETURNING
    *;

-- name: GetScienceGuildEvent :one
SELECT
    *
FROM
    science_guild_events
WHERE
    guild_event_uuid = $1;

