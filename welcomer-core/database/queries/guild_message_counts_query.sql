-- name: UpsertMessageCounts :exec
INSERT INTO guild_message_counts_hour (hour_ts, guild_id, channel_id, user_id, message_count, min_ts)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (hour_ts, guild_id, channel_id, user_id) DO UPDATE SET
	message_count = EXCLUDED.message_count,
	min_ts = LEAST(guild_message_counts_hour.min_ts, EXCLUDED.min_ts);

-- name: AggregateMessageCounts :many
SELECT date_trunc('hour', occurred_at)::timestamp AS hour_ts,
	guild_id,
	channel_id,
	user_id,
	COUNT(message_id)::int AS message_count,
	MIN(occurred_at)::timestamp AS min_ts,
	MAX(occurred_at)::timestamp AS max_ts
FROM ingest_message_events
WHERE occurred_at BETWEEN @date_from AND @date_to AND event_type IN (@event_type_create, @event_type_edit)
GROUP BY hour_ts, guild_id, channel_id, user_id;

-- name: GetActiveUserCounts :one
-- Returns the average daily, weekly and monthly active user counts for a guild within a specified date range.
WITH active_users AS (
	SELECT DISTINCT
		user_id,
		date_trunc('day', hour_ts)::timestamp AS day_ts,
		date_trunc('week', hour_ts)::timestamp AS week_ts,
		date_trunc('month', hour_ts)::timestamp AS month_ts
	FROM guild_message_counts_hour
	WHERE guild_id = @guild_id
		AND hour_ts BETWEEN @date_from AND @date_to
),
days AS (
	SELECT generate_series(
		date_trunc('day', @date_from::timestamp),
		date_trunc('day', @date_to::timestamp),
		INTERVAL '1 day'
	)::timestamp AS period_ts
),
weeks AS (
	SELECT generate_series(
		date_trunc('week', @date_from::timestamp),
		date_trunc('week', @date_to::timestamp),
		INTERVAL '1 week'
	)::timestamp AS period_ts
),
months AS (
	SELECT generate_series(
		date_trunc('month', @date_from::timestamp),
		date_trunc('month', @date_to::timestamp),
		INTERVAL '1 month'
	)::timestamp AS period_ts
),
daily_counts AS (
	SELECT d.period_ts, COUNT(DISTINCT au.user_id)::float AS active_user_count
	FROM days d
	LEFT JOIN active_users au ON au.day_ts = d.period_ts
	GROUP BY d.period_ts
),
weekly_counts AS (
	SELECT w.period_ts, COUNT(DISTINCT au.user_id)::float AS active_user_count
	FROM weeks w
	LEFT JOIN active_users au ON au.week_ts = w.period_ts
	GROUP BY w.period_ts
),
monthly_counts AS (
	SELECT m.period_ts, COUNT(DISTINCT au.user_id)::float AS active_user_count
	FROM months m
	LEFT JOIN active_users au ON au.month_ts = m.period_ts
	GROUP BY m.period_ts
)
SELECT
	COALESCE((SELECT AVG(active_user_count) FROM daily_counts), 0)::float AS average_daily_active_users,
	COALESCE((SELECT AVG(active_user_count) FROM weekly_counts), 0)::float AS average_weekly_active_users,
	COALESCE((SELECT AVG(active_user_count) FROM monthly_counts), 0)::float AS average_monthly_active_users;
