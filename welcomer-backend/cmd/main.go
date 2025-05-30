package main

import (
	"context"
	"flag"
	"os"

	backend "github.com/WelcomerTeam/Welcomer/welcomer-backend/backend"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
	paypalIsLive := flag.Bool("paypalIsLive", welcomer.TryParseBool(os.Getenv("PAYPAL_LIVE")), "Enable live mode for paypal")

	sandwichManagerName := flag.String("sandwichManagerName", os.Getenv("SANDWICH_MANAGER_NAME"), "Sandwich manager identifier name")

	flag.Parse()

	var err error

	ctx, cancel := context.WithCancel(context.Background())

	restInterface := welcomer.NewTwilightProxy(*proxyAddress)
	restInterface.SetDebug(*proxyDebug)

	welcomer.SetupDefaultManagerName(*sandwichManagerName)
	welcomer.SetupLogger(*loggingLevel)
	welcomer.SetupGRPCConnection(*sandwichGRPCHost,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(1024*1024*1024)), // Set max message size to 1GB
	)
	welcomer.SetupRESTInterface(restInterface)
	welcomer.SetupSandwichClient()
	welcomer.SetupDatabase(ctx, *postgresURL)

	gin.SetMode(*releaseMode)

	app, err := backend.NewBackend(backend.Options{
		Domain:            *domain,
		Host:              *host,
		KeyPairs:          *keyPairs,
		NginxAddress:      *nginxAddress,
		PostgresAddress:   *postgresURL,
		PrometheusAddress: *prometheusAddress,

		BotToken:        *botToken,
		DonatorBotToken: *fallbackBotToken,

		DiscordClientID:     *discordClientID,
		DiscordClientSecret: *discordClientSecret,
		DiscordRedirectURL:  *discordRedirectURL,

		PatreonClientID:     *patreonClientID,
		PatreonClientSecret: *patreonClientSecret,
		PatreonRedirectURL:  *patreonRedirectURL,

		PaypalClientID:     *paypalClientID,
		PaypalClientSecret: *paypalClientSecret,
		PaypalIsLive:       *paypalIsLive,
	})

	if err != nil || app == nil {
		welcomer.Logger.Panic().Err(err).Msg("Exception creating app")
	}

	if err = app.Open(); err != nil {
		welcomer.Logger.Warn().Err(err).Msg("Exceptions whilst starting app")
	}

	cancel()

	// Close app

	if err = app.Close(); err != nil {
		welcomer.Logger.Warn().Err(err).Msg("Exception whilst closing app")
	}

	if err = welcomer.GRPCConnection.Close(); err != nil {
		welcomer.Logger.Warn().Err(err).Msg("Exception whilst closing grpc client")
	}
}
