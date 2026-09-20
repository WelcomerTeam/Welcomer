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
        AND message_deleted.created_at > @welcome_message_lifetime_lookback
WHERE
    science_guild_events.guild_id = @guild_id
    AND science_guild_events.event_type = @science_guild_event_type_user_welcomed
    AND science_guild_events.data ->> 'message_id' IS NOT NULL
    AND science_guild_events.created_at < @welcome_message_lifetime
    AND science_guild_events.created_at > @welcome_message_lifetime_lookback
    AND message_deleted.guild_event_uuid IS NULL
LIMIT @event_limit;

-- name: SumScienceGuildEventsForGuild :one
SELECT
    COUNT(*)::INT AS event_count
FROM
    science_guild_events
WHERE
    science_guild_events.guild_id = @guild_id
    AND science_guild_events.event_type = @event_type
    AND science_guild_events.created_at BETWEEN @date_from AND @date_to;

-- name: GetScienceGuildEventsForGuild :many
SELECT DISTINCT
    science_guild_events.user_id,
    science_guild_events.event_type,
    science_guild_events.created_at
FROM
    science_guild_events
WHERE
    science_guild_events.guild_id = @guild_id
    AND science_guild_events.event_type = @event_type
    AND science_guild_events.created_at BETWEEN @date_from AND @date_to;

-- name: GetScienceGuildEventsForGuildGroupedByPeriod :many
SELECT
    date_trunc(@period, science_guild_events.created_at)::TIMESTAMP AS date,
    COUNT(*)::INT AS event_count
FROM
    science_guild_events
WHERE
    science_guild_events.guild_id = @guild_id
    AND science_guild_events.event_type = @event_type
    AND science_guild_events.created_at BETWEEN @date_from AND @date_to
GROUP BY
    date_trunc(@period, science_guild_events.created_at)
ORDER BY
    date_trunc(@period, science_guild_events.created_at) ASC;

-- name: GetGuildMemberRetention :one
WITH joined AS (
    SELECT DISTINCT ON (user_id)
        user_id,
        created_at AS joined_at
    FROM science_guild_events
    WHERE guild_id = @guild_id
        AND science_guild_events.event_type = @science_guild_event_type_user_join
        AND science_guild_events.created_at BETWEEN @date_from AND @date_to
    ORDER BY user_id, created_at DESC
)

SELECT
    COUNT(*)::int AS members_joined,
    COUNT(*) FILTER (
        WHERE EXISTS (
            SELECT 1
            FROM science_guild_events leave_event
            WHERE leave_event.guild_id = @guild_id
                AND leave_event.user_id = joined.user_id
                AND leave_event.event_type = @science_guild_event_type_user_leave
                AND leave_event.created_at > joined.joined_at
        )
    )::int AS members_left
FROM joined;

-- name: GetGuildMemberRetentionMatrix :many
-- Returns, for members who joined within [date_to - 1 year, date_to], how many months ago they
-- joined (X axis) and how many months ago they left (Y axis), 0 being the current month.
-- Members who have not left are included with a NULL left_months_ago.
WITH joined AS (
    SELECT DISTINCT ON (user_id)
        user_id,
        created_at AS joined_at
    FROM science_guild_events
    WHERE science_guild_events.guild_id = @guild_id
        AND science_guild_events.event_type = @science_guild_event_type_user_join
        AND science_guild_events.created_at BETWEEN ((@date_to)::timestamp - interval '1 year') AND @date_to
    ORDER BY user_id, created_at DESC
),
left_events AS (
    SELECT DISTINCT ON (user_id)
        user_id,
        created_at AS left_at
    FROM science_guild_events
    WHERE science_guild_events.guild_id = @guild_id
        AND science_guild_events.event_type = @science_guild_event_type_user_leave
        AND science_guild_events.created_at BETWEEN ((@date_to)::timestamp - interval '1 year') AND @date_to
    ORDER BY user_id, created_at DESC
),
cohorts AS (
    SELECT
        (DATE_PART('year', age(@date_to::timestamp, joined.joined_at)) * 12 + DATE_PART('month', age(@date_to::timestamp, joined.joined_at)))::int AS joined_months_ago,
        (DATE_PART('year', age(@date_to::timestamp, left_events.left_at)) * 12 + DATE_PART('month', age(@date_to::timestamp, left_events.left_at)))::int AS left_months_ago
    FROM joined
        LEFT JOIN left_events ON left_events.user_id = joined.user_id
            AND left_events.left_at > joined.joined_at
)

SELECT
    joined_months_ago,
    COALESCE(left_months_ago, -1) AS left_months_ago,
    COUNT(*)::int AS member_count
FROM cohorts
WHERE joined_months_ago BETWEEN 0 AND 12
    AND (left_months_ago IS NULL OR left_months_ago BETWEEN 0 AND 12)
GROUP BY joined_months_ago, left_months_ago
ORDER BY joined_months_ago, left_months_ago;

-- name: GetLeftGuildMemberDaysOnServer :many
WITH joined AS (
    SELECT DISTINCT ON (user_id)
        user_id,
        created_at AS joined_at
    FROM science_guild_events
    WHERE science_guild_events.guild_id = @guild_id
        AND science_guild_events.event_type = @science_guild_event_type_user_join
        AND science_guild_events.created_at BETWEEN @date_from AND @date_to
    ORDER BY science_guild_events.user_id, created_at DESC
),
left_events AS (
    SELECT DISTINCT ON (user_id)
        user_id,
        created_at AS left_at
    FROM science_guild_events
    WHERE science_guild_events.guild_id = @guild_id
        AND science_guild_events.event_type = @science_guild_event_type_user_leave
        AND science_guild_events.created_at >= @date_from
    ORDER BY science_guild_events.user_id, created_at DESC
)

SELECT
    COUNT(*)::int AS count,
    CASE WHEN left_events.left_at IS NULL THEN
        (EXTRACT(EPOCH FROM now() - joined.joined_at) / 86400)
    ELSE
        EXTRACT(EPOCH FROM COALESCE(left_events.left_at, now()) - joined.joined_at) / 86400
    END::int AS days_on_server
FROM joined
    LEFT JOIN left_events ON left_events.user_id = joined.user_id
GROUP BY days_on_server
ORDER BY days_on_server DESC;

-- name: GetGuildMemberJoins :many
SELECT
    date_trunc(@period, science_guild_events.created_at)::TIMESTAMP AS date,
    COUNT(*)::INT AS join_count
FROM
    science_guild_events
WHERE
    science_guild_events.guild_id = @guild_id
    AND science_guild_events.event_type = @science_guild_event_type_user_join
GROUP BY
    date_trunc(@period, science_guild_events.created_at)
ORDER BY
    date_trunc(@period, science_guild_events.created_at) ASC;
