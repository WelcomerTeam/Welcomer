-- name: CreateRulesGuildSettings :one
INSERT INTO guild_settings_rules (guild_id, toggle_enabled, toggle_dms_enabled, rules, moderation_checkup_uuid)
    VALUES ($1, $2, $3, $4, $5)
RETURNING
    *;

-- name: CreateOrUpdateRulesGuildSettings :one
INSERT INTO guild_settings_rules (guild_id, toggle_enabled, toggle_dms_enabled, rules, moderation_checkup_uuid)
    VALUES ($1, $2, $3, $4, $5)
ON CONFLICT(guild_id) DO UPDATE
    SET toggle_enabled = EXCLUDED.toggle_enabled,
        toggle_dms_enabled = EXCLUDED.toggle_dms_enabled,
        rules = EXCLUDED.rules,
        moderation_checkup_uuid = EXCLUDED.moderation_checkup_uuid
RETURNING
    *;

-- name: GetRulesGuildSettings :one
SELECT
    *
FROM
    guild_settings_rules
    LEFT JOIN moderation_checkup ON guild_settings_rules.moderation_checkup_uuid = moderation_checkup.checkup_uuid
WHERE
    guild_settings_rules.guild_id = $1;

-- name: UpdateRuleGuildSettings :execrows
UPDATE
    guild_settings_rules
SET
    toggle_enabled = $2,
    toggle_dms_enabled = $3,
    rules = $4,
    moderation_checkup_uuid = $5
WHERE
    guild_id = $1;

