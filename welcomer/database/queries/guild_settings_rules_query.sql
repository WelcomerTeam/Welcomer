-- name: CreateRulesGuildSettings :one
INSERT INTO guild_settings_rules (guild_id)
    VALUES ($1)
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

