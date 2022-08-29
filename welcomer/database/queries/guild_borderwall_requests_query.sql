-- name: CreateBorderwallRequest :one
INSERT INTO borderwall_requests (request_uuid, created_at, updated_at, guild_id, user_id, is_verified, verified_at)
    VALUES (uuid_generate_v4(), now(), now(), $1, $2, 'false', NULL)
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