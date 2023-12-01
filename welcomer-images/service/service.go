package service

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	"github.com/gin-contrib/logger"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
)

// VERSION follows semantic versioning.
const VERSION = "0.0.1"

//go:generate go run assets_gen.go assets backgrounds
//go:generate go run fonts_gen.go fonts fallback

type ImageService struct {
	ctx context.Context

	Logger    zerolog.Logger `json:"-"`
	StartTime time.Time      `json:"start_time" yaml:"start_time"`

	Options ImageServiceOptions `json:"options" yaml:"options"`

	Pool *pgxpool.Pool

	Database *database.Queries

	Client http.Client

	Fonts map[string]*Font
}

// ImageServiceOptions represents any options passable when creating
// the image generation service
type ImageServiceOptions struct {
	PrometheusAddress string
	PostgresAddress   string
	Host              string
	Debug             bool
}

// NewImageService creates the service and initializes it.
func NewImageService(ctx context.Context, logger io.Writer, options ImageServiceOptions) (is *ImageService, err error) {
	is = &ImageService{
		ctx: ctx,

		Logger: zerolog.New(logger).With().Timestamp().Logger(),

		Options: options,

		Client: http.Client{Timeout: 5 * time.Second},
	}

	// Setup postgres pool.
	pool, err := pgxpool.Connect(is.ctx, is.Options.PostgresAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	is.Pool = pool

	is.Database = database.New(is.Pool)

	return is, nil
}

func (is *ImageService) Open() {
	is.StartTime = time.Now().UTC()
	is.Logger.Info().Msgf("Starting image service. Version %s", VERSION)

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

	is.Logger.Info().Msgf("Serving http at %s", is.Options.Host)

	err := router.Run(is.Options.Host)
	if err != nil {
		is.Logger.Error().Err(err).Str("host", is.Options.Host).Msg("Failed to serve gRPC server")

		return fmt.Errorf("failed to serve grpc: %w", err)
	}

	return nil
}

func (is *ImageService) setupPrometheus() error {
	prometheus.MustRegister(imgenRequests)
	prometheus.MustRegister(imgenTotalRequests)
	prometheus.MustRegister(imgenTotalDuration)
	prometheus.MustRegister(imgenDuration)

	http.Handle("/metrics", promhttp.HandlerFor(
		prometheus.DefaultGatherer,
		promhttp.HandlerOpts{},
	))

	is.Logger.Info().Msgf("Serving prometheus at %s", is.Options.PrometheusAddress)

	err := http.ListenAndServe(is.Options.PrometheusAddress, nil)
	if err != nil {
		is.Logger.Error().Str("host", is.Options.PrometheusAddress).Err(err).Msg("Failed to serve prometheus server")

		return fmt.Errorf("failed to serve prometheus: %w", err)
	}

	return nil
}

func (is *ImageService) Close() error {
	is.Logger.Info().Msg("Closing image service")

	return nil
}
