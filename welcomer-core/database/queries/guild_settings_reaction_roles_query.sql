-- name: CreateOrUpdateReactionRoleSetting :one
INSERT INTO guild_settings_reaction_roles (reaction_role_id, guild_id, toggle_enabled, channel_id, message_id, is_system_message, system_message_format, reaction_role_type, roles)
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
ON CONFLICT(reaction_role_id) DO UPDATE
    SET toggle_enabled = EXCLUDED.toggle_enabled,
        channel_id = EXCLUDED.channel_id,
        is_system_message = EXCLUDED.is_system_message,
        system_message_format = EXCLUDED.system_message_format,
        reaction_role_type = EXCLUDED.reaction_role_type,
        roles = EXCLUDED.roles
RETURNING
    *;

-- name: DeleteReactionRoleSettings :execrows
DELETE FROM guild_settings_reaction_roles
WHERE reaction_role_id IN ($1) AND guild_id = $2;

-- name: GetReactionRoleSettingByGuildId :many
SELECT
    *
FROM
    guild_settings_reaction_roles
WHERE
    guild_id = $1;

-- name: GetReactionRoleSettingById :one
SELECT
    *
FROM
    guild_settings_reaction_roles
WHERE
    reaction_role_id = $1
    AND guild_id = $2;

-- name: GetReactionRoleSettingByMessageId :one
SELECT
    *
FROM
    guild_settings_reaction_roles
WHERE
    message_id = $1
    AND guild_id = $2;

-- name: UpdateReactionRoleSettingMessageId :execrows
UPDATE
    guild_settings_reaction_roles
SET
    message_id = $3
WHERE
    reaction_role_id = $1
    AND guild_id = $2;

-- name: DisableReactionRoleSettingByMessageId :execrows
UPDATE
    guild_settings_reaction_roles
SET
    toggle_enabled = FALSE,
    message_id = 0
WHERE
    message_id = $1
    AND guild_id = $2;