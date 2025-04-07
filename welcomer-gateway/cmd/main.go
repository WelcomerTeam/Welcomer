package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	messaging "github.com/WelcomerTeam/Sandwich/messaging"
	sandwich "github.com/WelcomerTeam/Sandwich/sandwich"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	gateway "github.com/WelcomerTeam/Welcomer/welcomer-gateway"
	_ "github.com/joho/godotenv/autoload"
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

	ctx, cancel := context.WithCancel(context.Background())

	restInterface := welcomer.NewTwilightProxy(*proxyAddress)
	restInterface.SetDebug(*proxyDebug)

	writer := welcomer.SetupLogger(*loggingLevel)
	welcomer.SetupGRPCConnection(*sandwichGRPCHost,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(1024*1024*1024)), // Set max message size to 1GB
	)
	welcomer.SetupGRPCInterface()
	welcomer.SetupRESTInterface(restInterface)
	welcomer.SetupSandwichClient()
	welcomer.SetupDatabase(ctx, *postgresURL)

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

	runPushGuildScience := welcomer.SetupPushGuildScience(1024)
	runPushGuildScience(ctx, time.Second*30)

	// Setup sandwich.

	sandwichClient := sandwich.NewSandwich(welcomer.GRPCConnection, restInterface, writer)

	bot := gateway.NewWelcomer(*sandwichProducerName, sandwichClient)
	sandwichClient.RegisterBot(*sandwichProducerName, bot.Bot)

	// We return if it a dry run. Any issues loading up the bot would've already caused a panic.

	if *dryRun {
		return
	}

	// Register message channels

	stanMessages := jetstreamClient.Chan()
	if err = sandwichClient.ListenToChannel(ctx, stanMessages); err != nil {
		welcomer.Logger.Panic().Err(err).Msg("Failed to listen to channel")
	}

	welcomer.PushGuildScience.Flush(ctx)
	jetstreamClient.Unsubscribe(ctx)

	if err = welcomer.GRPCConnection.Close(); err != nil {
		welcomer.Logger.Warn().Err(err).Msg("Exception whilst closing grpc client")
	}

	cancel()
}
