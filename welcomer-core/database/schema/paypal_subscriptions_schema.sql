CREATE TABLE IF NOT EXISTS paypal_subscriptions (
    subscription_id text NOT NULL UNIQUE PRIMARY KEY,
    created_at timestamp NOT NULL,
    updated_at timestamp NOT NULL,
    user_id bigint NOT NULL,
    payer_id text NOT NULL,
    last_billed_at timestamp NOT NULL,
    next_billing_at timestamp NOT NULL,
    subscription_status text NOT NULL,
    plan_id text NOT NULL,
    quantity text NOT NULL
);

CREATE INDEX IF NOT EXISTS paypal_subscriptions_user_id ON paypal_subscriptions (user_id);