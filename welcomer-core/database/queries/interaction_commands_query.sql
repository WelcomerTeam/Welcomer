-- name: CreateManyInteractionCommands :copyfrom
INSERT INTO interaction_commands (application_id, command, interaction_id, created_at)
    VALUES ($1, $2, $3, $4);

-- name: ClearInteractionCommands :execrows
DELETE FROM
    interaction_commands
WHERE
    application_id = $1;

-- name: GetInteractionCommand :one
SELECT
    *
FROM
    interaction_commands
WHERE
    application_id = $1
    AND command = $2;