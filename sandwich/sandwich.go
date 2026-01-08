package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"
	"syscall"
	"time"

	_ "net/http/pprof"

	sandwich_daemon "github.com/WelcomerTeam/Sandwich-Daemon"
	welcomer "github.com/WelcomerTeam/Welcomer/welcomer-core"
	jetstream_client "github.com/WelcomerTeam/Welcomer/welcomer-core/jetstream"
	"github.com/go-redis/redis/v8"
	_ "github.com/joho/godotenv/autoload"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
)

func panicHandler(_ *sandwich_daemon.Sandwich, err any) {
	slog.Error("Panic in Sandwich Daemon", "error", err)
	print(string(debug.Stack()))
}

func main() {
	loggingLevel := flag.String("level", os.Getenv("LOGGING_LEVEL"), "Logging level")
	postgresURL := flag.String("postgresURL", os.Getenv("POSTGRES_URL"), "Postgres connection URL")
	stanAddress := flag.String("stanAddress", os.Getenv("STAN_ADDRESS"), "NATs streaming Address")
	stanChannel := flag.String("stanChannel", os.Getenv("STAN_CHANNEL"), "NATs streaming Channel")
	prometheusHost := flag.String("prometheusHost", os.Getenv("PROMETHEUS_HOST"), "Prometheus host")
	redisHost := flag.String("redisHost", os.Getenv("REDIS_HOST"), "Redis host")
	grpcHost := flag.String("grpcHost", os.Getenv("GRPC_HOST"), "GRPC host")
	proxyHost := flag.String("proxyHost", os.Getenv("PROXY_HOST"), "Proxy host")

	enablePprof := flag.Bool("enablePprof", welcomer.TryParseBool(os.Getenv("ENABLE_PPROF")), "Enable pprof debugging server")

	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	welcomer.SetupLogger(*loggingLevel)
	welcomer.SetupDatabase(ctx, *postgresURL)

	redisClient := redis.NewClient(&redis.Options{
		Addr: *redisHost,
	})

	stateProvider := sandwich_daemon.NewStateProviderMemoryOptimized()
	dedupeProvider := welcomer.NewRedisDedupeProvider(redisClient, slog.Default())

	producerProvider, err := jetstream_client.NewJetstreamProducerProvider(
		ctx,
		*stanAddress,
		*stanChannel,
		nil,
		nil,
	)
	if err != nil {
		panic(fmt.Errorf("failed to create producer provider: %w", err))
	}

	proxyURL, err := url.Parse(*proxyHost)
	if err != nil {
		panic(fmt.Errorf("url.Parse(%s): %w", *proxyHost, err))
	}

	slog.SetLogLoggerLevel(slog.LevelDebug)
	logger := slog.Default()

	if *enablePprof {
		go func() {
			slog.Info("Starting pprof server on :6060", "service", "pprof")

			err := http.ListenAndServe(":6060", nil)
			if err != nil {
				slog.Error("Failed to start pprof server", "error", err)
			}
		}()
	}

	registry := prometheus.NewPedanticRegistry()
	registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	registry.MustRegister(collectors.NewGoCollector())

	sandwich := sandwich_daemon.NewSandwich(
		logger,
		welcomer.GetConfigurationGatherer(ctx),
		NewProxyClient(*http.DefaultClient, *proxyURL),
		sandwich_daemon.NewEventProviderWithBlacklist(sandwich_daemon.NewBuiltinDispatchProvider(true)),
		sandwich_daemon.NewIdentifyViaBuckets(),
		producerProvider,
		stateProvider,
		dedupeProvider,
	).
		WithPanicHandler(panicHandler).
		WithPrometheusAnalytics(
			&http.Server{
				Addr:              *prometheusHost,
				WriteTimeout:      time.Second * 10,
				ReadTimeout:       time.Second * 10,
				ReadHeaderTimeout: time.Second * 10,
				IdleTimeout:       time.Second * 10,
				ErrorLog:          slog.NewLogLogger(slog.With("service", "prometheus").Handler(), slog.LevelError),
			},
			registry,
			promhttp.HandlerOpts{},
		).
		WithGRPCServer(
			nil,
			"tcp",
			*grpcHost,
			grpc.NewServer(),
		)

	err = sandwich.Start(ctx)
	if err != nil {
		slog.Error("Failed to start Sandwich Daemon", "error", err)

		return
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	sandwich.Stop(ctx)
}

// Verbose Proxy Client

var UserAgent = fmt.Sprintf("Sandwich/%s (https://github.com/WelcomerTeam/Sandwich-Daemon)", sandwich_daemon.Version)

// NewProxyClient creates an HTTP client that redirects all requests through a specified host.
// This is useful when using a proxy such as twilight or nirn.
func NewProxyClient(client http.Client, host url.URL) *http.Client {
	if client.Transport == nil {
		client.Transport = http.DefaultTransport
	}

	client.Transport = &proxyTransport{
		host:      host,
		transport: client.Transport,
	}

	return &client
}

type proxyTransport struct {
	host      url.URL
	transport http.RoundTripper
}

func (t *proxyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Create a copy of the request to modify
	proxyReq := req.Clone(req.Context())

	// Set the new host while keeping the original path and query
	proxyReq.URL.Host = t.host.Host
	proxyReq.URL.Scheme = t.host.Scheme
	proxyReq.Host = t.host.Host

	if !strings.HasPrefix(proxyReq.URL.Path, "/api") {
		proxyReq.URL.Path = "/api/v10" + proxyReq.URL.Path
	}

	proxyReq.Header.Set("User-Agent", UserAgent)

	// Perform the request using the underlying transport
	resp, err := t.transport.RoundTrip(proxyReq)
	if err != nil {
		return nil, fmt.Errorf("failed to round trip: %w", err)
	}

	return resp, nil
}
