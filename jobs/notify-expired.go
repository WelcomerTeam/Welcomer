package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/WelcomerTeam/Discord/discord"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	"github.com/jackc/pgx/v4/pgxpool"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	var err error

	loggingLevel := flag.String("level", os.Getenv("LOGGING_LEVEL"), "Logging level")
	postgresURL := flag.String("postgresURL", os.Getenv("POSTGRES_URL"), "Postgres connection URL")

	webhookUrl := flag.String("patreonWebhookUrl", os.Getenv("JOB_NOTIFY_EXPIRED_WEBHOOK_URL"), "Webhook URL for logging")

	dryRun := flag.Bool("dryRun", false, "When true, will close after setting up the app")

	flag.Parse()

	welcomer.SetupLogger(*loggingLevel)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from panic:", r)

			err = welcomer.SendWebhookMessage(ctx, *webhookUrl, discord.WebhookMessageParams{
				Content: "<@143090142360371200>",
				Embeds: []discord.Embed{
					{
						Title:       "Notify Expired Job",
						Description: fmt.Sprintf("Recovered from panic: %v", r),
						Color:       int32(16760839),
						Timestamp:   welcomer.ToPointer(time.Now()),
					},
				},
			})
			if err != nil {
				welcomer.Logger.Warn().Err(err).Msg("Failed to send webhook message")
			}
		}
	}()

	// Setup postgres pool.
	var pool *pgxpool.Pool
	if pool, err = pgxpool.Connect(ctx, *postgresURL); err != nil {
		panic(fmt.Sprintf(`pgxpool.Connect(%s): %v`, *postgresURL, err.Error()))
	}

	// Setup database.
	db := database.New(pool)

	entrypoint(ctx, db, *webhookUrl, *dryRun)
}

func entrypoint(ctx context.Context, db *database.Queries, webhookUrl string, dryRun bool) {
	memberships, err := db.GetExpiringUserMemberships(ctx, int32(database.MembershipStatusExpired))
	if err != nil {
		panic(fmt.Sprintf("GetExpiringUserMemberships: %v", err))
	}

	fmt.Printf("Expiring memberships: %d\n", len(memberships))

	for _, membership := range memberships {
		println(membership.MembershipUuid.String(), membership.UserID, membership.StartedAt.String(), membership.ExpiresAt.String(), database.MembershipType(membership.MembershipType).Label(), database.MembershipStatus(membership.Status).Label(), membership.GuildID)

		// session, err := welcomer.AcquireSession(ctx, nil, nil, nil, "")
		// if err != nil {
		// 	welcomer.Logger.Error().Err(err).
		// 		Str("membership_uuid", membership.MembershipUuid.String()).
		// 		Msg("Failed to acquire session")
		// } else {
		// 	err = notifyMembershipCreated(ctx, queries, session, membership)
		// 	if err != nil {
		// 		welcomer.Logger.Error().Err(err).
		// 			Str("membership_uuid", membership.MembershipUuid.String()).
		// 			Msg("Failed to trigger onMembershipAdded")
		// 	}
		// }

		// DM the user
		// Mark the transaction as expired
	}

	// Notify of new membership?

	if dryRun {
		return
	}
}
