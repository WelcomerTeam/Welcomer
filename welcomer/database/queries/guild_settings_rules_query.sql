-- name: CreateRulesGuildSettings :one
INSERT INTO guild_settings_rules (guild_id, toggle_enabled, toggle_dms_enabled, rules)
    VALUES ($1, $2, $3, $4)
RETURNING
    *;

-- name: CreateOrUpdateRulesGuildSettings :one
INSERT INTO guild_settings_rules (guild_id, toggle_enabled, toggle_dms_enabled, rules)
    VALUES ($1, $2, $3, $4)
ON CONFLICT(guild_id) DO UPDATE
    SET toggle_enabled = EXCLUDED.toggle_enabled,
        toggle_dms_enabled = EXCLUDED.toggle_dms_enabled,
        rules = EXCLUDED.rules
RETURNING
    *;

-- name: GetRulesGuildSettings :one
SELECT
    *
FROM
    guild_settings_rules
WHERE
    guild_id = $1;

-- name: UpdateRuleGuildSettings :execrows
UPDATE
    guild_settings_rules
SET
    toggle_enabled = $2,
    toggle_dms_enabled = $3,
    rules = $4
WHERE
    guild_id = $1;

