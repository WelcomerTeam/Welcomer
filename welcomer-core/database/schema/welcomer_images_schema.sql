CREATE TABLE IF NOT EXISTS welcomer_images (
    welcomer_image_uuid uuid NOT NULL UNIQUE PRIMARY KEY,
    guild_id bigint NOT NULL,
    created_at timestamp NOT NULL,
    image_type text NOT NULL,
    data BYTEA NOT NULL
);
