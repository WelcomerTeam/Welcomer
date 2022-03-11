-- name: CreateScienceEvent :one
INSERT INTO science_events (event_uuid, created_at, event_type, data)
    VALUES (uuid_generate_v4 (), now(), $1, $2)
RETURNING
    *;

-- name: GetScienceEvent :one
SELECT
    *
FROM
    science_events
WHERE
    event_uuid = $1;

