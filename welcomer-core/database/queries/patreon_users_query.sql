-- name: CreatePatreonUser :one
INSERT INTO patreon_users (patreon_user_id, created_at, updated_at, user_id, full_name, email, thumb_url, pledge_created_at, pledge_ended_at, tier_id)
    VALUES ($1, now(), now(), $2, $3, $4, $5, $6, $7, $8)
RETURNING
    *;

-- name: CreateOrUpdatePatreonUser :one
INSERT INTO patreon_users (patreon_user_id, created_at, updated_at, user_id, full_name, email, thumb_url, pledge_created_at, pledge_ended_at, tier_id)
    VALUES ($1, now(), now(), $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT(patreon_user_id) DO UPDATE
    SET updated_at = EXCLUDED.updated_at,
        user_id = EXCLUDED.user_id,
        full_name = EXCLUDED.full_name,
        email = EXCLUDED.email,
        thumb_url = EXCLUDED.thumb_url,
        pledge_created_at = EXCLUDED.pledge_created_at,
        pledge_ended_at = EXCLUDED.pledge_ended_at,
        tier_id = EXCLUDED.tier_id
RETURNING
    *;

-- name: GetPatreonUser :one
SELECT
    *
FROM
    patreon_users
WHERE
    patreon_user_id = $1;

-- name: GetPatreonUsersByUserID :many
SELECT
    *
FROM
    patreon_users
WHERE
    user_id = $1;

-- name: UpdatePatreonUser :execrows
UPDATE
    patreon_users
SET
    updated_at = now(),
    user_id = $2,
    full_name = $3,
    email = $4,
    thumb_url = $5,
    pledge_created_at = $6,
    pledge_ended_at = $7,
    tier_id = $8
WHERE
    patreon_user_id = $1;

-- name: DeletePatreonUser :execrows
DELETE FROM patreon_users
WHERE patreon_user_id = $1 AND user_id = $2;

