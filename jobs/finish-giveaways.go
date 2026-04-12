package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"runtime/debug"
	"time"

	"github.com/WelcomerTeam/Discord/discord"
	sandwich_protobuf "github.com/WelcomerTeam/Sandwich-Daemon/proto"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	"github.com/jackc/pgx/v4"
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

	webhookUrl := flag.String("webhookUrl", os.Getenv("JOB_CLEANUP_CUSTOM_BOTS_WEBHOOK_URL"), "Webhook URL for logging")

	sandwichManagerName := flag.String("sandwichManagerName", os.Getenv("SANDWICH_MANAGER_NAME"), "Sandwich manager identifier name")

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
						Title:       "Finish Giveaways Job",
						Description: fmt.Sprintf("Recovered from panic: %v", r),
						Color:       int32(16760839),
						Timestamp:   new(time.Now()),
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

	entrypoint(ctx, *webhookUrl)

	if err := welcomer.Queries.UpsertJobCheckpoint(ctx, database.UpsertJobCheckpointParams{
		JobName:         "finish-giveaways",
		LastProcessedTs: time.Now().UTC(),
	}); err != nil {
		welcomer.Logger.Error().Err(err).Msg("Failed to upsert job checkpoint")
	}

	cancel()
}

func entrypoint(ctx context.Context, webhookUrl string) {
	expiredGiveaways, err := welcomer.Queries.GetExpiredGiveaways(ctx)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		welcomer.Logger.Error().Err(err).Msg("Failed to fetch expired giveaways")

		panic(err)
	}

	for _, giveaway := range expiredGiveaways {
		locationsPb, err := welcomer.SandwichClient.WhereIsGuild(ctx, &sandwich_protobuf.WhereIsGuildRequest{
			GuildId: giveaway.GuildID,
		})
		if err != nil {
			welcomer.Logger.Warn().Err(err).Int64("guild_id", giveaway.GuildID).Msg("Failed to do guild lookup for giveaway")

			continue
		}

		locations := locationsPb.GetLocations()
		if len(locations) == 0 {
			welcomer.Logger.Warn().Int64("guild_id", giveaway.GuildID).Msg("No applications found for guild in giveaway")

			continue
		}

		data, _ := json.Marshal(welcomer.CustomEventInvokeEndGiveawayStructure{
			GiveawayUUID: giveaway.GiveawayUuid,
			GuildID:      discord.Snowflake(giveaway.GuildID),
		})

		for _, location := range locations {
			_, err = welcomer.SandwichClient.RelayMessage(ctx, &sandwich_protobuf.RelayMessageRequest{
				Identifier: location.GetIdentifier(),
				Type:       welcomer.CustomEventInvokeEndGiveaway,
				Data:       data,
			})
			if err != nil {
				welcomer.Logger.Warn().Err(err).Int64("guild_id", giveaway.GuildID).Str("identifier", location.GetIdentifier()).Msg("Failed to relay end giveaway message")

				continue
			}

			if err == nil {
				break
			}
		}

		welcomer.Logger.Info().Int64("guild_id", giveaway.GuildID).Msg("Finished giveaway")
	}
}
