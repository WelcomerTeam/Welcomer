-- name: CreateBorderwallRequest :one
INSERT INTO borderwall_requests (request_uuid, created_at, updated_at, guild_id, user_id, is_verified)
    VALUES (uuid_generate_v7(), now(), now(), $1, $2, FALSE)
RETURNING
    *;

-- name: InsertBorderwallRequest :one
INSERT INTO borderwall_requests (request_uuid, created_at, updated_at, guild_id, user_id, is_verified, ip_address)
    VALUES ($1, now(), now(), $2, $3, $4, $5)
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

-- name: GetBorderwallRequestsByIPAddress :many
SELECT
    *
FROM
    borderwall_requests
WHERE
    ip_address = $1;

-- name: UpdateBorderwallRequest :execrows
UPDATE
    borderwall_requests
SET
    updated_at = now(),
    is_verified = $2,
    verified_at = $3,
    ip_address = $4,
    recaptcha_score = $5,
    ipintel_score = $6,
    country_code = $7,
    ua_family = $8,
    ua_family_version = $9,
    ua_os = $10,
    ua_os_version = $11
WHERE
    request_uuid = $1;