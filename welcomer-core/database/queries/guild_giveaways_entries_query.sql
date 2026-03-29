-- name: AddGiveawayEntry :one
INSERT INTO guild_giveaways_entries (guild_giveaway_entry_uuid, giveaway_uuid, user_id, created_at)
VALUES (uuid_generate_v7(), $1, $2, NOW())
ON CONFLICT (giveaway_uuid, user_id) DO NOTHING
RETURNING guild_giveaway_entry_uuid;

-- name: RemoveGiveawayEntry :exec
DELETE FROM guild_giveaways_entries
WHERE giveaway_uuid = $1 AND user_id = $2;

-- name: CountGiveawayEntries :one
SELECT COUNT(*)::int FROM guild_giveaways_entries
WHERE giveaway_uuid = $1;

-- name: GetGiveawayEntryUsers :many
SELECT user_id FROM guild_giveaways_entries
WHERE giveaway_uuid = $1;