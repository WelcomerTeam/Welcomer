-- name: CreateWelcomerBuilderArtifacts :one
INSERT INTO welcomer_builder_artifacts (artifact_uuid, guild_id, user_id, created_at, image_type, data, reference)
    VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING
    *;

-- name: RemoveWelcomerArtifactsByReference :execrows
DELETE FROM welcomer_builder_artifacts
WHERE
    reference = ANY(@refs::string[])
    AND guild_id = $1;

-- name: GetMinimalWelcomerBuilderArtifactByGuildId :many
SELECT
    artifact_uuid, reference
FROM
    welcomer_builder_artifacts
WHERE
    guild_id = $1;

-- name: GetWelcomerBuilderArtifactsByGuildId :many
SELECT
    *
FROM
    welcomer_builder_artifacts
WHERE
    guild_id = $1;

-- name: GetWelcomerBuilderArtifactByArtifactUUID :one
SELECT
    *
FROM
    welcomer_builder_artifacts
WHERE
    artifact_uuid = $1;