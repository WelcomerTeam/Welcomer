-- name: CreateGiveaway :one
INSERT INTO guild_giveaways (giveaway_uuid, created_at, guild_id, created_by, is_setup, title, end_time, start_time, announce_winners, giveaway_prizes, roles_allowed, roles_excluded, minimum_join_date)
VALUES (uuid_generate_v7(), NOW(), $1, $2, FALSE, $3, $4, NOW(), TRUE, '[]', '[]', '[]', 'epoch')
RETURNING
    *;

-- name: UpdateGiveaway :one
UPDATE
    guild_giveaways
SET
    is_setup = $2,
    title = $3,
    start_time = $4,
    end_time = $5,
    announce_winners = $6,
    giveaway_prizes = $7,
    roles_allowed = $8,
    roles_excluded = $9,
    minimum_join_date = $10
WHERE
    giveaway_uuid = $1
RETURNING
    *;

-- name: GetGiveaway :one
SELECT
    *
FROM
    guild_giveaways
WHERE
    guild_id = $1
    AND giveaway_uuid = $2;