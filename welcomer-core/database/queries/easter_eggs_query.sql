-- name: InsertEasterEgg :one
INSERT INTO easter_eggs(easter_egg_uuid, guild_id, channel_id, message_id, user_id, claimed_egg, created_at, wm_user_id)
VALUES (uuid_generate_v7(), $1, $2, $3, $4, $5, NOW(), $6)
ON CONFLICT (guild_id, channel_id, message_id) DO NOTHING
RETURNING easter_egg_uuid;

-- name: GetEasterEggsByUserID :many
SELECT claimed_egg, MIN(created_at), MAX(created_at), COUNT(*)
FROM easter_eggs
WHERE user_id = $1
GROUP BY claimed_egg
ORDER BY COUNT(*) DESC;

-- name: GetCollectedEasterEggsByGuildID :many
SELECT COUNT(*)::int AS count, user_id, SUM(COUNT(*)) OVER()::int AS total
FROM easter_eggs
WHERE guild_id = $1
GROUP BY user_id
ORDER BY COUNT(*) DESC
LIMIT 20;

-- name: GetCollectedEasterEggs :many
SELECT COUNT(*)::int AS count, user_id, SUM(COUNT(*)) OVER()::int AS total
FROM easter_eggs
GROUP BY user_id
ORDER BY COUNT(*) DESC
LIMIT 20;

-- name: GetClaimedWM :one
SELECT COUNT(*)::int
FROM easter_eggs
WHERE guild_id = $1 AND wm_user_id = $2;