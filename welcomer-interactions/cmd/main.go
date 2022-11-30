package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"time"

	"github.com/WelcomerTeam/Discord/discord"
	protobuf "github.com/WelcomerTeam/Sandwich-Daemon/protobuf"
	sandwich "github.com/WelcomerTeam/Sandwich/sandwich"
	subway "github.com/WelcomerTeam/Subway/subway"
	core "github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-interactions"
	"github.com/WelcomerTeam/Welcomer/welcomer-interactions/internal"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gopkg.in/natefinch/lumberjack.v2"

	_ "github.com/joho/godotenv/autoload"
)

const (
	PermissionsDefault = 0o744
)

func main() {
	managerName := flag.String("managerName", os.Getenv("MANAGER_NAME"), "Sandwich manager identifier name")

	grpcAddress := flag.String("grpcAddress", os.Getenv("GRPC_ADDRESS"), "GRPC Address")
	proxyAddress := flag.String("proxyAddress", os.Getenv("PROXY_ADDRESS"), "Twilight proxy Address")

	loggingLevel := flag.String("level", os.Getenv("LOGGING_LEVEL"), "Logging level")

	loggingFileLoggingEnabled := flag.Bool("fileLoggingEnabled", core.MustParseBool(os.Getenv("LOGGING_FILE_LOGGING_ENABLED")), "When enabled, will save logs to files")
	loggingEncodeAsJSON := flag.Bool("encodeAsJSON", core.MustParseBool(os.Getenv("LOGGING_ENCODE_AS_JSON")), "When enabled, will save logs as JSON")
	loggingCompress := flag.Bool("compress", core.MustParseBool(os.Getenv("LOGGING_COMPRESS")), "If true, will compress log files once reached max size")
	loggingDirectory := flag.String("directory", os.Getenv("LOGGING_DIRECTORY"), "Directory to store logs in")
	loggingFilename := flag.String("filename", os.Getenv("LOGGING_FILENAME"), "Filename to store logs as")
	loggingMaxSize := flag.Int("maxSize", core.MustParseInt(os.Getenv("LOGGING_MAX_SIZE")), "Maximum size for log files before being split into seperate files")
	loggingMaxBackups := flag.Int("maxBackups", core.MustParseInt(os.Getenv("LOGGING_MAX_BACKUPS")), "Maximum number of log files before being deleted")
	loggingMaxAge := flag.Int("maxAge", core.MustParseInt(os.Getenv("LOGGING_MAX_AGE")), "Maximum age in days for a log file")

	host := flag.String("host", os.Getenv("HOST"), "Host")
	prometheusAddress := flag.String("prometheusAddress", os.Getenv("PROMETHEUS_ADDRESS"), "Prometheus address")
	publicKey := flag.String("publicKey", os.Getenv("PUBLIC_KEY"), "Public key for signature validation")
	webhookURL := flag.String("webhookURL", os.Getenv("WEBHOOK"), "Webhook to send status messages to")

	oneShot := flag.Bool("oneshot", false, "If true, will close the app after setting up the app")
	syncCommands := flag.Bool("syncCommands", false, "If true, will bulk update commands")

	proxyDebug := flag.Bool("proxyDebug", false, "Enable debug on proxy")

	postgresAddress := flag.String("postgresAddress", os.Getenv("POSTGRES_ADDRESS"), "Postgres connection URL")

	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())

	// Setup Rest
	proxyURL, err := url.Parse(*proxyAddress)
	if err != nil {
		panic(fmt.Errorf("failed to parse proxy address. url.Parse(%s): %w", *proxyAddress, err))
	}

	restInterface := discord.NewTwilightProxy(*proxyURL)

	restInterface = discord.NewInterface(&http.Client{
		Timeout: 20 * time.Second,
	}, discord.EndpointDiscord, "v10", discord.UserAgent)
	restInterface.SetDebug(*proxyDebug)

	// Setup GRPC
	grpcConnection, err := grpc.Dial(*grpcAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(fmt.Sprintf(`grpc.Dial(%s): %w`, *grpcAddress, err.Error()))
	}

	// Setup Logger
	level, err := zerolog.ParseLevel(*loggingLevel)
	if err != nil {
		panic(fmt.Errorf(`failed to parse loggingLevel. zerolog.ParseLevel(%s): %w`, *loggingLevel, err))
	}

	zerolog.SetGlobalLevel(level)

	writer := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.Stamp,
	}

	var writers []io.Writer

	writers = append(writers, writer)

	if *loggingFileLoggingEnabled {
		if err := os.MkdirAll(*loggingDirectory, PermissionsDefault); err != nil {
			log.Error().Err(err).Str("path", *loggingDirectory).Msg("Unable to create log directory")
		} else {
			lumber := &lumberjack.Logger{
				Filename:   path.Join(*loggingDirectory, *loggingFilename),
				MaxBackups: *loggingMaxBackups,
				MaxSize:    *loggingMaxSize,
				MaxAge:     *loggingMaxAge,
				Compress:   *loggingCompress,
			}

			if *loggingEncodeAsJSON {
				writers = append(writers, lumber)
			} else {
				writers = append(writers, zerolog.ConsoleWriter{
					Out:        lumber,
					TimeFormat: time.Stamp,
					NoColor:    true,
				})
			}
		}
	}

	mw := io.MultiWriter(writers...)
	logger := zerolog.New(mw).With().Timestamp().Logger()
	logger.Info().Msg("Logging configured")

	var webhook []string
	if webhookURL != nil && *webhookURL != "" {
		webhook = []string{*webhookURL}
	}

	sandwichClient := protobuf.NewSandwichClient(grpcConnection)

	// Setup postgres pool.
	pool, err := pgxpool.Connect(ctx, *postgresAddress)
	if err != nil {
		panic(fmt.Sprintf("pgxpool.Connect(%s): %w", *postgresAddress, err))
	}

	ctx = internal.AddPoolToContext(ctx, pool)
	ctx = internal.AddManagerNameToContext(ctx, *managerName)

	// Setup app.
	app := welcomer.NewWelcomer(ctx, subway.SubwayOptions{
		SandwichClient:    sandwichClient,
		RESTInterface:     restInterface,
		Logger:            logger,
		PublicKey:         *publicKey,
		PrometheusAddress: *prometheusAddress,
		Webhooks:          webhook,
	})
	if err != nil {
		logger.Panic().Err(err).Msg("Exception creating app")
	}

	if *syncCommands {
		grpcInterface := sandwich.NewDefaultGRPCClient()
		configurations, _ := grpcInterface.FetchConsumerConfiguration(&sandwich.GRPCContext{
			Context:        ctx,
			SandwichClient: sandwichClient,
		}, *managerName)

		configuration, ok := configurations.Identifiers[*managerName]
		if !ok {
			panic(fmt.Errorf(`failed to sync command: could not find manager matching "%s"`, *managerName))
		}

		err = app.SyncCommands(ctx, "Bot "+configuration.Token, configuration.ID)
		if err != nil {
			panic(fmt.Errorf(`failed to sync commands. app.SyncCommands(): %w`, err))
		}

		logger.Info().Msg("Synced commands")
	}

	if !*oneShot {
		err = app.ListenAndServe("", *host)
		if err != nil {
			logger.Warn().Err(err).Msg("Exceptions whilst starting app")
		}
	}

	cancel()

	err = grpcConnection.Close()
	if err != nil {
		logger.Warn().Err(err).Msg("Exception whilst closing grpc client")
	}
}
