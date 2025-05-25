package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	sandwich_daemon "github.com/WelcomerTeam/Sandwich-Daemon"
	jetstream_client "github.com/WelcomerTeam/Welcomer/welcomer-core/jetstream"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	_ "github.com/joho/godotenv/autoload"
)

func panicHandler(_ *sandwich_daemon.Sandwich, err any) {
	slog.Error("Panic in Sandwich Daemon", "error", err)
}

func main() {
	stanAddress := flag.String("stanAddress", os.Getenv("STAN_ADDRESS"), "NATs streaming Address")
	stanChannel := flag.String("stanChannel", os.Getenv("STAN_CHANNEL"), "NATs streaming Channel")

	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stateProvider := sandwich_daemon.NewStateProviderMemoryOptimized()

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

	sandwich := sandwich_daemon.NewSandwich(
		slog.Default(),
		sandwich_daemon.NewConfigProviderFromPath("sandwich.json.local"),
		sandwich_daemon.NewProxyClient(*http.DefaultClient, url.URL{
			Scheme: "https",
			Host:   "discord.com",
		}),
		sandwich_daemon.NewEventProviderWithBlacklist(sandwich_daemon.NewBuiltinDispatchProvider(true)),
		sandwich_daemon.NewIdentifyViaBuckets(),
		producerProvider,
		stateProvider,
	).
		WithPanicHandler(panicHandler).
		WithPrometheusAnalytics(
			&http.Server{
				Addr:              ":10000",
				WriteTimeout:      time.Second * 10,
				ReadTimeout:       time.Second * 10,
				ReadHeaderTimeout: time.Second * 10,
				IdleTimeout:       time.Second * 10,
				ErrorLog:          slog.NewLogLogger(slog.With("service", "prometheus").Handler(), slog.LevelError),
			},
			prometheus.NewPedanticRegistry(),
			promhttp.HandlerOpts{},
		).
		WithGRPCServer(
			nil,
			"tcp",
			":15008",
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
