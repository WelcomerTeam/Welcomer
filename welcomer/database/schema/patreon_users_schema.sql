CREATE TABLE IF NOT EXISTS patreon_users (
    patreon_user_id bigint NOT NULL UNIQUE PRIMARY KEY,
    created_at timestamp NOT NULL,
    updated_at timestamp NOT NULL,
    user_id bigint NOT NULL,
    full_name text NOT NULL,
    email text NULL,
    thumb_url text NULL
);

CREATE INDEX IF NOT EXISTS patreon_users_user_id ON patreon_users (user_id);

