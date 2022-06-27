-- name: CreateNewMembership :one
INSERT INTO user_memberships (membership_uuid, origin_membership_id, created_at, updated_at, started_at, expires_at, status, membership_type, transaction_uuid, user_id, guild_id)
    VALUES (uuid_generate_v4(), $1, now(), now(), $2, $3, $4, $5, $6, $7, $8)
RETURNING
    *;

-- name: GetUserMembership :one
SELECT
    *
FROM
    user_memberships
WHERE
    membership_uuid = $1;

-- name: GetUserMembershipsByUserID :many
SELECT
    *
FROM
    user_memberships
WHERE
    user_id = $1;

-- name: GetUserMembershipsByGuildID :many
SELECT
    *
FROM
    user_memberships
WHERE
    guild_id = $1;

-- name: UpdateUserMembership :execrows
UPDATE
    user_memberships
SET
    origin_membership_id = $2,
    started_at = $3,
    expires_at = $4,
    status = $5,
    transaction_uuid = $6,
    user_id = $7,
    guild_id = $8,
    updated_at = now()
WHERE
    membership_uuid = $1;