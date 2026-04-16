-- name: AddPollEntry :one
INSERT INTO guild_polls_entries (guild_poll_entry_uuid, poll_uuid, user_id, created_at, option_index)
VALUES (uuid_generate_v7(), $1, $2, NOW(), $3)
ON CONFLICT (poll_uuid, user_id) DO UPDATE SET option_index = EXCLUDED.option_index, created_at = NOW()
RETURNING guild_poll_entry_uuid;

-- name: RemovePollEntriesNotMatching :exec
DELETE FROM guild_polls_entries
WHERE poll_uuid = $1 AND user_id = $2 AND option_index NOT IN ($3);

-- name: CountPollEntriesByUniqueUsers :one
SELECT COUNT(user_id)::int FROM guild_polls_entries
WHERE poll_uuid = $1;

-- name: GetPollEntryUsers :many
SELECT DISTINCT user_id FROM guild_polls_entries
WHERE poll_uuid = $1;

-- name: GetPollEntries :many
SELECT * FROM guild_polls_entries
WHERE poll_uuid = $1
ORDER BY created_at DESC;

-- name: GetPollEntriesForUser :many
SELECT * FROM guild_polls_entries
WHERE poll_uuid = $1 AND user_id = $2
ORDER BY created_at DESC;

-- name: GetPollEntriesFromMessageID :many
SELECT
    *
FROM
    guild_polls_entries
    JOIN guild_polls ON guild_polls.poll_uuid = guild_polls_entries.poll_uuid
WHERE
    guild_polls.guild_id = $1
    AND guild_polls.channel_id = $2
    AND guild_polls.message_id = $3;