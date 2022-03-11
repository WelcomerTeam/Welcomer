-- name: CreateBorderwallRequest :one
INSERT INTO borderwall_requests (request_uuid, created_at, updated_at, guild_id, user_id, is_verified, verified_at)
    VALUES (uuid_generate_v4 (), now(), now(), $1, $2, 'false', NULL)
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
    request_uuid,
    created_at,
    updated_at,
    guild_id,
    user_id,
    is_verified,
    verified_at
FROM
    borderwall_requests
WHERE
    guild_id = $1
    AND user_id = $2;

