package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/WelcomerTeam/Discord/discord"
	sandwich_protobuf "github.com/WelcomerTeam/Sandwich-Daemon/proto"
	subway "github.com/WelcomerTeam/Subway/subway"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	interactions "github.com/WelcomerTeam/Welcomer/welcomer-interactions"
	_ "github.com/joho/godotenv/autoload"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	PermissionsDefault = 0o744
)

func main() {
	loggingLevel := flag.String("level", os.Getenv("LOGGING_LEVEL"), "Logging level")

	sandwichManagerName := flag.String("sandwichManagerName", os.Getenv("SANDWICH_MANAGER_NAME"), "Sandwich manager identifier name")

	sandwichGRPCHost := flag.String("sandwichGRPCHost", os.Getenv("SANDWICH_GRPC_HOST"), "GRPC Address for the Sandwich Daemon service")

	proxyAddress := flag.String("proxyAddress", os.Getenv("PROXY_ADDRESS"), "Address to proxy requests through. This can be 'https://discord.com', if one is not setup")

	proxyDebug := flag.Bool("proxyDebug", false, "Enable debugging requests to the proxy")

	prometheusAddress := flag.String("prometheusAddress", os.Getenv("INTERACTIONS_PROMETHEUS_ADDRESS"), "Prometheus address")

	postgresURL := flag.String("postgresURL", os.Getenv("POSTGRES_URL"), "Postgres connection URL")

	host := flag.String("host", os.Getenv("INTERACTIONS_HOST"), "Host to serve interactions from")

	publicKeys := flag.String("publicKey", os.Getenv("INTERACTIONS_PUBLIC_KEY"), "Public key(s) for signature validation. Comma delimited.")

	dryRun := flag.Bool("dryRun", false, "When true, will close after setting up the app")

	syncCommands := flag.Bool("syncCommands", false, "If true, will update commands")

	flag.Parse()

	var err error

	ctx, cancel := context.WithCancel(context.Background())

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

	// Setup app.

	app := interactions.NewWelcomer(ctx, subway.SubwayOptions{
		SandwichClient:    welcomer.SandwichClient,
		RESTInterface:     welcomer.RESTInterface,
		Logger:            slog.Default(),
		PublicKeys:        *publicKeys,
		PrometheusAddress: *prometheusAddress,
	})

	if err != nil {
		welcomer.Logger.Panic().Err(err).Msg("Exception creating app")
	}

	if *syncCommands {
		configurationsPb, err := welcomer.SandwichClient.FetchApplication(ctx, &sandwich_protobuf.ApplicationIdentifier{
			ApplicationIdentifier: *sandwichManagerName,
		})
		if err != nil {
			panic(fmt.Errorf(`failed to sync command: grpcInterface.FetchConsumerConfiguration(): %w`, err))
		}

		configuration, ok := configurationsPb.GetApplications()[*sandwichManagerName]

		if !ok {
			panic(fmt.Errorf(`failed to sync command: could not find manager matching "%s"`, *sandwichManagerName))
		}

		err = app.SyncCommands(ctx, "Bot "+configuration.BotToken, discord.Snowflake(configuration.UserId))
		if err != nil {
			panic(fmt.Errorf(`failed to sync commands. app.SyncCommands(): %w`, err))
		}

		welcomer.Logger.Info().Msg("Successfully synced commands")

	}

	// We return if it a dry run. Any issues loading up the bot would've already caused a panic.

	if *dryRun {
		return
	}

	if err = app.ListenAndServe("", *host); err != nil {
		welcomer.Logger.Panic().Err(err).Msg("Exceptions whilst starting app")
	}

	cancel()

	if err = welcomer.GRPCConnection.Close(); err != nil {
		welcomer.Logger.Warn().Err(err).Msg("Exception whilst closing grpc client")
	}
}
