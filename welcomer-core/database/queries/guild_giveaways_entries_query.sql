-- name: AddGiveawayEntry :exec
INSERT INTO guild_giveaways_entries (guild_giveaway_entry_uuid, giveaway_uuid, user_id, created_at)
VALUES (uuid_generate_v7(), $1, $2, NOW());

-- name: RemoveGiveawayEntry :exec
DELETE FROM guild_giveaways_entries
WHERE giveaway_uuid = $1 AND user_id = $2;