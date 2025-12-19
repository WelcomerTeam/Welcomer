CREATE TABLE IF NOT EXISTS welcomer_builder_artifacts (
    artifact_uuid uuid NOT NULL UNIQUE PRIMARY KEY,
    reference text NOT NULL,
    guild_id bigint NOT NULL,
    user_id bigint NOT NULL,
    created_at timestamp NOT NULL,
    image_type text NOT NULL,
    data BYTEA NOT NULL,
    FOREIGN KEY (guild_id) REFERENCES guilds (guild_id) ON DELETE CASCADE ON UPDATE CASCADE
);