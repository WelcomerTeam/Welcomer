-- name: CreateFreeRolesGuildSettings :one
INSERT INTO guild_settings_freeroles (guild_id)
    VALUES ($1)
RETURNING
    *;

-- name: GetFreeRolesGuildSettings :one
SELECT
    *
FROM
    guild_settings_freeroles
WHERE
    guild_id = $1;

-- name: UpdateFreeRolesGuildSettings :execrows
UPDATE
    guild_settings_freeroles
SET
    toggle_enabled = $2,
    roles = $3
WHERE
    guild_id = $1;

