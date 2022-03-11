-- name: CreateUser :one
INSERT INTO users (user_id, created_at, updated_at, name, discriminator, avatar_hash)
    VALUES ($1, now(), now(), $2, $3, $4)
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
    updated_at = now()
WHERE
    user_id = $1;

