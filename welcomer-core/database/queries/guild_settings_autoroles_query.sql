-- name: CreateAutoRolesGuildSettings :one
INSERT INTO guild_settings_autoroles (guild_id, toggle_enabled, roles)
    VALUES ($1, $2, $3)
RETURNING
    *;

-- name: CreateOrUpdateAutoRolesGuildSettings :one
INSERT INTO guild_settings_autoroles (guild_id, toggle_enabled, roles)
    VALUES ($1, $2, $3)
ON CONFLICT(guild_id) DO UPDATE
    SET toggle_enabled = EXCLUDED.toggle_enabled,
        roles = EXCLUDED.roles
RETURNING
    *;

-- name: GetAutoRolesGuildSettings :one
SELECT
    *
FROM
    guild_settings_autoroles
WHERE
    guild_id = $1;

-- name: UpdateAutoRolesGuildSettings :execrows
UPDATE
    guild_settings_autoroles
SET
    toggle_enabled = $2,
    roles = $3
WHERE
    guild_id = $1;

