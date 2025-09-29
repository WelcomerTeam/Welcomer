-- name: CreateWelcomerGuildSettings :one
INSERT INTO guild_settings_welcomer (guild_id, auto_delete_welcome_messages, welcome_message_lifetime, auto_delete_welcome_messages_on_leave)
    VALUES ($1, $2, $3, $4)
RETURNING
    *;

-- name: CreateOrUpdateWelcomerGuildSettings :one
INSERT INTO guild_settings_welcomer (guild_id, auto_delete_welcome_messages, welcome_message_lifetime, auto_delete_welcome_messages_on_leave)
    VALUES ($1, $2, $3, $4)
ON CONFLICT(guild_id) DO UPDATE
    SET auto_delete_welcome_messages = EXCLUDED.auto_delete_welcome_messages,
        welcome_message_lifetime = EXCLUDED.welcome_message_lifetime,
        auto_delete_welcome_messages_on_leave = EXCLUDED.auto_delete_welcome_messages_on_leave
RETURNING
    *;

-- name: GetWelcomerGuildSettings :one
SELECT
    *
FROM
    guild_settings_welcomer
WHERE
    guild_id = $1;

-- name: UpdateWelcomerGuildSettings :execrows
UPDATE
    guild_settings_welcomer
SET
    auto_delete_welcome_messages = $2,
    welcome_message_lifetime = $3,
    auto_delete_welcome_messages_on_leave = $4
WHERE
    guild_id = $1;