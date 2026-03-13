-- name: InsertAuditLog :one
INSERT INTO audit_logs (audit_uuid, created_at, guild_id, user_id, audit_type, changes)
    VALUES (uuid_generate_v7(), now(), $1, $2, $3, $4)
RETURNING
    *;

-- name: GetGuildAuditLogs :many
SELECT
    created_at, user_id, audit_type, changes
FROM
    audit_logs
WHERE
    guild_id = $1
ORDER BY
    created_at ASC;