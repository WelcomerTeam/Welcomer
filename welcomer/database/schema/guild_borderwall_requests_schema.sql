CREATE TABLE IF NOT EXISTS borderwall_requests (
    request_uuid uuid NOT NULL UNIQUE PRIMARY KEY,
    created_at timestamp NOT NULL,
    updated_at timestamp NOT NULL,
    guild_id bigint NOT NULL,
    user_id bigint NOT NULL,
    is_verified boolean NOT NULL,
    verified_at timestamp NULL
);

CREATE INDEX IF NOT EXISTS borderwall_requests_guild_id_user_id ON borderwall_requests (guild_id, user_id);

CREATE INDEX IF NOT EXISTS borderwall_requests_user_id ON borderwall_requests (user_id);

