-- name: CreateGiveaway :one
INSERT INTO guild_giveaways (giveaway_uuid, created_at, guild_id, created_by, allow_entries, has_ended, is_setup, title, description, end_time, start_time, announce_winners, giveaway_prizes, roles_allowed, roles_excluded, minimum_join_date, message_id, channel_id, accent_colour, image_url, show_prizes, show_entries)
VALUES (uuid_generate_v7(), NOW(), $1, $2, TRUE, FALSE, TRUE, $3, $4, $5, NOW(), TRUE, '[]', '[]', '[]', 'epoch', 0, 0, -1, '', TRUE, TRUE)
RETURNING
    *;

-- name: UpdateGiveaway :one
UPDATE
    guild_giveaways
SET
    is_setup = $2,
    allow_entries = $3,
    has_ended = $4,
    title = $5,
    description = $6,
    start_time = $7,
    end_time = $8,
    announce_winners = $9,
    giveaway_prizes = $10,
    roles_allowed = $11,
    roles_excluded = $12,
    minimum_join_date = $13,
    accent_colour = $14,
    image_url = $15,
    show_prizes = $16,
    show_entries = $17
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

-- name: GetGiveawayFromMessageID :one
SELECT
    *
FROM
    guild_giveaways
WHERE
    guild_id = $1
    AND channel_id = $2
    AND message_id = $3;

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

-- name: SetGiveawayEnded :one
UPDATE
    guild_giveaways
SET
    has_ended = $2
WHERE
    giveaway_uuid = $1
RETURNING
    *;