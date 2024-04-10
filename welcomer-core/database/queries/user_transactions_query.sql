-- name: CreateUserTransaction :one
INSERT INTO user_transactions (transaction_uuid, created_at, updated_at, user_id, platform_type, transaction_id, transaction_status, currency_code, amount)
    VALUES (uuid_generate_v7(), now(), now(), $1, $2, $3, $4, $5, $6)
RETURNING
    *;

-- name: CreateOrUpdateUserTransaction :one
INSERT INTO user_transactions (transaction_uuid, created_at, updated_at, user_id, platform_type, transaction_id, transaction_status, currency_code, amount)
    VALUES (uuid_generate_v7(), now(), now(), $1, $2, $3, $4, $5, $6)
ON CONFLICT(transaction_uuid) DO UPDATE
    SET user_id = EXCLUDED.user_id,
        platform_type = EXCLUDED.platform_type,
        transaction_id = EXCLUDED.transaction_id,
        transaction_status = EXCLUDED.transaction_status,
        currency_code = EXCLUDED.currency_code,
        amount = EXCLUDED.amount,
        updated_at = EXCLUDED.updated_at
RETURNING
    *;

-- name: GetUserTransaction :one
SELECT
    *
FROM
    user_transactions
WHERE
    transaction_uuid = $1;

-- name: GetUserTransactionsByUserID :many
SELECT
    *
FROM
    user_transactions
WHERE
    user_id = $1;

-- name: UpdateUserTransaction :execrows
UPDATE
    user_transactions
SET
    user_id = $2,
    platform_type = $3,
    transaction_id = $4,
    transaction_status = $5,
    currency_code = $6,
    amount = $7,
    updated_at = now()
WHERE
    transaction_uuid = $1;

-- name: GetUserTransactionsByTransactionID :many
SELECT
    *
FROM
    user_transactions
WHERE
    transaction_id = $1
ORDER BY created_at DESC;