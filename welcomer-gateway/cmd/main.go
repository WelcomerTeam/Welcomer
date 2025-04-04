package main

import (
	"context"
	"flag"
	"fmt"
	"net/url"
	"os"
	"time"

	messaging "github.com/WelcomerTeam/Sandwich/messaging"
	sandwich "github.com/WelcomerTeam/Sandwich/sandwich"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	gateway "github.com/WelcomerTeam/Welcomer/welcomer-gateway"
	"github.com/jackc/pgx/v4/pgxpool"
	_ "github.com/joho/godotenv/autoload"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

	jetstreamClientName := flag.String("jetstreamClientName", "welcomer-gateway", "NATs client name")

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
		Out: os.Stdout,

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

	// Setup NATs

	jetstreamClient := messaging.NewJetstreamMQClient()

	if err = jetstreamClient.Connect(ctx, *jetstreamClientName, map[string]any{
		"Address": *stanAddress,

		"Channel": *stanChannel,
	}); err != nil {
		panic(fmt.Sprintf(`jetstreamClient.Connect(): %v`, err.Error()))
	}

	if err = jetstreamClient.Subscribe(ctx, *stanChannel); err != nil {
		panic(fmt.Sprintf(`jetstreamClient.Subscribe(%s): %v`, *stanChannel, err.Error()))
	}

	// Setup postgres pool.

	pool, err := pgxpool.Connect(ctx, *postgresURL)
	if err != nil {
		panic(fmt.Sprintf("pgxpool.Connect(%s): %v", *postgresURL, err))
	}

	queries := database.New(pool)

	pushGuildScienceHandler := welcomer.NewPushGuildScienceHandler(queries, logger, 1024)
	pushGuildScienceHandler.Run(ctx, time.Second*30)

	ctx = welcomer.AddPushGuildScienceToContext(ctx, pushGuildScienceHandler)

	ctx = welcomer.AddPoolToContext(ctx, pool)

	ctx = welcomer.AddQueriesToContext(ctx, queries)

	// Setup sandwich.

	sandwichClient := sandwich.NewSandwich(grpcConnection, restInterface, writer)

	bot := gateway.NewWelcomer(*sandwichProducerName, sandwichClient)

	sandwichClient.RegisterBot(*sandwichProducerName, bot.Bot)

	ctx = welcomer.AddSandwichClientToContext(ctx, sandwichClient.SandwichClient)

	ctx = welcomer.AddGRPCInterfaceToContext(ctx, sandwichClient.GRPCInterface)

	ctx = welcomer.AddRESTInterfaceToContext(ctx, sandwichClient.RESTInterface)

	// We return if it a dry run. Any issues loading up the bot would've already caused a panic.

	if *dryRun {
		return
	}

	// Register message channels

	stanMessages := jetstreamClient.Chan()

	if err = sandwichClient.ListenToChannel(ctx, stanMessages); err != nil {
		logger.Panic().Err(err).Msg("Failed to listen to channel")
	}

	pushGuildScienceHandler.Flush(ctx)

	jetstreamClient.Unsubscribe(ctx)

	if err = grpcConnection.Close(); err != nil {
		logger.Warn().Err(err).Msg("Exception whilst closing grpc client")
	}

	// Close sandwich
	cancel()
}
