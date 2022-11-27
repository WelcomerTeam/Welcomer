package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	"github.com/WelcomerTeam/Discord/discord"
	messaging "github.com/WelcomerTeam/Sandwich/messaging"
	sandwich "github.com/WelcomerTeam/Sandwich/sandwich"
	core "github.com/WelcomerTeam/Welcomer/welcomer-core"
	gateway "github.com/WelcomerTeam/Welcomer/welcomer-gateway"
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
	identifierName := flag.String("identifierName", os.Getenv("IDENTIFIER_NAME"), "Sandwich identifier name")

	grpcAddress := flag.String("grpcAddress", os.Getenv("GRPC_ADDRESS"), "GRPC Address")
	proxyAddress := flag.String("proxyAddress", os.Getenv("PROXY_ADDRESS"), "Twilight proxy Address")

	stanAddress := flag.String("stanAddress", os.Getenv("STAN_ADDRESS"), "NATs Streaming Address")
	stanCluster := flag.String("stanCluster", os.Getenv("STAN_CLUSTER"), "NATs Streaming Cluster")
	stanChannel := flag.String("stanChannel", os.Getenv("STAN_CHANNEL"), "NATs Streaming Channel")

	loggingLevel := flag.String("level", os.Getenv("LOGGING_LEVEL"), "Logging level")

	loggingFileLoggingEnabled := flag.Bool("fileLoggingEnabled", core.MustParseBool(os.Getenv("LOGGING_FILE_LOGGING_ENABLED")), "When enabled, will save logs to files")
	loggingEncodeAsJSON := flag.Bool("encodeAsJSON", core.MustParseBool(os.Getenv("LOGGING_ENCODE_AS_JSON")), "When enabled, will save logs as JSON")
	loggingCompress := flag.Bool("compress", core.MustParseBool(os.Getenv("LOGGING_COMPRESS")), "If true, will compress log files once reached max size")
	loggingDirectory := flag.String("directory", os.Getenv("LOGGING_DIRECTORY"), "Directory to store logs in")
	loggingFilename := flag.String("filename", os.Getenv("LOGGING_FILENAME"), "Filename to store logs as")
	loggingMaxSize := flag.Int("maxSize", core.MustParseInt(os.Getenv("LOGGING_MAX_SIZE")), "Maximum size for log files before being split into seperate files")
	loggingMaxBackups := flag.Int("maxBackups", core.MustParseInt(os.Getenv("LOGGING_MAX_BACKUPS")), "Maximum number of log files before being deleted")
	loggingMaxAge := flag.Int("maxAge", core.MustParseInt(os.Getenv("LOGGING_MAX_AGE")), "Maximum age in days for a log file")

	oneShot := flag.Bool("oneshot", false, "If true, will close the app after setting up the app")

	proxyDebug := flag.Bool("proxyDebug", false, "Enable debug on proxy")

	flag.Parse()

	// Setup Rest
	proxyURL, err := url.Parse(*proxyAddress)
	if err != nil {
		panic(fmt.Errorf("failed to parse proxy address. url.Parse(%s): %w", *proxyAddress, err))
	}

	restInterface := discord.NewTwilightProxy(*proxyURL)
	restInterface.SetDebug(*proxyDebug)

	// Setup GRPC
	grpcConnection, err := grpc.Dial(*grpcAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(fmt.Sprintf(`grpc.Dial(%s): %v`, *grpcAddress, err.Error()))
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

	context, cancel := context.WithCancel(context.Background())

	// Setup NATs
	stanClient := messaging.NewStanMQClient()

	err = stanClient.Connect(context, "sandwich", map[string]interface{}{
		"Address": *stanAddress,
		"Cluster": *stanCluster,
		"Channel": *stanChannel,
	})
	if err != nil {
		panic(fmt.Sprintf(`stanClient.Connect(): %v`, err.Error()))
	}

	err = stanClient.Subscribe(context, *stanChannel)
	if err != nil {
		panic(fmt.Sprintf(`stanClient.Subscribe(%s): %v`, *stanChannel, err.Error()))
	}

	// Setup sandwich.
	sandwichClient := sandwich.NewSandwich(grpcConnection, restInterface, writer)

	welcomer := gateway.NewWelcomer(*identifierName, sandwichClient)
	sandwichClient.RegisterBot(*identifierName, welcomer.Bot)

	// We return if it a dry run. Any issues loading up the bot would've already caused a panic.
	if *oneShot {
		return
	}

	// Signal
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	// Register message channels
	stanMessages := stanClient.Chan()

	err = sandwichClient.ListenToChannel(context, stanMessages)
	if err != nil {
		panic(fmt.Sprintf(`sandwichClient.ListenToChannel(): %w`, err))
	}

	cancel()

	// Close sandwich
	stanClient.Unsubscribe()

	err = grpcConnection.Close()
	if err != nil {
		logger.Warn().Err(err).Msg("Exception whilst closing grpc client")
	}
}
