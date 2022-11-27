-- name: CreateFreeRolesGuildSettings :one
INSERT INTO guild_settings_freeroles (guild_id, toggle_enabled, roles)
    VALUES ($1, $2, $3)
RETURNING
    *;

-- name: CreateOrUpdateFreeRolesGuildSettings :one
INSERT INTO guild_settings_freeroles (guild_id, toggle_enabled, roles)
    VALUES ($1, $2, $3)
ON CONFLICT(guild_id) DO UPDATE
    SET toggle_enabled = EXCLUDED.toggle_enabled,
        roles = EXCLUDED.roles
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

