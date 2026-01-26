-- name: CreateReactionRolesGuildSettings :one
INSERT INTO guild_settings_reaction_roles (guild_id, toggle_enabled, reaction_roles)
    VALUES ($1, $2, $3)
RETURNING
    *;

-- name: CreateOrUpdateReactionRolesGuildSettings :one
INSERT INTO guild_settings_reaction_roles (guild_id, toggle_enabled, reaction_roles)
    VALUES ($1, $2, $3)
ON CONFLICT(guild_id) DO UPDATE
    SET toggle_enabled = EXCLUDED.toggle_enabled,
        reaction_roles = EXCLUDED.reaction_roles
RETURNING
    *;

-- name: GetReactionRolesGuildSettings :one
SELECT
    *
FROM
    guild_settings_reaction_roles
WHERE
    guild_id = $1;

-- name: UpdateReactionRolesGuildSettings :execrows
UPDATE
    guild_settings_reaction_roles
SET
    toggle_enabled = $2,
    reaction_roles = $3
WHERE
    guild_id = $1;