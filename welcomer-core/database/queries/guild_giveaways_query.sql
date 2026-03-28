-- name: CreateGiveaway :one
INSERT INTO guild_giveaways (giveaway_uuid, created_at, guild_id, created_by, is_setup, title, description, end_time, start_time, announce_winners, giveaway_prizes, roles_allowed, roles_excluded, minimum_join_date, message_id, channel_id, accent_colour, image_url)
VALUES (uuid_generate_v7(), NOW(), $1, $2, TRUE, $3, $4, $5, NOW(), TRUE, '[]', '[]', '[]', 'epoch', 0, 0, -1, '')
RETURNING
    *;

-- name: UpdateGiveaway :one
UPDATE
    guild_giveaways
SET
    is_setup = $2,
    title = $3,
    description = $4,
    start_time = $5,
    end_time = $6,
    announce_winners = $7,
    giveaway_prizes = $8,
    roles_allowed = $9,
    roles_excluded = $10,
    minimum_join_date = $11,
    accent_colour = $12,
    image_url = $13
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

-- name: UpdateGiveawayMessage :one
UPDATE
    guild_giveaways
SET
    message_id = $2,
    channel_id = $3
WHERE
    giveaway_uuid = $1
RETURNING
    *;