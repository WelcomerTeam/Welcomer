package main

import (
	"context"
	"flag"
	"fmt"
	"net/url"
	"os"
	"time"

	backend "github.com/WelcomerTeam/Welcomer/welcomer-backend/backend"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	utils "github.com/WelcomerTeam/Welcomer/welcomer-utils"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
	_ "github.com/joho/godotenv/autoload"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	_ "github.com/joho/godotenv/autoload"
)

func main() {

	loggingLevel := flag.String("level", os.Getenv("LOGGING_LEVEL"), "Logging level")

	releaseMode := flag.String("ginMode", os.Getenv("GIN_MODE"), "gin mode (release/debug)")

	domain := flag.String("domain", os.Getenv("DOMAIN"), "Domain name for the backend")

	host := flag.String("host", os.Getenv("HOST"), "Host to serve backend from")

	keyPairs := flag.String("keypairs", os.Getenv("KEYPAIRS"), "Comma separated list of keypairs to use for sessions. This should be in the format <newhashkey>,<newblockkey>,<oldhashkey>,<oldblockkey>... to allow for key rotation")

	nginxAddress := flag.String("nginxProxy", os.Getenv("NGINX_PROXY"), "NGINX Proxy Address. Used to set trusted proxies.")

	postgresURL := flag.String("postgresURL", os.Getenv("POSTGRES_URL"), "Postgres connection URL")

	prometheusAddress := flag.String("prometheusAddress", os.Getenv("PROMETHEUS_ADDRESS"), "Prometheus address")

	sandwichGRPCHost := flag.String("sandwichGRPCHost", os.Getenv("SANDWICH_GRPC_HOST"), "GRPC Address for the Sandwich Daemon service")

	proxyAddress := flag.String("proxyAddress", os.Getenv("PROXY_ADDRESS"), "Address to proxy requests through. This can be 'https://discord.com', if one is not setup.")

	proxyDebug := flag.Bool("proxyDebug", false, "Enable debugging requests to the proxy")

	botToken := flag.String("botToken", os.Getenv("BOT_TOKEN"), "Primary bot token")

	fallbackBotToken := flag.String("donatorBotToken", os.Getenv("BOT_TOKEN_DONATOR"), "Donator bot token")

	discordClientID := flag.String("discordClientID", os.Getenv("DISCORD_CLIENT_ID"), "OAuth2 Client ID")

	discordClientSecret := flag.String("discordClientSecret", os.Getenv("DISCORD_CLIENT_SECRET"), "OAuth2 Client Secret")

	discordRedirectURL := flag.String("discordRedirectURL", os.Getenv("DISCORD_REDIRECT_URL"), "OAuth2 Redirect URL")

	patreonClientID := flag.String("patreonClientID", os.Getenv("PATREON_CLIENT_ID"), "OAuth2 Client ID")

	patreonClientSecret := flag.String("patreonClientSecret", os.Getenv("PATREON_CLIENT_SECRET"), "OAuth2 Client Secret")

	patreonRedirectURL := flag.String("patreonRedirectURL", os.Getenv("PATREON_REDIRECT_URL"), "OAuth2 Redirect URL")

	paypalClientID := flag.String("paypalClientID", os.Getenv("PAYPAL_CLIENT_ID"), "Paypal client ID")

	paypalClientSecret := flag.String("paypalSecretID", os.Getenv("PAYPAL_CLIENT_SECRET"), "Paypal client secret")

	paypalIsLive := flag.Bool("paypalIsLive", utils.TryParseBool(os.Getenv("PAYPAL_LIVE")), "Enable live mode for paypal")

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

		Out: os.Stdout,

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

	restInterface := welcomer.NewTwilightProxy(*proxyURL)

	restInterface.SetDebug(*proxyDebug)

	// Setup GRPC

	var grpcConnection *grpc.ClientConn

	if grpcConnection, err = grpc.NewClient(

		*sandwichGRPCHost,

		grpc.WithTransportCredentials(insecure.NewCredentials()),

		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(1024*1024*1024)), // Set max message size to 1GB

	); err != nil {

		panic(fmt.Sprintf(`grpc.NewClient(%s): %v`, *sandwichGRPCHost, err.Error()))

	}

	// Setup postgres pool.

	var pool *pgxpool.Pool

	if pool, err = pgxpool.Connect(ctx, *postgresURL); err != nil {

		panic(fmt.Sprintf(`pgxpool.Connect(%s): %v`, *postgresURL, err.Error()))

	}

	// Setup app.

	app, err := backend.NewBackend(ctx, logger, backend.BackendOptions{

		Domain: *domain,

		Host: *host,

		KeyPairs: *keyPairs,

		NginxAddress: *nginxAddress,

		PostgresAddress: *postgresURL,

		PrometheusAddress: *prometheusAddress,

		Conn: grpcConnection,

		RESTInterface: restInterface,

		Pool: pool,

		BotToken: *botToken,

		DonatorBotToken: *fallbackBotToken,

		DiscordClientID: *discordClientID,

		DiscordClientSecret: *discordClientSecret,

		DiscordRedirectURL: *discordRedirectURL,

		PatreonClientID: *patreonClientID,

		PatreonClientSecret: *patreonClientSecret,

		PatreonRedirectURL: *patreonRedirectURL,

		PaypalClientID: *paypalClientID,

		PaypalClientSecret: *paypalClientSecret,

		PaypalIsLive: *paypalIsLive,
	})

	if err != nil || app == nil {

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
