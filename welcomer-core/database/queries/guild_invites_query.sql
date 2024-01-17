-- name: CreateGuildInvites :one
INSERT INTO guild_invites (invite_code, guild_id, created_by, created_at, uses)
    VALUES ($1, $2, $3, $4, $5)
RETURNING
    *;

-- name: CreateOrUpdateGuildInvites :one
INSERT INTO guild_invites (invite_code, guild_id, created_by, created_at, uses)
    VALUES ($1, $2, $3, $4, $5)
ON CONFLICT(invite_code) DO UPDATE
    SET guild_id = EXCLUDED.guild_id,
        created_by = EXCLUDED.created_by,
        created_at = EXCLUDED.created_at,
        uses = EXCLUDED.uses
RETURNING
    *;

-- name: GetGuildInvites :many
SELECT
    *
FROM
    guild_invites
WHERE
    guild_id = $1;

-- name: DeleteGuildInvites :execrows
DELETE FROM
    guild_invites
WHERE
    invite_code = $1
    AND guild_id = $2;