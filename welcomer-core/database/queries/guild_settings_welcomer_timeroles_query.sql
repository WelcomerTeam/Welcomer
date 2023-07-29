-- name: CreateTimeRolesGuildSettings :one
INSERT INTO guild_settings_timeroles (guild_id, toggle_enabled, timeroles)
    VALUES ($1, $2, $3)
RETURNING
    *;

-- name: CreateOrUpdateTimeRolesGuildSettings :one
INSERT INTO guild_settings_timeroles (guild_id, toggle_enabled, timeroles)
    VALUES ($1, $2, $3)
ON CONFLICT(guild_id) DO UPDATE
    SET toggle_enabled = EXCLUDED.toggle_enabled,
        timeroles = EXCLUDED.timeroles
RETURNING
    *;

-- name: GetTimeRolesGuildSettings :one
SELECT
    *
FROM
    guild_settings_timeroles
WHERE
    guild_id = $1;

-- name: UpdateTimeRolesGuildSettings :execrows
UPDATE
    guild_settings_timeroles
SET
    toggle_enabled = $2,
    timeroles = $3
WHERE
    guild_id = $1;

