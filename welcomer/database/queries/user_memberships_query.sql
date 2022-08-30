-- name: CreateNewMembership :one
INSERT INTO user_memberships (membership_uuid, created_at, updated_at, started_at, expires_at, status, membership_type, transaction_uuid, user_id, guild_id)
    VALUES (uuid_generate_v4(), now(), now(), $1, $2, $3, $4, $5, $6, $7)
RETURNING
    *;

-- name: GetUserMembership :one
SELECT
    *
FROM
    user_memberships
    LEFT JOIN user_transactions ON (user_memberships.transaction_uuid = user_transactions.transaction_uuid)
WHERE
    user_memberships.membership_uuid = $1;

-- name: GetUserMembershipsByUserID :many
SELECT
    *
FROM
    user_memberships
    LEFT JOIN user_transactions ON (user_memberships.transaction_uuid = user_transactions.transaction_uuid)
WHERE
    user_memberships.user_id = $1;

-- name: GetUserMembershipsByGuildID :many
SELECT
    *
FROM
    user_memberships
    LEFT JOIN user_transactions ON (user_memberships.transaction_uuid = user_transactions.transaction_uuid)
WHERE
    user_memberships.guild_id = $1;

-- name: UpdateUserMembership :execrows
UPDATE
    user_memberships
SET
    started_at = $2,
    expires_at = $3,
    status = $4,
    transaction_uuid = $5,
    user_id = $6,
    guild_id = $7,
    updated_at = now()
WHERE
    membership_uuid = $1;