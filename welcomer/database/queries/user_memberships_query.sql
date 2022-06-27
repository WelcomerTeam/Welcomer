-- name: CreateNewMembership :one
INSERT INTO user_memberships (membership_uuid, origin_membership_id, created_at, updated_at, status, transaction_uuid, user_id, guild_id)
    VALUES (uuid_generate_v4(), $1, now(), now(), $2, $3, $4, $5)
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
    status = $3,
    transaction_uuid = $4,
    user_id = $5,
    guild_id = $6,
    updated_at = now()
WHERE
    membership_uuid = $1;