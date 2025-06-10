-- name: CreateUser :one
INSERT INTO users (user_id, created_at, updated_at, name, discriminator, avatar_hash, background)
    VALUES ($1, now(), now(), $2, $3, $4, $5)
RETURNING
    *;

-- name: CreateOrUpdateUser :one
INSERT INTO users (user_id, created_at, updated_at, name, discriminator, avatar_hash, background)
    VALUES ($1, now(), now(), $2, $3, $4, $5)
ON CONFLICT(user_id) DO UPDATE
    SET name = EXCLUDED.name,
        discriminator = EXCLUDED.discriminator,
        avatar_hash = EXCLUDED.avatar_hash,
        updated_at = EXCLUDED.updated_at,
        background = EXCLUDED.background
RETURNING
    *;

-- name: GetUser :one
SELECT
    *
FROM
    users
WHERE
    user_id = $1;

-- name: UpdateUser :execrows
UPDATE
    users
SET
    name = $2,
    discriminator = $3,
    avatar_hash = $4,
    updated_at = now(),
    background = $5
WHERE
    user_id = $1;

