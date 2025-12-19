-- name: GetGuildFeatures :many
SELECT feature
FROM guild_features
WHERE guild_id = $1;

-- name: HasGuildFeature :one
SELECT 1
FROM guild_features
WHERE guild_id = $1 AND feature = $2;

-- name: AddGuildFeature :exec
INSERT INTO guild_features (guild_id, created_at, feature)
VALUES ($1, NOW(), $2)
ON CONFLICT (guild_id, feature) DO NOTHING;

-- name: RemoveGuildFeature :exec
DELETE FROM guild_features
WHERE guild_id = $1 AND feature = $2;