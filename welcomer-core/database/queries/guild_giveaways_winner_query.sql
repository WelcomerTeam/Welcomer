-- name: CreateGiveawayWinner :one
INSERT INTO guild_giveaways_winners (giveaway_winner_uuid, giveaway_uuid, user_id, prize, message_id)
VALUES (uuid_generate_v7(), $1, $2, $3, $4)
RETURNING
    *;

-- name: GetGiveawayWinners :many
SELECT
    *
FROM
    guild_giveaways_winners
WHERE
    giveaway_uuid = $1;

-- name: UpdateGiveawayWinnerUserID :one
UPDATE
    guild_giveaways_winners
SET
    user_id = $2
WHERE
    giveaway_winner_uuid = $1
RETURNING
    *;

-- name: UpdateGiveawayWinnerMessageID :one
UPDATE
    guild_giveaways_winners
SET
    message_id = $2
WHERE
    giveaway_winner_uuid = $1
RETURNING
    *;