package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/WelcomerTeam/Welcomer/welcomer-images/service"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	loggingLevel := flag.String("level", os.Getenv("LOGGING_LEVEL"), "Logging level")

	prometheusAddress := flag.String("prometheusAddress", os.Getenv("PROMETHEUS_ADDRESS"), "Prometheus address")
	postgresURL := flag.String("postgresURL", os.Getenv("POSTGRES_URL"), "Postgres connection URL")
	imageHost := flag.String("host", os.Getenv("IMAGE_HOST"), "Host to serve the image service interface from")

	releaseMode := flag.String("ginMode", os.Getenv("GIN_MODE"), "gin mode (release/debug)")
	debug := flag.Bool("debug", false, "When enabled, images will be saved to a file.")

	flag.Parse()

	gin.SetMode(*releaseMode)

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

	// Image Service initialization
	var imageService *service.ImageService
	if imageService, err = service.NewImageService(ctx, writer, service.ImageServiceOptions{
		PrometheusAddress: *prometheusAddress,
		PostgresAddress:   *postgresURL,

		Host:  *imageHost,
		Debug: *debug,
	}); err != nil {
		logger.Panic().Err(err).Msg("Cannot create image service")
	}

	imageService.Open()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-signalCh

	cancel()

	if err = imageService.Close(); err != nil {
		logger.Warn().Err(err).Msg("Exception whilst closing image service")
	}
}
