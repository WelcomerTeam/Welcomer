package main

import (
	"context"
	"flag"
	"fmt"
	protobuf "github.com/WelcomerTeam/Sandwich-Daemon/protobuf"
	sandwich "github.com/WelcomerTeam/Sandwich/sandwich"
	subway "github.com/WelcomerTeam/Subway/subway"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	interactions "github.com/WelcomerTeam/Welcomer/welcomer-interactions"
	"github.com/jackc/pgx/v4/pgxpool"
	_ "github.com/joho/godotenv/autoload"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net/url"
	"os"
	"time"
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

	// Setup Logger

	var level zerolog.Level

	if level, err = zerolog.ParseLevel(*loggingLevel); err != nil {

		panic(fmt.Errorf(`failed to parse loggingLevel. zerolog.ParseLevel(%s): %w`, *loggingLevel, err))

	}

	zerolog.SetGlobalLevel(level)

	writer := zerolog.ConsoleWriter{

		Out: os.Stdout,

		TimeFormat: time.Stamp,
	}

	logger := zerolog.New(writer).With().Timestamp().Logger()

	logger.Info().Msg("Logging configured")

	ctx, cancel := context.WithCancel(context.Background())

	// Setup Rest

	var proxyURL *url.URL

	if proxyURL, err = url.Parse(*proxyAddress); err != nil {

		panic(fmt.Sprintf("url.Parse(%s): %v", *proxyAddress, err.Error()))

	}

	restInterface := welcomer.NewTwilightProxy(*proxyURL)

	restInterface.SetDebug(*proxyDebug)

	// Setup GRPC

	var grpcConnection *grpc.ClientConn

	if grpcConnection, err = grpc.NewClient(

		*sandwichGRPCHost,

		grpc.WithTransportCredentials(insecure.NewCredentials()),

		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(1024*1024*1024)), // Set max message size to 1GB

	); err != nil {

		panic(fmt.Sprintf(`grpc.NewClient(%s): %v`, *sandwichGRPCHost, err.Error()))

	}

	sandwichClient := protobuf.NewSandwichClient(grpcConnection)

	// Setup postgres pool.

	pool, err := pgxpool.Connect(ctx, *postgresURL)

	if err != nil {

		panic(fmt.Sprintf("pgxpool.Connect(%s): %v", *postgresURL, err))

	}

	queries := database.New(pool)

	ctx = welcomer.AddPoolToContext(ctx, pool)

	ctx = welcomer.AddQueriesToContext(ctx, queries)

	ctx = welcomer.AddManagerNameToContext(ctx, *sandwichManagerName)

	// Setup app.

	app := interactions.NewWelcomer(ctx, subway.SubwayOptions{

		SandwichClient: sandwichClient,

		RESTInterface: restInterface,

		Logger: logger,

		PublicKeys: *publicKeys,

		PrometheusAddress: *prometheusAddress,
	})

	if err != nil {

		logger.Panic().Err(err).Msg("Exception creating app")

	}

	if *syncCommands {

		grpcInterface := sandwich.NewDefaultGRPCClient()

		configurations, err := grpcInterface.FetchConsumerConfiguration(&sandwich.GRPCContext{

			Context: ctx,

			SandwichClient: sandwichClient,
		}, *sandwichManagerName)

		if err != nil {

			panic(fmt.Errorf(`failed to sync command: grpcInterface.FetchConsumerConfiguration(): %w`, err))

		}

		configuration, ok := configurations.Identifiers[*sandwichManagerName]

		if !ok {

			panic(fmt.Errorf(`failed to sync command: could not find manager matching "%s"`, *sandwichManagerName))

		}

		err = app.SyncCommands(ctx, "Bot "+configuration.Token, configuration.ID)

		if err != nil {

			panic(fmt.Errorf(`failed to sync commands. app.SyncCommands(): %w`, err))

		}

		logger.Info().Msg("Successfully synced commands")

	}

	// We return if it a dry run. Any issues loading up the bot would've already caused a panic.

	if *dryRun {

		return

	}

	if err = app.ListenAndServe("", *host); err != nil {

		logger.Panic().Err(err).Msg("Exceptions whilst starting app")

	}

	cancel()

	if err = grpcConnection.Close(); err != nil {

		logger.Warn().Err(err).Msg("Exception whilst closing grpc client")

	}

}
