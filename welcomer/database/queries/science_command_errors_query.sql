-- name: CreateCommandError :one
INSERT INTO science_command_errors (command_uuid, trace, data)
    VALUES ($1, $2, $3)
RETURNING
    *;

-- name: GetCommandError :one
SELECT
    *
FROM
    science_command_usages
    LEFT JOIN science_command_errors ON (science_command_errors.command_uuid = science_command_usages.command_uuid)
WHERE
    science_command_usages.command_uuid = $1;

