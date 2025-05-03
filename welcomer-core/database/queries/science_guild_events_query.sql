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
