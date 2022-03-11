-- name: CreateScienceGuildModerationEvent :one
INSERT INTO science_guild_moderation_events (moderation_event_id, guild_id, created_at, event_type, user_id, moderator_id, data)
    VALUES (uuid_generate_v4 (), $1, now(), $2, $3, $4, $5)
RETURNING
    *;

-- name: GetScienceGuildModerationEvent :one
SELECT
    *
FROM
    science_guild_moderation_events
WHERE
    moderation_event_id = $1;

