CREATE TABLE IF NOT EXISTS guild_welcomer_images (
    image_uuid uuid NOT NULL UNIQUE PRIMARY KEY,
    guild_id bigint NOT NULL,
    created_at timestamp NOT NULL,
    image_format integer NOT NULL
);

