CREATE TABLE IF NOT EXISTS custom_bots (
    custom_bot_uuid uuid NOT NULL UNIQUE PRIMARY KEY,
    guild_id bigint NOT NULL,
    token text NOT NULL,
    created_at timestamp NOT NULL,
    is_active boolean NOT NULL DEFAULT true,
    application_id bigint NOT NULL,
    application_name text NOT NULL,
    application_avatar text NOT NULL
);

CREATE INDEX IF NOT EXISTS custom_bots_guild_id_idx ON custom_bots (guild_id);
CREATE INDEX IF NOT EXISTS custom_bots_application_id_idx ON custom_bots (application_id);