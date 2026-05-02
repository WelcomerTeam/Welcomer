-- name: CreatePoll :one
INSERT INTO guild_polls (poll_uuid, created_at, guild_id, created_by, has_ended, is_setup, title, description, accent_colour, image_url, start_time, end_time, poll_options, is_anonymous, maximum_selections, allow_entries, resubmissions, results_visibility, roles_allowed, roles_excluded, minimum_join_date, message_id, channel_id)
VALUES (uuid_generate_v7(), NOW(), $1, $2, FALSE, TRUE, '', '', -1, '', NOW(), $3, '[]', FALSE, 1, TRUE, $4, $5, '[]', '[]', 'epoch', 0, 0)
RETURNING
    *;

-- name: UpdatePoll :one
UPDATE
    guild_polls
SET
    is_setup = $2,
    has_ended = $3,
    title = $4,
    description = $5,
    accent_colour = $6,
    image_url = $7,
    start_time = $8,
    end_time = $9,
    poll_options = $10,
    is_anonymous = $11,
    maximum_selections = $12,
    allow_entries = $13,
    resubmissions = $14,
    results_visibility = $15,
    roles_allowed = $16,
    roles_excluded = $17,
    minimum_join_date = $18,
    message_id = $19,
    channel_id = $20
WHERE
    poll_uuid = $1
RETURNING
    *;

-- name: GetPoll :one
SELECT
    *
FROM
    guild_polls
WHERE
    guild_id = $1
    AND poll_uuid = $2;

-- name: GetPollFromMessageID :one
SELECT
    *
FROM
    guild_polls
WHERE
    guild_id = $1
    AND channel_id = $2
    AND message_id = $3;

-- name: UpdatePollMessage :one
UPDATE
    guild_polls
SET
    message_id = $2,
    channel_id = $3
WHERE
    poll_uuid = $1
RETURNING
    *;

-- name: SetPollEnded :one
UPDATE
    guild_polls
SET
    has_ended = $2
WHERE
    poll_uuid = $1
RETURNING
    *;

-- name: GetExpiredPolls :many
SELECT
    *
FROM
    guild_polls
WHERE
    has_ended = FALSE
    AND is_setup = FALSE
    AND end_time <= NOW()
    AND end_time > 'epoch';