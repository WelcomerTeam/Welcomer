-- name: CreateAutoModerationGuildSettings :one
INSERT INTO guild_settings_automoderation (guild_id)
    VALUES ($1)
RETURNING
    *;

-- name: GetAutoModerationGuildSettings :one
SELECT
    *
FROM
    guild_settings_automoderation
WHERE
    guild_id = $1;

-- name: UpdateAutoModerationGuildSettings :execrows
UPDATE
    guild_settings_automoderation
SET
    toggle_masscaps = $2,
    toggle_massmentions = $3,
    toggle_urls = $4,
    toggle_invites = $5,
    toggle_regex = $6,
    toggle_phishing = $7,
    toggle_massemojis = $8,
    threshold_mass_caps_percentage = $9,
    threshold_mass_mentions = $10,
    threshold_mass_emojis = $11,
    blacklisted_urls = $12,
    whitelisted_urls = $13,
    regex_patterns = $14,
    toggle_smartmod = $15
WHERE
    guild_id = $1;

