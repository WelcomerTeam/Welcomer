-- name: CreateNewMembership :one
INSERT INTO user_memberships (membership_uuid, created_at, updated_at, started_at, expires_at, status, membership_type, transaction_uuid, user_id, guild_id)
    VALUES (uuid_generate_v7(), now(), now(), $1, $2, $3, $4, $5, $6, $7)
RETURNING
    *;

-- name: CreateOrUpdateNewMembership :one
INSERT INTO user_memberships (membership_uuid, created_at, updated_at, started_at, expires_at, status, membership_type, transaction_uuid, user_id, guild_id)
    VALUES (uuid_generate_v7(), now(), now(), $1, $2, $3, $4, $5, $6, $7)
ON CONFLICT(membership_uuid) DO UPDATE
    SET started_at = EXCLUDED.started_at,
        expires_at = EXCLUDED.expires_at,
        status = EXCLUDED.status,
        transaction_uuid = EXCLUDED.transaction_uuid,
        user_id = EXCLUDED.user_id,
        guild_id = EXCLUDED.guild_id,
        updated_at = EXCLUDED.updated_at
RETURNING
    *;

-- name: DeleteUserMembership :execrows
DELETE FROM
    user_memberships
WHERE
    membership_uuid = $1;

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
    LEFT JOIN guilds ON (user_memberships.guild_id = guilds.guild_id)
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