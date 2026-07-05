CREATE TABLE IF NOT EXISTS guild_polls (
    poll_uuid uuid NOT NULL UNIQUE PRIMARY KEY,
    created_at timestamp NOT NULL,
    guild_id bigint NOT NULL,
    created_by bigint NOT NULL,

    has_ended boolean NOT NULL,
    is_setup boolean NOT NULL,

    title text NOT NULL,
    description text NOT NULL,
    accent_colour bigint NOT NULL,
    image_url text NOT NULL,

    start_time timestamp NOT NULL,
    end_time timestamp NOT NULL,

    poll_options jsonb NOT NULL,
    is_anonymous boolean NOT NULL,
    maximum_selections int NOT NULL,

    -- message configuration
    allow_entries boolean NOT NULL,
    resubmissions text NOT NULL,
    results_visibility text NOT NULL,

    -- eligibility configuration
    roles_allowed jsonb NOT NULL,
    roles_excluded jsonb NOT NULL,
    minimum_join_date timestamp NOT NULL,

    message_id bigint NOT NULL,
    channel_id bigint NOT NULL,
    FOREIGN KEY (guild_id) REFERENCES guilds (guild_id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE INDEX IF NOT EXISTS guild_polls_guild_id ON guild_polls (guild_id);