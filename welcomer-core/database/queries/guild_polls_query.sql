-- name: CreatePoll :one
INSERT INTO guild_polls (poll_uuid, created_at, guild_id, created_by, has_ended, is_setup, title, description, accent_colour, image_url, start_time, end_time, poll_options, is_anonymous, maximum_selections, resubmissions, results_visibility, roles_allowed, roles_excluded, minimum_join_date, message_id, channel_id)
VALUES (uuid_generate_v7(), NOW(), $1, $2, FALSE, TRUE, '', '', -1, '', NOW(), $3, '[]', FALSE, 1, $4, $5, '[]', '[]', 'epoch', 0, 0)
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
    resubmissions = $13,
    results_visibility = $14,
    roles_allowed = $15,
    roles_excluded = $16,
    minimum_join_date = $17,
    message_id = $18,
    channel_id = $19
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