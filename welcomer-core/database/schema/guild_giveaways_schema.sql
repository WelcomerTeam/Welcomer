CREATE TABLE IF NOT EXISTS guild_giveaways (
    giveaway_uuid uuid NOT NULL UNIQUE PRIMARY KEY,
    created_at timestamp NOT NULL,
    guild_id bigint NOT NULL,
    created_by bigint NOT NULL,

    is_setup boolean NOT NULL,
    title text NOT NULL,
    start_time timestamp NOT NULL,
    end_time timestamp NOT NULL,
    announce_winners boolean NOT NULL,
    giveaway_prizes jsonb NOT NULL,
    roles_allowed jsonb NOT NULL,
    roles_excluded jsonb NOT NULL,
    minimum_join_date timestamp NOT NULL,
    FOREIGN KEY (guild_id) REFERENCES guilds (guild_id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE INDEX IF NOT EXISTS guild_giveaways_guild_id ON guild_giveaways (guild_id);  