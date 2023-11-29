package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	core "github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-images/service"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/natefinch/lumberjack.v2"

	_ "github.com/joho/godotenv/autoload"
)

const (
	PermissionsDefault = 0o744
)

func main() {
	prometheusAddress := flag.String("prometheusAddress", os.Getenv("PROMETHEUS_ADDRESS"), "Prometheus address")

	postgresURL := flag.String("postgresURL", os.Getenv("POSTGRES_URL"), "Postgres connection URL")

	grpcNetwork := flag.String("grpcNetwork", os.Getenv("GRPC_NETWORK"), "GRPC network type. The network must be \"tcp\", \"tcp4\", \"tcp6\", \"unix\" or \"unixpacket\".")
	grpcHost := flag.String("grpcHost", os.Getenv("GRPC_HOST"), "Host for GRPC.")
	grpcCertFile := flag.String("grpcCertFile", os.Getenv("GRPC_CERT_FILE"), "Optional cert file to use.")
	grpcServerNameOverride := flag.String("grpcServerNameOverride", os.Getenv("GRPC_SERVER_NAME_OVERRIDE"), "For testing only. If set to a non empty string, it will override the virtual host name of authority (e.g. :authority header field) in requests.")

	loggingLevel := flag.String("level", os.Getenv("LOGGING_LEVEL"), "Logging level")

	loggingFileLoggingEnabled := flag.Bool("fileLoggingEnabled", core.MustParseBool(os.Getenv("LOGGING_FILE_LOGGING_ENABLED")), "When enabled, will save logs to files")
	loggingEncodeAsJSON := flag.Bool("encodeAsJSON", core.MustParseBool(os.Getenv("LOGGING_ENCODE_AS_JSON")), "When enabled, will save logs as JSON")
	loggingCompress := flag.Bool("compress", core.MustParseBool(os.Getenv("LOGGING_COMPRESS")), "If true, will compress log files once reached max size")
	loggingDirectory := flag.String("directory", os.Getenv("LOGGING_DIRECTORY"), "Directory to store logs in")
	loggingFilename := flag.String("filename", os.Getenv("LOGGING_FILENAME"), "Filename to store logs as")
	loggingMaxSize := flag.Int("maxSize", core.MustParseInt(os.Getenv("LOGGING_MAX_SIZE")), "Maximum size for log files before being split into seperate files")
	loggingMaxBackups := flag.Int("maxBackups", core.MustParseInt(os.Getenv("LOGGING_MAX_BACKUPS")), "Maximum number of log files before being deleted")
	loggingMaxAge := flag.Int("maxAge", core.MustParseInt(os.Getenv("LOGGING_MAX_AGE")), "Maximum age in days for a log file")

	flag.Parse()

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

	options := service.ImageServiceOptions{
		PrometheusAddress: *prometheusAddress,
		PostgresAddress:   *postgresURL,

		GRPCNetwork:            *grpcNetwork,
		GRPCHost:               *grpcHost,
		GRPCCertFile:           *grpcCertFile,
		GRPCServerNameOverride: *grpcServerNameOverride,
	}

	// Image Service initialization
	image_service, err := service.NewImageService(writer, options)
	if err != nil {
		logger.Panic().Err(err).Msg("Cannot create image service")
	}

	image_service.Open()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-signalCh

	err = image_service.Close()
	if err != nil {
		logger.Warn().Err(err).Msg("Exception whilst closing image service")
	}
}
