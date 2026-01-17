package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	ingest "github.com/WelcomerTeam/Welcomer/welcomer-ingest/ingest"
	_ "github.com/joho/godotenv/autoload"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/WelcomerTeam/Welcomer/welcomer-core"
)

const (
	IngestBufferSize = 1024
)

func main() {
	loggingLevel := flag.String("level", os.Getenv("LOGGING_LEVEL"), "Logging level")
	postgresURL := flag.String("postgresURL", os.Getenv("POSTGRES_URL"), "Postgres connection URL")
	sandwichGRPCHost := flag.String("sandwichGRPCHost", os.Getenv("SANDWICH_GRPC_HOST"), "GRPC Address for the Sandwich Daemon service")

	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())

	welcomer.SetupLogger(*loggingLevel)
	welcomer.SetupGRPCConnection(*sandwichGRPCHost,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(1024*1024*1024)), // Set max message size to 1GB
	)
	welcomer.SetupSandwichClient()
	welcomer.SetupDatabase(ctx, *postgresURL)

	waitGroup := &sync.WaitGroup{}

	ingest.AggregateMessageCounts(ctx, waitGroup, time.Minute*1)
	ingest.AggregateVoiceChannels(ctx, waitGroup, time.Minute*1)

	pusherIngestVoicechannelEvents := welcomer.SetupPusherIngestVoiceChannelEvents(IngestBufferSize)
	pusherIngestVoicechannelEvents(ctx, time.Minute*1)

	ingest.CheckpointVoiceChannels(ctx, waitGroup, time.Minute*1)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	welcomer.Logger.Info().Msg("flushing")

	welcomer.PusherIngestVoiceChannelEvents.Flush(ctx)

	welcomer.Logger.Info().Msg("waiting for jobs to finish")

	cancel()
	waitGroup.Wait()
}
