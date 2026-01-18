package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"time"

	sandwich "github.com/WelcomerTeam/Sandwich/sandwich"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	jetstream_client "github.com/WelcomerTeam/Welcomer/welcomer-core/jetstream"
	gateway "github.com/WelcomerTeam/Welcomer/welcomer-gateway"
	_ "github.com/joho/godotenv/autoload"
	"github.com/nats-io/nats.go/jetstream"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	ScienceIngestFlushInterval = 30 * time.Second
	ScienceIngestBufferSize    = 1024

	MessageIngestFlushInterval = 10 * time.Second
	MessageIngestBufferSize    = 4096
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
	redisHost := flag.String("redisHost", os.Getenv("REDIS_HOST"), "Redis host")
	dryRun := flag.Bool("dryRun", false, "When true, will close after setting up the app")

	flag.Parse()

	var err error

	ctx, cancel := context.WithCancel(context.Background())

	restInterface := welcomer.NewTwilightProxy(*proxyAddress)
	restInterface.SetDebug(*proxyDebug)

	slog.SetLogLoggerLevel(slog.LevelDebug)

	welcomer.SetupLogger(*loggingLevel)
	welcomer.SetupGRPCConnection(*sandwichGRPCHost,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(1024*1024*1024)), // Set max message size to 1GB
	)
	welcomer.SetupRESTInterface(restInterface)
	welcomer.SetupSandwichClient()
	welcomer.SetupDatabase(ctx, *postgresURL)

	welcomer.SetupRedisClient(*redisHost)
	welcomer.SetupDedupeProvider(welcomer.NewRedisDedupeProvider(welcomer.RedisClient, slog.Default()))

	jetstreamClient, err := jetstream_client.SetupJetstreamConsumer(
		ctx,
		*stanAddress,
		*stanChannel,
		*jetstreamClientName,
		nil,
		nil,
	)
	if err != nil {
		panic(fmt.Errorf(`jetstream_client.SetupJetstreamConsumer(): %w`, err))
	}

	eventsChannel := make(chan []byte)

	consumeContext, err := jetstreamClient.Consume(func(msg jetstream.Msg) {
		msg.Ack()
		eventsChannel <- msg.Data()
	})
	if err != nil {
		panic(fmt.Errorf("jetstreamClient.Consume(): %w", err))
	}

	pusherGuildScience := welcomer.SetupPusherGuildScience(ScienceIngestBufferSize)
	pusherGuildScience(ctx, ScienceIngestFlushInterval)

	pusherIngestMessageEvents := welcomer.SetupPusherIngestMessageEvents(MessageIngestBufferSize)
	pusherIngestMessageEvents(ctx, MessageIngestFlushInterval)

	// Setup sandwich.

	sandwichClient := sandwich.NewSandwich(welcomer.GRPCConnection, restInterface, os.Stdout)
	sandwichClient.SetErrorOnInvalidIdentifier(true)

	bot := gateway.NewWelcomer(*sandwichProducerName, sandwichClient)
	sandwichClient.RegisterBot(*sandwichProducerName, bot.Bot)

	// We return if it a dry run. Any issues loading up the bot would've already caused a panic.

	if *dryRun {
		return
	}

	// Debug: Print worker queue sizes every second
	go func() {
		for {
			time.Sleep(time.Second * 1)
			sizes := [][2]int{}
			sum := 0

			bot.Bot.Handlers.WorkerPoolMu.Lock()
			for i, ch := range bot.Bot.Handlers.WorkerPool {
				sizes = append(sizes, [2]int{int(i), ch.Len()})
				sum += sizes[len(sizes)-1][1]
			}

			// sort sizes
			slices.SortFunc(sizes, func(a, b [2]int) int {
				return a[1] - b[1]
			})

			totalSize := len(bot.Bot.Handlers.WorkerPool)
			bot.Bot.Handlers.WorkerPoolMu.Unlock()

			println("==================")
			println(fmt.Sprintf("Worker queue sizes: %d", totalSize))
			for _, size := range sizes {
				if size[1] == 0 {
					continue
				}
				println(fmt.Sprintf("Worker %d: %d messages", size[0], size[1]))
			}
			println(fmt.Sprintf("Total queued messages: %d", sum))
			println("==================")
		}
	}()

	// Register message channels

	if err = sandwichClient.ListenToChannel(ctx, eventsChannel); err != nil {
		welcomer.Logger.Panic().Err(err).Msg("Failed to listen to channel")
	}

	consumeContext.Drain()

	if err = welcomer.GRPCConnection.Close(); err != nil {
		welcomer.Logger.Warn().Err(err).Msg("Exception whilst closing grpc client")
	}

	welcomer.PusherGuildScience.Flush(ctx)
	welcomer.PusherIngestMessageEvents.Flush(ctx)

	cancel()
}
