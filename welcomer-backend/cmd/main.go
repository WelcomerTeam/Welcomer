package main

import (
	"context"
	"flag"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/WelcomerTeam/Discord/discord"
	backend "github.com/WelcomerTeam/Welcomer/welcomer-backend/backend"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	loggingLevel := flag.String("level", os.Getenv("LOGGING_LEVEL"), "Logging level")

	sandwichGRPCHost := flag.String("sandwichGRPCHost", os.Getenv("SANDWICH_GRPC_HOST"), "GRPC Address for the Sandwich Daemon service")
	proxyAddress := flag.String("proxyAddress", os.Getenv("PROXY_ADDRESS"), "Address to proxy requests through. This can be 'https://discord.com', if one is not setup.")
	proxyDebug := flag.Bool("proxyDebug", false, "Enable debugging requests to the proxy")
	prometheusAddress := flag.String("prometheusAddress", os.Getenv("BACKEND_PROMETHEUS_ADDRESS"), "Prometheus address")
	postgresURL := flag.String("postgresURL", os.Getenv("POSTGRES_URL"), "Postgres connection URL")

	botToken := flag.String("botToken", os.Getenv("BOT_TOKEN"), "Primary bot token")
	fallbackBotToken := flag.String("donatorBotToken", os.Getenv("BOT_TOKEN_DONATOR"), "Donator bot token")
	host := flag.String("host", os.Getenv("BACKEND_HOST"), "Host to serve backend from")

	clientID := flag.String("clientID", os.Getenv("BOT_CLIENT_ID"), "OAuth2 Client ID")
	clientSecret := flag.String("clientSecret", os.Getenv("BOT_CLIENT_SECRET"), "OAuth2 Client Secret")
	redirectURL := flag.String("redirectURL", os.Getenv("BACKEND_REDIRECT_URL"), "OAuth2 Redirect URL")

	nginxAddress := flag.String("nginxProxy", os.Getenv("NGINX_PROXY"), "NGINX Proxy Address. Used to set trusted proxies.")
	releaseMode := flag.String("ginMode", os.Getenv("GIN_MODE"), "gin mode (release/debug)")

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

	// Setup Rest
	var proxyURL *url.URL
	if proxyURL, err = url.Parse(*proxyAddress); err != nil {
		panic(fmt.Sprintf("url.Parse(%s): %v", *proxyAddress, err.Error()))
	}

	restInterface := discord.NewTwilightProxy(*proxyURL)
	restInterface.SetDebug(*proxyDebug)

	// Setup GRPC
	var grpcConnection *grpc.ClientConn
	if grpcConnection, err = grpc.Dial(*sandwichGRPCHost, grpc.WithTransportCredentials(insecure.NewCredentials())); err != nil {
		panic(fmt.Sprintf(`grpc.Dial(%s): %v`, *sandwichGRPCHost, err.Error()))
	}

	// Setup postgres pool.
	var pool *pgxpool.Pool
	if pool, err = pgxpool.Connect(ctx, *postgresURL); err != nil {
		panic(fmt.Sprintf(`pgxpool.Connect(%s): %v`, *postgresURL, err.Error()))
	}

	// Setup app.
	var app *backend.Backend
	if app, err = backend.NewBackend(ctx, grpcConnection, restInterface, writer, pool, *host, *botToken, *fallbackBotToken, *prometheusAddress, *postgresURL, *nginxAddress, *clientID, *clientSecret, *redirectURL); err != nil {
		logger.Panic().Err(err).Msg("Exception creating app")
	}

	if err = app.Open(); err != nil {
		logger.Warn().Err(err).Msg("Exceptions whilst starting app")
	}

	cancel()

	// Close app
	if err = app.Close(); err != nil {
		logger.Warn().Err(err).Msg("Exception whilst closing app")
	}

	if err = grpcConnection.Close(); err != nil {
		logger.Warn().Err(err).Msg("Exception whilst closing grpc client")
	}
}
