-- name: GetGuildMessageActivity :many
SELECT
    date_trunc(@period, hour_ts)::timestamp AS date,
    SUM(message_count)::bigint AS message_count,
    COUNT(DISTINCT user_id)::bigint AS unique_users
FROM guild_message_counts_hour
WHERE guild_id = @guild_id
    AND hour_ts BETWEEN @date_from AND @date_to
GROUP BY date_trunc(@period, hour_ts)
ORDER BY date_trunc(@period, hour_ts);

-- name: GetGuildMessageSummary :one
SELECT
    COALESCE(SUM(message_count), 0)::bigint AS total_messages,
    COUNT(DISTINCT user_id)::bigint AS unique_users
FROM guild_message_counts_hour
WHERE guild_id = @guild_id
    AND hour_ts BETWEEN @date_from AND @date_to;

-- name: GetGuildAnalyticsBenchmarkInputs :many
-- Returns per-guild messaging and voice totals for the fixed, complete UTC activity week.
WITH weekly_messages AS (
    SELECT
        message_count_row.guild_id,
        SUM(message_count_row.message_count)::bigint AS total_messages,
        COUNT(DISTINCT message_count_row.user_id)::bigint AS unique_users
    FROM guild_message_counts_hour AS message_count_row
    WHERE message_count_row.guild_id = @guild_id
        AND message_count_row.hour_ts >= @week_from
        AND message_count_row.hour_ts < @week_to
    GROUP BY message_count_row.guild_id
),
weekly_voice AS (
    SELECT
        COUNT(*)::bigint AS sessions,
        COUNT(DISTINCT user_id)::bigint AS unique_users,
        SUM(total_time_ms)::bigint AS total_time_ms
    FROM guild_voice_channel_stats
    WHERE guild_id = @guild_id
        AND start_ts >= @week_from
        AND start_ts < @week_to
)
SELECT
    guild.guild_id,
    guild.member_count,
    weekly_messages.total_messages,
    weekly_messages.unique_users,
    weekly_voice.sessions,
    weekly_voice.unique_users,
    weekly_voice.total_time_ms
FROM guilds guild
LEFT JOIN weekly_messages ON weekly_messages.guild_id = guild.guild_id
LEFT JOIN weekly_voice ON true
WHERE guild.guild_id = @guild_id
    AND guild.member_count > 0;

-- name: GetGuildAnalyticsBenchmarkGuilds :many
-- Restricts benchmark candidates to guilds with recent analytics events before per-bucket sampling.
SELECT guild.guild_id, guild.member_count
FROM guilds AS guild
WHERE guild.member_count > 0
    AND EXISTS (
        SELECT 1
        FROM science_guild_events AS recent_event
        WHERE recent_event.guild_id = guild.guild_id
            AND recent_event.created_at >= @active_since
            AND recent_event.created_at < @active_until
    )
ORDER BY guild.guild_id;

-- name: GetGuildAnalyticsBenchmarkJoinMetrics :one
-- Cohort members receive the same 30-day observation window for both retention and onboarding.
-- First-message timing uses the hourly min_ts aggregate, not raw message events.
WITH joined AS (
    SELECT DISTINCT ON (guild_event.user_id)
        guild_event.user_id,
        guild_event.created_at AS joined_at
    FROM science_guild_events AS guild_event
    WHERE guild_event.guild_id = @guild_id
        AND guild_event.event_type = @join_event_type
        AND guild_event.user_id IS NOT NULL
        AND guild_event.created_at >= @cohort_from
        AND guild_event.created_at < @cohort_to
    ORDER BY guild_event.user_id, guild_event.created_at DESC
),
member_retention AS (
    SELECT
        joined.user_id,
        joined.joined_at,
        NOT EXISTS (
            SELECT 1
            FROM science_guild_events leave_event
            WHERE leave_event.guild_id = @guild_id
                AND leave_event.user_id = joined.user_id
                AND leave_event.event_type = @leave_event_type
                AND leave_event.created_at > @cohort_from
                AND leave_event.created_at <= @cohort_to + INTERVAL '30 days'
                AND leave_event.created_at > joined.joined_at
                AND leave_event.created_at <= joined.joined_at + INTERVAL '30 days'
        ) AS retained
    FROM joined
),
first_messages AS (
    SELECT
        joined.user_id,
        joined.joined_at,
        MIN(message.min_ts) AS first_message_at
    FROM joined
    INNER JOIN guild_message_counts_hour message
        ON message.guild_id = @guild_id
        AND message.user_id = joined.user_id
        AND message.hour_ts >= date_trunc('hour', @cohort_from)
        AND message.hour_ts < @cohort_to + INTERVAL '30 days'
        AND message.hour_ts >= date_trunc('hour', joined.joined_at)
        AND message.hour_ts < joined.joined_at + INTERVAL '30 days'
        AND message.min_ts >= joined.joined_at
        AND message.min_ts < joined.joined_at + INTERVAL '30 days'
    GROUP BY joined.user_id, joined.joined_at
)
SELECT
    COUNT(member_retention.user_id)::bigint AS cohort_size,
    COUNT(*) FILTER (WHERE member_retention.retained)::bigint AS retained_members,
    COUNT(first_messages.user_id)::bigint AS messaged_members,
    COALESCE(
        AVG(EXTRACT(EPOCH FROM first_messages.first_message_at - member_retention.joined_at))
            FILTER (WHERE first_messages.first_message_at IS NOT NULL),
        0
    )::double precision AS average_time_to_first_message_seconds
FROM member_retention
LEFT JOIN first_messages
    ON first_messages.user_id = member_retention.user_id;

-- name: GetGuildAnalyticsBenchmarkBucket :many
SELECT
    metric_key,
    benchmark_value,
    sample_count,
    updated_at
FROM guild_analytics_benchmarks
WHERE member_count_bucket = @member_count_bucket
    AND statistic = 'p90'
ORDER BY metric_key;

-- name: UpsertGuildAnalyticsBenchmark :exec
INSERT INTO guild_analytics_benchmarks (
    member_count_bucket,
    metric_key,
    benchmark_value,
    sample_count,
    updated_at,
    statistic
)
VALUES (
    @member_count_bucket,
    @metric_key,
    @benchmark_value,
    @sample_count,
    @updated_at,
    'p90'
)
ON CONFLICT (member_count_bucket, metric_key) DO UPDATE SET
    benchmark_value = EXCLUDED.benchmark_value,
    sample_count = EXCLUDED.sample_count,
    updated_at = EXCLUDED.updated_at,
    statistic = EXCLUDED.statistic;

-- name: DeleteGuildAnalyticsBenchmarks :exec
DELETE FROM guild_analytics_benchmarks;

-- name: GetGuildTimeToFirstMessage :one
WITH joined AS (
    SELECT DISTINCT ON (user_id)
        user_id,
        created_at AS joined_at
    FROM science_guild_events
    WHERE science_guild_events.guild_id = @guild_id
        AND science_guild_events.event_type = @join_event_type
        AND science_guild_events.user_id IS NOT NULL
        AND science_guild_events.created_at BETWEEN @date_from AND @date_to
    ORDER BY science_guild_events.user_id, science_guild_events.created_at
),
first_messages AS (
    SELECT guild_message_counts_hour.user_id, MIN(guild_message_counts_hour.min_ts) AS first_at
    FROM guild_message_counts_hour
    WHERE guild_message_counts_hour.guild_id = @guild_id
        AND guild_message_counts_hour.hour_ts BETWEEN @date_from AND @date_to
    GROUP BY guild_message_counts_hour.user_id
),
durations AS (
    SELECT EXTRACT(EPOCH FROM first_messages.first_at - joined.joined_at)::bigint AS seconds
    FROM joined
    INNER JOIN first_messages USING (user_id)
    WHERE first_messages.first_at >= joined.joined_at
)
SELECT
    COUNT(*)::bigint AS users,
    COALESCE(AVG(seconds), 0)::bigint AS average_seconds,
    COALESCE(percentile_cont(0.5) WITHIN GROUP (ORDER BY seconds), 0)::bigint AS median_seconds
FROM durations;

-- name: GetGuildMessageChannels :many
SELECT channel_id::text AS key, SUM(message_count)::bigint AS value
FROM guild_message_counts_hour
WHERE guild_id = @guild_id
    AND hour_ts BETWEEN @date_from AND @date_to
GROUP BY channel_id
ORDER BY value DESC
LIMIT 10;

-- name: GetGuildMessageAuthors :many
SELECT user_id::text AS key, SUM(message_count)::bigint AS value
FROM guild_message_counts_hour
WHERE guild_id = @guild_id
    AND hour_ts BETWEEN @date_from AND @date_to
GROUP BY user_id
ORDER BY value DESC
LIMIT 10;

-- name: GetGuildMessageContributors :many
SELECT user_id::text AS key, COUNT(DISTINCT hour_ts)::bigint AS value
FROM guild_message_counts_hour
WHERE guild_id = @guild_id
    AND hour_ts BETWEEN @date_from AND @date_to
GROUP BY user_id
ORDER BY value DESC
LIMIT 10;

-- name: GetGuildMessageHeatmap :many
SELECT
    EXTRACT(DOW FROM hour_ts AT TIME ZONE (@timezone::text))::int AS day_of_week,
    EXTRACT(HOUR FROM hour_ts AT TIME ZONE (@timezone::text))::int AS hour_of_day,
    SUM(message_count)::bigint AS message_count
FROM guild_message_counts_hour
WHERE guild_id = @guild_id
    AND hour_ts BETWEEN @date_from AND @date_to
GROUP BY day_of_week, hour_of_day
ORDER BY day_of_week, hour_of_day;

-- name: GetGuildVoiceActivity :many
SELECT
    date_trunc(@period, start_ts)::timestamp AS date,
    COUNT(*)::bigint AS sessions,
    COUNT(DISTINCT user_id)::bigint AS unique_users,
    SUM(total_time_ms)::bigint AS total_time_ms
FROM guild_voice_channel_stats
WHERE guild_id = @guild_id
    AND start_ts BETWEEN @date_from AND @date_to
GROUP BY date_trunc(@period, start_ts)
ORDER BY date_trunc(@period, start_ts);

-- name: GetGuildVoiceAuthors :many
SELECT user_id::text AS key, SUM(total_time_ms)::bigint AS value
FROM guild_voice_channel_stats
WHERE guild_id = @guild_id
    AND start_ts BETWEEN @date_from AND @date_to
GROUP BY user_id
ORDER BY value DESC
LIMIT 10;

-- name: GetGuildVoiceSummary :one
SELECT
    COUNT(*)::bigint AS sessions,
    COUNT(DISTINCT user_id)::bigint AS unique_users,
    COALESCE(SUM(total_time_ms), 0)::bigint AS total_time_ms
FROM guild_voice_channel_stats
WHERE guild_id = @guild_id
    AND start_ts BETWEEN @date_from AND @date_to;

-- name: GetGuildInviteJoinCounts :many
SELECT
    guild_invites.invite_code AS key,
    guild_invites.created_by,
    COUNT(*)::bigint AS joins
FROM science_guild_events
INNER JOIN guild_invites
    ON guild_invites.guild_id = science_guild_events.guild_id
    AND guild_invites.invite_code = science_guild_events.data ->> 'invite_code'
WHERE science_guild_events.guild_id = @guild_id
    AND science_guild_events.event_type = @event_type
    AND science_guild_events.created_at BETWEEN @date_from AND @date_to
GROUP BY guild_invites.invite_code, guild_invites.created_by
ORDER BY joins DESC;

-- name: GetGuildInviteActivity :many
SELECT
    date_trunc(@period, science_guild_events.created_at)::timestamp AS date,
    COUNT(*)::bigint AS joins,
    COUNT(*) FILTER (
        WHERE guild_invites.invite_code IS NOT NULL
    )::bigint AS tracked_joins
FROM science_guild_events
LEFT JOIN guild_invites
    ON guild_invites.guild_id = science_guild_events.guild_id
    AND guild_invites.invite_code = science_guild_events.data ->> 'invite_code'
WHERE science_guild_events.guild_id = @guild_id
    AND science_guild_events.event_type = @event_type
    AND science_guild_events.created_at BETWEEN @date_from AND @date_to
GROUP BY date_trunc(@period, science_guild_events.created_at)
ORDER BY date_trunc(@period, science_guild_events.created_at);

-- name: GetGuildBorderwallActivity :many
SELECT
    date_trunc(@period, created_at)::timestamp AS date,
    COUNT(*)::bigint AS requests,
    COUNT(*) FILTER (WHERE is_verified)::bigint AS verified
FROM borderwall_requests
WHERE guild_id = @guild_id
    AND created_at BETWEEN @date_from AND @date_to
GROUP BY date_trunc(@period, created_at)
ORDER BY date_trunc(@period, created_at);

-- name: GetGuildBorderwallCountries :many
SELECT COALESCE(NULLIF(country_code, ''), 'Unknown') AS key, COUNT(*)::bigint AS value
FROM borderwall_requests
WHERE guild_id = @guild_id
    AND created_at BETWEEN @date_from AND @date_to
GROUP BY COALESCE(NULLIF(country_code, ''), 'Unknown')
ORDER BY value DESC
LIMIT 10;

-- name: GetGuildBorderwallDevices :many
SELECT COALESCE(NULLIF(ua_os, ''), 'Unknown') AS key, COUNT(*)::bigint AS value
FROM borderwall_requests
WHERE guild_id = @guild_id
    AND created_at BETWEEN @date_from AND @date_to
GROUP BY COALESCE(NULLIF(ua_os, ''), 'Unknown')
ORDER BY value DESC
LIMIT 10;

-- name: GetGuildBorderwallSummary :one
SELECT
    COUNT(*)::bigint AS requests,
    COUNT(*) FILTER (WHERE is_verified)::bigint AS verified,
    COALESCE(AVG(EXTRACT(EPOCH FROM (verified_at - created_at)) * 1000)
        FILTER (WHERE is_verified AND verified_at IS NOT NULL), 0)::bigint AS average_verification_time_ms
FROM borderwall_requests
WHERE guild_id = @guild_id
    AND created_at BETWEEN @date_from AND @date_to;

-- name: GetGuildRoleActivity :many
SELECT
    event_type,
    COALESCE(data ->> 'role_id', '') AS key,
    COUNT(*)::bigint AS value
FROM science_guild_events
WHERE guild_id = @guild_id
    AND event_type IN (
        @timerole_event_type,
        @reactionrole_event_type,
        @autorole_event_type,
        @freerole_event_type
    )
    AND created_at BETWEEN @date_from AND @date_to
GROUP BY event_type, data ->> 'role_id'
ORDER BY value DESC;
