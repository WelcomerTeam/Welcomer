-- name: CreateScienceGuildEvent :one
INSERT INTO science_guild_events (guild_event_uuid, guild_id, user_id, created_at, event_type, data)
    VALUES (uuid_generate_v7(), $1, $2, now(), $3, $4)
RETURNING
    *;

-- name: CreateManyScienceGuildEvents :copyfrom
INSERT INTO science_guild_events (guild_event_uuid, guild_id, user_id, created_at, event_type, data)
    VALUES ($1, $2, $3, $4, $5, $6);

-- name: GetScienceGuildEvent :one
SELECT
    *
FROM
    science_guild_events
WHERE
    guild_event_uuid = $1;

