package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-images/service"
	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	loggingLevel := flag.String("level", os.Getenv("LOGGING_LEVEL"), "Logging level")

	prometheusAddress := flag.String("prometheusAddress", os.Getenv("IMAGE_PROMETHEUS_ADDRESS"), "Prometheus address")
	postgresURL := flag.String("postgresURL", os.Getenv("POSTGRES_URL"), "Postgres connection URL")
	imageHost := flag.String("host", os.Getenv("IMAGE_HOST"), "Host to serve the image service interface from")

	releaseMode := flag.String("ginMode", os.Getenv("GIN_MODE"), "gin mode (release/debug)")
	debug := flag.Bool("debug", false, "When enabled, images will be saved to a file.")

	flag.Parse()

	gin.SetMode(*releaseMode)

	var err error

	ctx, cancel := context.WithCancel(context.Background())

	welcomer.SetupLogger(*loggingLevel)
	welcomer.SetupSandwichClient()
	welcomer.SetupDatabase(ctx, *postgresURL)

	// Image Service initialization
	var imageService *service.ImageService
	if imageService, err = service.NewImageService(ctx, service.ImageServiceOptions{
		Debug:             *debug,
		Host:              *imageHost,
		PostgresAddress:   *postgresURL,
		PrometheusAddress: *prometheusAddress,
	}); err != nil {
		welcomer.Logger.Panic().Err(err).Msg("Cannot create image service")
	}

	imageService.Open()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-signalCh

	cancel()

	if err = imageService.Close(); err != nil {
		welcomer.Logger.Warn().Err(err).Msg("Exception whilst closing image service")
	}
}
