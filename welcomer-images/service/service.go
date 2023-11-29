package service

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	grpcServer "github.com/WelcomerTeam/Welcomer/welcomer-images/protobuf"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
)

// VERSION follows semantic versioning.
const VERSION = "0.0.1"

//go:generate go run assets_gen.go assets backgrounds
//go:generate go run fonts_gen.go fonts fallback

type ImageService struct {
	ctx    context.Context
	cancel func()

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
	PrometheusAddress string `json:"prometheus_address" yaml:"prometheus_address"`
	PostgresAddress   string `json:"postgres_address" yaml:"postgres_address"`

	GRPCNetwork            string `json:"grpc_network" yaml:"grpc_network"`
	GRPCHost               string `json:"grpc_host" yaml:"grpc_host"`
	GRPCCertFile           string `json:"grpc_cert_file" yaml:"grpc_cert_file"`
	GRPCServerNameOverride string `json:"grpc_server_name_override" yaml:"grpc_server_name_override"`
}

// NewImageService creates the service and initializes it.
func NewImageService(logger io.Writer, options ImageServiceOptions) (is *ImageService, err error) {
	is = &ImageService{
		Logger: zerolog.New(logger).With().Timestamp().Logger(),

		Options: options,

		Client: http.Client{},
	}

	is.ctx, is.cancel = context.WithCancel(context.Background())

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

	// Setup GRPC
	go is.setupGRPC()

	// Setup Prometheus
	go is.setupPrometheus()
}

func (is *ImageService) setupGRPC() error {
	network := is.Options.GRPCNetwork
	host := is.Options.GRPCHost
	certpath := is.Options.GRPCCertFile
	servernameoverride := is.Options.GRPCServerNameOverride

	var grpcOptions []grpc.ServerOption

	if certpath != "" {
		var creds credentials.TransportCredentials

		creds, err := credentials.NewClientTLSFromFile(certpath, servernameoverride)
		if err != nil {
			is.Logger.Error().Err(err).Msg("Failed to create new client TLS from file for gRPC")

			return err
		}

		grpcOptions = append(grpcOptions, grpc.Creds(creds))
	}

	grpcListener := grpc.NewServer(grpcOptions...)
	grpcServer.RegisterImageGenerationServiceServer(grpcListener, is.newImageGenerationServiceServer())
	reflection.Register(grpcListener)

	listener, err := net.Listen(network, host)
	if err != nil {
		is.Logger.Panic().Str("host", host).Err(err).Msg("Failed to bind to host")

		return err
	}

	is.Logger.Info().Msgf("Serving gRPC at %s", host)

	err = grpcListener.Serve(listener)
	if err != nil {
		is.Logger.Error().Str("host", host).Err(err).Msg("Failed to serve gRPC server")

		return fmt.Errorf("failed to serve grpc: %w", err)
	}

	return nil
}

func (is *ImageService) setupPrometheus() error {
	prometheus.MustRegister(grpcImgenRequests)

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

	if is.cancel != nil {
		is.cancel()
	}

	return nil
}
