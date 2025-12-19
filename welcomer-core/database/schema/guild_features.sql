CREATE TABLE IF NOT EXISTS guild_features (
    guild_id bigint NOT NULL UNIQUE PRIMARY KEY,
    created_at timestamp NOT NULL,
    feature text NOT NULL,
    FOREIGN KEY (guild_id) REFERENCES guilds (guild_id),
    UNIQUE (guild_id, feature)
);

CREATE INDEX IF NOT EXISTS guild_features_guild_id ON guild_features (guild_id);