package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/WelcomerTeam/Discord/discord"
	subway "github.com/WelcomerTeam/Subway/subway"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	interactions "github.com/WelcomerTeam/Welcomer/welcomer-interactions"
	_ "github.com/joho/godotenv/autoload"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	loggingLevel := flag.String("level", os.Getenv("LOGGING_LEVEL"), "Logging level")

	sandwichGRPCHost := flag.String("sandwichGRPCHost", os.Getenv("SANDWICH_GRPC_HOST"), "GRPC Address for the Sandwich Daemon service")

	proxyAddress := flag.String("proxyAddress", os.Getenv("PROXY_ADDRESS"), "Address to proxy requests through. This can be 'https://discord.com', if one is not setup")

	proxyDebug := flag.Bool("proxyDebug", false, "Enable debugging requests to the proxy")

	prometheusAddress := flag.String("prometheusAddress", os.Getenv("INTERACTIONS_PROMETHEUS_ADDRESS"), "Prometheus address")

	postgresURL := flag.String("postgresURL", os.Getenv("POSTGRES_URL"), "Postgres connection URL")

	publicKeys := flag.String("publicKey", os.Getenv("INTERACTIONS_PUBLIC_KEY"), "Public key(s) for signature validation. Comma delimited.")

	flag.Parse()

	var err error

	ctx, _ := context.WithCancel(context.Background())

	restInterface := welcomer.NewTwilightProxy(*proxyAddress)
	restInterface.SetDebug(*proxyDebug)

	welcomer.SetupLogger(*loggingLevel)
	welcomer.SetupGRPCConnection(*sandwichGRPCHost,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(1024*1024*1024)), // Set max message size to 1GB
	)
	welcomer.SetupRESTInterface(restInterface)
	welcomer.SetupSandwichClient()
	welcomer.SetupDatabase(ctx, *postgresURL)

	app := interactions.NewWelcomer(ctx, subway.SubwayOptions{
		SandwichClient:    welcomer.SandwichClient,
		RESTInterface:     welcomer.RESTInterface,
		Logger:            slog.Default(),
		PublicKeys:        *publicKeys,
		PrometheusAddress: *prometheusAddress,
	})

	configurationGatherer := welcomer.GetConfigurationGatherer(ctx)

	configuration, err := configurationGatherer.GetConfig(ctx)
	if err != nil {
		panic(fmt.Errorf("failed to gather configuration: %w", err))
	}

	applicationCommands := app.Commands.MapApplicationCommands()

	for _, application := range configuration.Applications {
		currentUser, err := discord.GetCurrentUser(ctx, discord.NewSession("Bot "+application.BotToken, restInterface))
		if err != nil {
			slog.Error("Failed to get current user", "application", application.DisplayName, "error", err)
			continue
		}

		session := discord.NewSession("Bot "+application.BotToken, restInterface)

		now := time.Now()

		applicationCommands, err := discord.BulkOverwriteGlobalApplicationCommands(ctx, session, currentUser.ID, applicationCommands)
		if err != nil {
			slog.Error("Failed to sync commands", "applicationID", currentUser.ID, "application", application.DisplayName, "error", err)
			continue
		}

		slog.Info("Successfully synced commands", "applicationID", currentUser.ID, "application", application.DisplayName)

		_, err = welcomer.Queries.ClearInteractionCommands(ctx, int64(currentUser.ID))
		if err != nil {
			slog.Error("Failed to clear interaction commands", "applicationID", currentUser.ID, "application", application.DisplayName, "error", err)
		}

		manyInteractionCommands := make([]database.CreateManyInteractionCommandsParams, 0)

		for _, command := range applicationCommands {
			manyInteractionCommands = append(manyInteractionCommands, database.CreateManyInteractionCommandsParams{
				ApplicationID: int64(currentUser.ID),
				Command:       command.Name,
				InteractionID: int64(*command.ID),
				CreatedAt:     now,
			})
		}

		_, err = welcomer.Queries.CreateManyInteractionCommands(ctx, manyInteractionCommands)
		if err != nil {
			slog.Error("Failed to create many interaction commands", "applicationID", currentUser.ID, "application", application.DisplayName, "error", err)
			continue
		}

		slog.Info("Successfully created interaction commands", "applicationID", currentUser.ID, "application", application.DisplayName)
	}
}
