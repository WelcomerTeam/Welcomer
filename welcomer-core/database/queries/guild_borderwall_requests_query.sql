-- name: CreateBorderwallRequest :one
INSERT INTO borderwall_requests (request_uuid, created_at, updated_at, guild_id, user_id, is_verified, verified_at)
    VALUES (uuid_generate_v7(), now(), now(), $1, $2, $3, $4)
RETURNING
    *;

-- name: CreateOrUpdateBorderwallRequest :one
INSERT INTO borderwall_requests (request_uuid, created_at, updated_at, guild_id, user_id, is_verified, verified_at)
    VALUES (uuid_generate_v7(), now(), now(), $1, $2, $3, $4)
ON CONFLICT(request_uuid) DO UPDATE
    SET updated_at = EXCLUDED.updated_at,
        guild_id = EXCLUDED.guild_id,
        is_verified = EXCLUDED.is_verified,
        verified_at = EXCLUDED.verified_at
RETURNING
    *;

-- name: GetBorderwallRequest :one
SELECT
    *
FROM
    borderwall_requests
WHERE
    request_uuid = $1;

-- name: GetBorderwallRequestsByGuildIDUserID :many
SELECT
    *
FROM
    borderwall_requests
WHERE
    guild_id = $1
    AND user_id = $2;

-- name: UpdateBorderwallRequest :execrows
UPDATE
    borderwall_requests
SET
    updated_at = now(),
    guild_id = $2,
    is_verified = $3,
    verified_at = $4
WHERE
    request_uuid = $1;