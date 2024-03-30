package main

import (
	"context"
	"flag"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/WelcomerTeam/Discord/discord"
	messaging "github.com/WelcomerTeam/Sandwich/messaging"
	sandwich "github.com/WelcomerTeam/Sandwich/sandwich"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	gateway "github.com/WelcomerTeam/Welcomer/welcomer-gateway"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	loggingLevel := flag.String("level", os.Getenv("LOGGING_LEVEL"), "Logging level")

	sandwichGRPCHost := flag.String("grpcAddress", os.Getenv("SANDWICH_GRPC_HOST"), "GRPC Address for the Sandwich Daemon service")
	sandwichProducerName := flag.String("producerName", os.Getenv("SANDWICH_PRODUCER_NAME"), "Sandwich producer identifier name")
	proxyAddress := flag.String("proxyAddress", os.Getenv("PROXY_ADDRESS"), "Address to proxy requests through. This can be 'https://discord.com', if one is not setup.")
	proxyDebug := flag.Bool("proxyDebug", false, "Enable debugging requests to the proxy")
	postgresURL := flag.String("postgresURL", os.Getenv("POSTGRES_URL"), "Postgres connection URL")

	stanAddress := flag.String("stanAddress", os.Getenv("STAN_ADDRESS"), "NATs streaming Address")
	stanChannel := flag.String("stanChannel", os.Getenv("STAN_CHANNEL"), "NATs streaming Channel")
	stanCluster := flag.String("stanCluster", os.Getenv("STAN_CLUSTER"), "NATs streaming Cluster")
	stanClientName := flag.String("stanClientName", "welcomer-gateway", "NATs client name")

	dryRun := flag.Bool("dryRun", false, "When true, will close after setting up the app")

	flag.Parse()

	var err error

	// Setup Logger
	var level zerolog.Level
	if level, err = zerolog.ParseLevel(*loggingLevel); err != nil {
		panic(fmt.Errorf(`failed to parse loggingLevel. zerolog.ParseLevel(%s): %w`, *loggingLevel, err))
	}

	zerolog.SetGlobalLevel(level)

	writer := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.Stamp,
	}

	logger := zerolog.New(writer).With().Timestamp().Logger()
	logger.Info().Msg("Logging configured")

	ctx, cancel := context.WithCancel(context.Background())

	// Setup Rest
	var proxyURL *url.URL
	if proxyURL, err = url.Parse(*proxyAddress); err != nil {
		panic(fmt.Errorf("failed to parse proxy address. url.Parse(%s): %w", *proxyAddress, err))
	}

	restInterface := discord.NewTwilightProxy(*proxyURL)
	restInterface.SetDebug(*proxyDebug)

	// Setup GRPC
	var grpcConnection *grpc.ClientConn
	if grpcConnection, err = grpc.Dial(*sandwichGRPCHost, grpc.WithTransportCredentials(insecure.NewCredentials())); err != nil {
		panic(fmt.Sprintf(`grpc.Dial(%s): %v`, *sandwichGRPCHost, err.Error()))
	}

	// Setup NATs
	stanClient := messaging.NewStanMQClient()

	if err = stanClient.Connect(ctx, *stanClientName, map[string]interface{}{
		"Address": *stanAddress,
		"Cluster": *stanCluster,
		"Channel": *stanChannel,
	}); err != nil {
		panic(fmt.Sprintf(`stanClient.Connect(): %v`, err.Error()))
	}

	if err = stanClient.Subscribe(ctx, *stanChannel); err != nil {
		panic(fmt.Sprintf(`stanClient.Subscribe(%s): %v`, *stanChannel, err.Error()))
	}

	// Setup postgres pool.
	pool, err := pgxpool.Connect(ctx, *postgresURL)
	if err != nil {
		panic(fmt.Sprintf("pgxpool.Connect(%s): %v", *postgresURL, err))
	}

	queries := database.New(pool)

	ctx = welcomer.AddPoolToContext(ctx, pool)
	ctx = welcomer.AddQueriesToContext(ctx, queries)

	// Setup sandwich.
	sandwichClient := sandwich.NewSandwich(grpcConnection, restInterface, writer)

	welcomer := gateway.NewWelcomer(*sandwichProducerName, sandwichClient)
	sandwichClient.RegisterBot(*sandwichProducerName, welcomer.Bot)

	// We return if it a dry run. Any issues loading up the bot would've already caused a panic.
	if *dryRun {
		return
	}

	// Register message channels
	stanMessages := stanClient.Chan()

	if err = sandwichClient.ListenToChannel(ctx, stanMessages); err != nil {
		logger.Panic().Err(err).Msg("Failed to listen to channel")
	}

	cancel()

	// Close sandwich
	stanClient.Unsubscribe()

	if err = grpcConnection.Close(); err != nil {
		logger.Warn().Err(err).Msg("Exception whilst closing grpc client")
	}
}
