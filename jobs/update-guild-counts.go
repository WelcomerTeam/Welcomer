package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/WelcomerTeam/Discord/discord"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	"github.com/jackc/pgx/v4"
	_ "github.com/joho/godotenv/autoload"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const FETCH_GUILD_COUNTS = `-- name: FetchGuildCounts :many
SELECT
    g.guild_id,
    GREATEST(NULLIF(CAST(e.data ->> 'member_count' AS INTEGER), 0) + GREATEST(NULLIF(counts.event_count, 0) - 1, 0), 0) member_count
FROM
    guilds g
LEFT JOIN LATERAL (
    SELECT *
    FROM science_guild_events e
    WHERE e.guild_id = g.guild_id AND e.event_type = 1
    ORDER BY e.created_at DESC
    LIMIT 1
) e ON true
LEFT JOIN (
    SELECT guild_id, COUNT(*) AS event_count
    FROM science_guild_events
    WHERE event_type = 1
    GROUP BY guild_id
) counts ON counts.guild_id = g.guild_id
WHERE g.guild_id > $1
ORDER BY g.guild_id ASC
LIMIT 10000`

type FetchGuildCountsRow struct {
	GuildID     int64 `json:"guild_id"`
	MemberCount int32 `json:"member_count"`
}

func fetchGuildCounts(ctx context.Context, db *pgx.Conn, lastGuildID int64) ([]FetchGuildCountsRow, error) {
	rows, err := db.Query(ctx, FETCH_GUILD_COUNTS, lastGuildID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var items []FetchGuildCountsRow

	for rows.Next() {
		var item FetchGuildCountsRow

		if err := rows.Scan(&item.GuildID, &item.MemberCount); err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	return items, rows.Err()
}

func main() {
	var err error

	postgresURL := flag.String("postgresURL", os.Getenv("POSTGRES_URL"), "Postgres connection URL")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	welcomer.SetupDatabase(ctx, *postgresURL)
	welcomer.SetupGRPCConnection(os.Getenv("SANDWICH_GRPC_HOST"),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(1024*1024*1024)), // Set max message size to 1GB
	)

	welcomer.SetupLogger(os.Getenv("LOGGING_LEVEL"))
	welcomer.SetupSandwichClient()

	rows := 0
	lastGuildID := int64(0)

	db, err := pgx.Connect(ctx, *postgresURL)
	if err != nil {
		panic(fmt.Sprintf(`pgx.Connect(%s): %v`, *postgresURL, err.Error()))
	}

	for {
		guildCounts, err := fetchGuildCounts(ctx, db, lastGuildID)
		if err != nil {
			panic(err)
		}

		rows += len(guildCounts)
		println(lastGuildID, rows)

		if len(guildCounts) == 0 {
			break
		}

		for _, guildCount := range guildCounts {
			if guildCount.MemberCount == 0 {
				guild, err := welcomer.FetchGuild(ctx, discord.Snowflake(guildCount.GuildID))
				if err != nil {
					println("Failed to fetch guild:", guildCount.GuildID, err.Error())

					continue
				}

				guildCount.MemberCount = int32(guild.MemberCount)
			}

			_, err = welcomer.Queries.SetGuildMemberCount(ctx, database.SetGuildMemberCountParams{
				GuildID:     guildCount.GuildID,
				MemberCount: guildCount.MemberCount,
			})
			if err != nil {
				panic(fmt.Sprintf("SetGuildMemberCount(%d, %d): %v", guildCount.GuildID, guildCount.MemberCount, err))
			}

			lastGuildID = guildCount.GuildID
		}
	}
}
