-- Stores only bucket-level statistics. No guild IDs or individual guild metric values are retained.
CREATE TABLE IF NOT EXISTS guild_analytics_benchmarks (
    member_count_bucket smallint NOT NULL CHECK (member_count_bucket BETWEEN 0 AND 6),
    metric_key text NOT NULL CHECK (
        metric_key IN (
            'retention_rate_30d',
            'daily_active_users',
            'weekly_unique_messaging_users',
            'weekly_messages_per_user',
            'weekly_unique_voice_users',
            'average_voice_session_ms',
            'active_members_percentage',
            'average_time_to_first_message_seconds',
            'average_account_age_days'
        )
    ),
    benchmark_value double precision,
    sample_count integer NOT NULL CHECK (sample_count >= 0),
    updated_at timestamp with time zone NOT NULL,
    statistic text NOT NULL DEFAULT 'median',
    PRIMARY KEY (member_count_bucket, metric_key),
    CHECK (
        (sample_count >= 30 AND benchmark_value IS NOT NULL)
        OR (sample_count < 30 AND benchmark_value IS NULL)
    )
);
