-- name: CreateOrUpdatePaypalSubscription :one
INSERT INTO paypal_subscriptions (subscription_id, created_at, updated_at, user_id, payer_id, last_billed_at, next_billing_at, subscription_status, plan_id, quantity)
    VALUES ($1, now(), now(), $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT(subscription_id) DO UPDATE
    SET updated_at = EXCLUDED.updated_at,
        user_id = EXCLUDED.user_id,
        payer_id = EXCLUDED.payer_id,
        last_billed_at = EXCLUDED.last_billed_at,
        next_billing_at = EXCLUDED.next_billing_at,
        subscription_status = EXCLUDED.subscription_status,
        plan_id = EXCLUDED.plan_id,
        quantity = EXCLUDED.quantity
RETURNING
    *;

-- name: GetPaypalSubscriptionsByUserID :many
SELECT
    *
FROM
    paypal_subscriptions
WHERE
    user_id = $1;

-- name: GetAllPaypalSubscriptions :many
SELECT
    *
FROM
    paypal_subscriptions;

-- name: GetPaypalSubscriptionBySubscriptionID :one
SELECT
    *
FROM
    paypal_subscriptions
WHERE
    subscription_id = $1;

-- name: UpdatePaypalSubscriptionLastBilledAt :execrows
UPDATE
    paypal_subscriptions
SET
    last_billed_at = GREATEST(last_billed_at, $2),
    updated_at = now()
WHERE
    subscription_id = $1;