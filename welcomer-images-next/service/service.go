package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/gin-contrib/logger"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// VERSION follows semantic versioning.
const VERSION = "0.0.1"

type ImageService struct {
	ctx context.Context

	StartTime time.Time

	Options ImageServiceOptions

	Client http.Client
}

// ImageServiceOptions represents any options passable when creating
// the image generation service
type ImageServiceOptions struct {
	Debug             bool
	Host              string
	PostgresAddress   string
	PrometheusAddress string
}

// NewImageService creates the service and initializes it.
func NewImageService(ctx context.Context, options ImageServiceOptions) (is *ImageService, err error) {
	is = &ImageService{
		ctx:     ctx,
		Options: options,
		Client:  http.Client{Timeout: 5 * time.Second},
	}

	return is, nil
}

func (is *ImageService) Open() {
	is.StartTime = time.Now()
	welcomer.Logger.Info().Msgf("Starting image service. Version %s", VERSION)

	// Setup HTTP
	go is.setupHTTP()

	// Setup Prometheus
	go is.setupPrometheus()
}

func (is *ImageService) setupHTTP() error {
	router := gin.New()

	router.Use(logger.SetLogger())
	router.Use(gin.Recovery())

	is.registerRoutes(router)

	welcomer.Logger.Info().Msgf("Serving http at %s", is.Options.Host)

	err := router.Run(is.Options.Host)
	if err != nil {
		welcomer.Logger.Error().Err(err).Str("host", is.Options.Host).Msg("Failed to serve gRPC server")

		return fmt.Errorf("failed to serve grpc: %w", err)
	}

	return nil
}

func (is *ImageService) setupPrometheus() error {
	http.Handle("/metrics", promhttp.HandlerFor(
		prometheus.DefaultGatherer,
		promhttp.HandlerOpts{},
	))

	welcomer.Logger.Info().Msgf("Serving prometheus at %s", is.Options.PrometheusAddress)

	err := http.ListenAndServe(is.Options.PrometheusAddress, nil)
	if err != nil {
		welcomer.Logger.Error().Str("host", is.Options.PrometheusAddress).Err(err).Msg("Failed to serve prometheus server")

		return fmt.Errorf("failed to serve prometheus: %w", err)
	}

	return nil
}

func (is *ImageService) Close() error {
	welcomer.Logger.Info().Msg("Closing image service")

	return nil
}
