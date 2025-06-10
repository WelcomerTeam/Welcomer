CREATE TABLE IF NOT EXISTS users (
    user_id bigint NOT NULL UNIQUE PRIMARY KEY,
    created_at timestamp NOT NULL,
    updated_at timestamp NOT NULL,
    name text NOT NULL,
    discriminator text NOT NULL,
    avatar_hash text NOT NULL,
    background text NOT NULL
);

