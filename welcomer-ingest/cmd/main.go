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

	"github.com/WelcomerTeam/Welcomer/welcomer-core"
)

const JobInterval = time.Minute * 1

func main() {
	loggingLevel := flag.String("level", os.Getenv("LOGGING_LEVEL"), "Logging level")
	postgresURL := flag.String("postgresURL", os.Getenv("POSTGRES_URL"), "Postgres connection URL")

	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())

	welcomer.SetupLogger(*loggingLevel)
	welcomer.SetupDatabase(ctx, *postgresURL)

	waitGroup := &sync.WaitGroup{}

	ingest.AggregateMessageCounts(ctx, waitGroup, JobInterval)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	cancel()

	welcomer.Logger.Info().Msg("waiting for jobs to finish")

	waitGroup.Wait()
}
