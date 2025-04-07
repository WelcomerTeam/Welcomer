CREATE TABLE IF NOT EXISTS discord_subscriptions (
    subscription_id text NOT NULL UNIQUE PRIMARY KEY,
    created_at timestamp NOT NULL,
    updated_at timestamp NOT NULL,
    user_id bigint NOT NULL,
    gift_code_flags bigint,
    guild_id bigint,
    starts_at timestamp,
    ends_at timestamp,
    sku_id bigint NOT NULL,
    application_id bigint NOT NULL,
    entitlement_type bigint NOT NULL,
    deleted boolean NOT NULL,
    consumed boolean NOT NULL
);

CREATE INDEX IF NOT EXISTS discord_subscriptions_user_id ON discord_subscriptions (user_id);