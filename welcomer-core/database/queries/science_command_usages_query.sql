-- name: CreateCommandUsage :one
INSERT INTO science_command_usages (command_uuid, created_at, updated_at, guild_id, user_id, channel_id, command, errored, execution_time_ms)
    VALUES (uuid_generate_v4(), now(), now(), $1, $2, $3, $4, $5, $6)
RETURNING
    *;

-- name: GetCommandUsage :one
SELECT
    *
FROM
    science_command_usages
WHERE
    command_uuid = $1;

