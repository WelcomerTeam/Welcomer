package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"runtime/debug"
	"time"

	"github.com/WelcomerTeam/Discord/discord"
	sandwich "github.com/WelcomerTeam/Sandwich-Daemon/proto"
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

	webhookUrl := flag.String("webhookUrl", os.Getenv("JOB_CLEANUP_CUSTOM_BOTS_WEBHOOK_URL"), "Webhook URL for logging")

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
						Title:       "Cleanup Custom Bots Job",
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

	welcomer.SetupLogger(*loggingLevel)
	welcomer.SetupGRPCConnection(*sandwichGRPCHost,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(1024*1024*1024)), // Set max message size to 1GB
	)
	welcomer.SetupSandwichClient()
	welcomer.SetupDatabase(ctx, *postgresURL)

	entrypoint(ctx, *webhookUrl)

	if err := welcomer.Queries.UpsertJobCheckpoint(ctx, database.UpsertJobCheckpointParams{
		JobName:         "cleanup-custom-bots",
		LastProcessedTs: time.Now().UTC(),
	}); err != nil {
		welcomer.Logger.Error().Err(err).Msg("Failed to upsert job checkpoint")
	}

	cancel()
}

func entrypoint(ctx context.Context, webhookUrl string) {
	applications, err := welcomer.SandwichClient.FetchApplication(ctx, &sandwich.ApplicationIdentifier{
		ApplicationIdentifier: "",
	})
	if err != nil {
		panic(fmt.Errorf("failed to fetch applications: %w", err))
	}

	guildApplicationsByGuild := map[discord.Snowflake][]*sandwich.SandwichApplication{}

	for _, application := range applications.GetApplications() {
		applicationValues := welcomer.ApplicationValues{}

		// Unmarshal the application values.
		err = json.Unmarshal(application.GetValues(), &applicationValues)
		if err != nil {
			welcomer.Logger.Error().Err(err).Msg("Failed to unmarshal application values")
		}

		if !applicationValues.IsCustomBot {
			welcomer.Logger.Debug().Str("application_identifier", application.GetApplicationIdentifier()).
				Msg("Skipping non-custom bot application")
			continue // Skip non-custom bots
		}

		apps, ok := guildApplicationsByGuild[discord.Snowflake(applicationValues.GuildID)]
		if !ok {
			apps = []*sandwich.SandwichApplication{}
		}

		apps = append(apps, application)

		guildApplicationsByGuild[discord.Snowflake(applicationValues.GuildID)] = apps
	}

	for guildID, applications := range guildApplicationsByGuild {
		guildLimit := welcomer.GetGuildCustomBotLimit(ctx, guildID)

		if len(applications) > guildLimit {
			welcomer.Logger.Warn().Int64("guild_id", int64(guildID)).
				Int("application_count", len(applications)).
				Int("guild_limit", guildLimit).
				Msg("Guild has more custom bots than allowed by the limit")

			botsToKill := applications[guildLimit:]

			for _, bot := range botsToKill {
				welcomer.Logger.Info().Str("application_identifier", bot.GetApplicationIdentifier()).
					Msg("Killing custom bot due to exceeding guild limit")

				// Stop the custom bot
				_, err := welcomer.SandwichClient.StopApplication(ctx, &sandwich.ApplicationIdentifierWithBlocking{
					ApplicationIdentifier: bot.GetApplicationIdentifier(),
					Blocking:              true,
				})
				if err != nil {
					welcomer.Logger.Error().Err(err).Str("application_identifier", bot.GetApplicationIdentifier()).
						Msg("Failed to stop custom bot")
				}
			}
		} else {
			for _, application := range applications {
				welcomer.Logger.Info().Str("application_identifier", application.GetApplicationIdentifier()).
					Msg("Custom bot is within guild limit, no action needed")
			}
		}
	}
}
