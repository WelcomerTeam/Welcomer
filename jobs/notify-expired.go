package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"runtime/debug"
	"time"

	"github.com/WelcomerTeam/Discord/discord"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	_ "github.com/joho/godotenv/autoload"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	var err error

	loggingLevel := flag.String("level", os.Getenv("LOGGING_LEVEL"), "Logging level")

	postgresURL := flag.String("postgresURL", os.Getenv("POSTGRES_URL"), "Postgres connection URL")
	sandwichGRPCHost := flag.String("sandwichGRPCHost", os.Getenv("SANDWICH_GRPC_HOST"), "GRPC Address for the Sandwich Daemon service")

	proxyAddress := flag.String("proxyAddress", os.Getenv("PROXY_ADDRESS"), "Address to proxy requests through. This can be 'https://discord.com', if one is not setup.")
	proxyDebug := flag.Bool("proxyDebug", false, "Enable debugging requests to the proxy")

	webhookUrl := flag.String("webhookUrl", os.Getenv("JOB_NOTIFY_EXPIRED_WEBHOOK_URL"), "Webhook URL for logging")

	sandwichManagerName := flag.String("sandwichManagerName", os.Getenv("SANDWICH_MANAGER_NAME"), "Sandwich manager identifier name")

	dryRun := flag.Bool("dryRun", false, "When true, will close after setting up the app")

	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from panic:", r)
			println(string(debug.Stack()))

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

	restInterface := welcomer.NewTwilightProxy(*proxyAddress)
	restInterface.SetDebug(*proxyDebug)

	welcomer.SetupDefaultManagerName(*sandwichManagerName)
	welcomer.SetupLogger(*loggingLevel)
	welcomer.SetupGRPCConnection(*sandwichGRPCHost,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(1024*1024*1024)), // Set max message size to 1GB
	)
	welcomer.SetupRESTInterface(restInterface)
	welcomer.SetupSandwichClient()
	welcomer.SetupDatabase(ctx, *postgresURL)

	entrypoint(ctx, *webhookUrl, *dryRun)

	cancel()
}

func entrypoint(ctx context.Context, webhookUrl string, dryRun bool) {
	memberships, err := welcomer.Queries.GetExpiringUserMemberships(ctx, int32(database.MembershipStatusExpired))
	if err != nil {
		panic(fmt.Sprintf("GetExpiringUserMemberships: %v", err))
	}

	fmt.Printf("Expiring memberships: %d\n", len(memberships))

	for _, membership := range memberships {
		println(membership.MembershipUuid.String(), membership.UserID, membership.StartedAt.String(), membership.ExpiresAt.String(), database.MembershipType(membership.MembershipType).Label(), database.MembershipStatus(membership.Status).Label(), membership.GuildID)
	}

	if dryRun {
		return
	}

	for _, membership := range memberships {
		session, err := welcomer.AcquireSession(ctx, welcomer.DefaultManagerName)
		if err != nil {
			welcomer.Logger.Error().Err(err).
				Str("membership_uuid", membership.MembershipUuid.String()).
				Msg("Failed to acquire session")

			continue
		}

		err = welcomer.NotifyMembershipExpired(ctx, session, *membership)
		if err != nil {
			welcomer.Logger.Warn().Err(err).
				Str("membership_uuid", membership.MembershipUuid.String()).
				Msg("Failed to trigger onMembershipExpired")
		}

		membership.Status = int32(database.MembershipStatusExpired)

		_, err = welcomer.Queries.UpdateUserMembership(ctx, database.UpdateUserMembershipParams{
			MembershipUuid:  membership.MembershipUuid,
			StartedAt:       membership.StartedAt,
			ExpiresAt:       membership.ExpiresAt,
			Status:          membership.Status,
			TransactionUuid: membership.TransactionUuid,
			UserID:          membership.UserID,
			GuildID:         membership.GuildID,
		})
		if err != nil {
			welcomer.Logger.Error().Err(err).
				Str("membership_uuid", membership.MembershipUuid.String()).
				Msg("Failed to update membership")

			continue
		}
	}

	welcomer.Logger.Info().
		Int("memberships_count", len(memberships)).
		Msg("Expired memberships notified successfully")
}
