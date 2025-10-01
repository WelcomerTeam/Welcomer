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

-- name: GetScienceGuildJoinLeaveEventForUser :one
SELECT
    *
FROM
    science_guild_events
    LEFT JOIN guild_invites ON guild_invites.invite_code = science_guild_events.data ->> 'invite_code'
WHERE
    science_guild_events.event_type IN ($1, $2)
    AND science_guild_events.guild_id = $3
    AND science_guild_events.user_id = $4
ORDER BY
    science_guild_events.created_at DESC
LIMIT 1;

-- name: GetExpiredWelcomeMessageEvents :many
SELECT
    science_guild_events.guild_id,
    science_guild_events.user_id,
    CAST(science_guild_events.data ->> 'channel_id' AS BIGINT) AS channel_id,
    CAST(science_guild_events.data ->> 'message_id' AS BIGINT) AS message_id
FROM 
    science_guild_events
    LEFT JOIN
        science_guild_events message_deleted
    ON
        message_deleted.guild_id = science_guild_events.guild_id
        AND message_deleted.user_id = science_guild_events.user_id
        AND message_deleted.event_type = @science_guild_event_type_welcome_message_removed
        AND message_deleted.data ->> 'message_id' = science_guild_events.data ->> 'message_id'
WHERE
    science_guild_events.guild_id = @guild_id
    AND science_guild_events.event_type = @science_guild_event_type_user_welcomed
    AND science_guild_events.data ->> 'message_id' IS NOT NULL
    AND science_guild_events.created_at < @welcome_message_lifetime
    AND message_deleted.guild_event_uuid IS NULL
LIMIT @event_limit;