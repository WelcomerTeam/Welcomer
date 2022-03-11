CREATE TABLE IF NOT EXISTS user_transactions (
    transaction_uuid uuid NOT NULL UNIQUE PRIMARY KEY,
    created_at timestamp NOT NULL,
    updated_at timestamp NOT NULL,
    user_id bigint NOT NULL,
    platform_type integer NOT NULL,
    transaction_id text NOT NULL,
    transaction_status integer NOT NULL,
    currency_code text NOT NULL,
    amount integer NOT NULL
);

CREATE INDEX IF NOT EXISTS user_transactions_user_id ON user_transactions (user_id);

