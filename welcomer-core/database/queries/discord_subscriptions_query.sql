-- name: CreateOrUpdateDiscordSubscription :one
INSERT INTO discord_subscriptions (subscription_id, created_at, updated_at, user_id, gift_code_flags, guild_id, starts_at, ends_at, sku_id, application_id, entitlement_type, deleted, consumed)
    VALUES ($1, now(), now(), $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
ON CONFLICT(subscription_id) DO UPDATE
    SET updated_at = EXCLUDED.updated_at,
        user_id = EXCLUDED.user_id,
        gift_code_flags = EXCLUDED.gift_code_flags,
        guild_id = EXCLUDED.guild_id,
        starts_at = EXCLUDED.starts_at,
        ends_at = EXCLUDED.ends_at,
        sku_id = EXCLUDED.sku_id,
        application_id = EXCLUDED.application_id,
        entitlement_type = EXCLUDED.entitlement_type,
        deleted = EXCLUDED.deleted,
        consumed = EXCLUDED.consumed
RETURNING
    *;

-- name: GetDiscordSubscriptionsByUserID :many
SELECT
    *
FROM
    discord_subscriptions
WHERE
    user_id = $1;