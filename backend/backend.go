package backend

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"sync"
	"time"

	discord "github.com/WelcomerTeam/Discord/discord"
	protobuf "github.com/WelcomerTeam/Sandwich-Daemon/protobuf"
	sandwich "github.com/WelcomerTeam/Sandwich/sandwich"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
	yaml "gopkg.in/yaml.v3"
)

const VERSION = "0.1"

const (
	PermissionsDefault = 0o744
	PermissionWrite    = 0o600
)

type Backend struct {
	sync.Mutex

	ConfigurationLocation string `json:"configuration_location"`

	ctx    context.Context
	cancel func()

	Logger    zerolog.Logger `json:"-"`
	StartTime time.Time      `json:"start_time" yaml:"start_time"`

	configurationMu sync.RWMutex
	Configuration   *BackendConfiguration `json:"configuration" yaml:"configuration"`

	RESTInterface discord.RESTInterface

	SandwichClient protobuf.SandwichClient
	GRPCInterface  sandwich.GRPC

	Route *gin.Engine

	EmptySession    *discord.Session
	BotSession      *discord.Session
	FallbackSession *discord.Session

	// Environment Variables.
	host              string
	botToken          string
	fallbackBotToken  string
	prometheusAddress string
	postgresAddress   string
	nginxAddress      string
}

// BackendConfiguration represents the configuration file.
type BackendConfiguration struct {
	Logging struct {
		Level              string `json:"level" yaml:"level"`
		FileLoggingEnabled bool   `json:"file_logging_enabled" yaml:"file_logging_enabled"`

		EncodeAsJSON bool `json:"encode_as_json" yaml:"encode_as_json"`

		Directory  string `json:"directory" yaml:"directory"`
		Filename   string `json:"filename" yaml:"filename"`
		MaxSize    int    `json:"max_size" yaml:"max_size"`
		MaxBackups int    `json:"max_backups" yaml:"max_backups"`
		MaxAge     int    `json:"max_age" yaml:"max_age"`
		Compress   bool   `json:"compress" yaml:"compress"`
	} `json:"logging" yaml:"logging"`

	Webhooks []string `json:"webhooks" yaml:"webhooks"`
}

// NewBackend creates a new backend.
func NewBackend(conn grpc.ClientConnInterface, restInterface discord.RESTInterface, logger io.Writer, isReleaseMode bool, configurationLocation, host, botToken, fallbackBotToken, prometheusAddress, postgresAddress, nginxAddress string) (b *Backend, err error) {
	b = &Backend{
		Logger: zerolog.New(logger).With().Timestamp().Logger(),

		ConfigurationLocation: configurationLocation,

		configurationMu: sync.RWMutex{},
		Configuration:   &BackendConfiguration{},

		RESTInterface: restInterface,

		SandwichClient: protobuf.NewSandwichClient(conn),
		GRPCInterface:  sandwich.NewDefaultGRPCClient(),
	}

	b.ctx, b.cancel = context.WithCancel(context.Background())

	if isReleaseMode {
		gin.SetMode(gin.ReleaseMode)
	}

	b.host = host
	b.botToken = botToken
	b.fallbackBotToken = fallbackBotToken
	b.prometheusAddress = prometheusAddress
	b.postgresAddress = postgresAddress
	b.nginxAddress = nginxAddress

	// Setup sessions
	b.EmptySession = discord.NewSession(b.ctx, "", b.RESTInterface, b.Logger)
	b.BotSession = discord.NewSession(b.ctx, b.botToken, b.RESTInterface, b.Logger)
	b.FallbackSession = discord.NewSession(b.ctx, b.fallbackBotToken, b.RESTInterface, b.Logger)

	b.Route = b.PrepareGin()

	if nginxAddress != "" {
		err = b.Route.SetTrustedProxies([]string{nginxAddress})
		if err != nil {
			return nil, fmt.Errorf("Failed to set trusted proxies: %w", err)
		}
	}

	b.Lock()
	defer b.Unlock()

	configuration, err := b.LoadConfiguration(b.ConfigurationLocation)
	if err != nil {
		return nil, err
	}

	b.Configuration = configuration

	var writers []io.Writer

	writers = append(writers, logger)

	if b.Configuration.Logging.FileLoggingEnabled {
		if err := os.MkdirAll(b.Configuration.Logging.Directory, PermissionsDefault); err != nil {
			log.Error().Err(err).Str("path", b.Configuration.Logging.Directory).Msg("Unable to create log directory")
		} else {
			lumber := &lumberjack.Logger{
				Filename:   path.Join(b.Configuration.Logging.Directory, b.Configuration.Logging.Filename),
				MaxBackups: b.Configuration.Logging.MaxBackups,
				MaxSize:    b.Configuration.Logging.MaxSize,
				MaxAge:     b.Configuration.Logging.MaxAge,
				Compress:   b.Configuration.Logging.Compress,
			}

			if b.Configuration.Logging.EncodeAsJSON {
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
	b.Logger = zerolog.New(mw).With().Timestamp().Logger()
	b.Logger.Info().Msg("Logging configured")

	return b, nil
}

// LoadConfiguration handles loading the configuration file.
func (b *Backend) LoadConfiguration(path string) (configuration *BackendConfiguration, err error) {
	b.Logger.Debug().
		Str("path", path).
		Msg("Loading configuration")

	defer func() {
		if err == nil {
			b.Logger.Info().Msg("Configuration loaded")
		}
	}()

	file, err := ioutil.ReadFile(path)
	if err != nil {
		return configuration, ErrReadConfigurationFailure
	}

	configuration = &BackendConfiguration{}

	err = yaml.Unmarshal(file, configuration)
	if err != nil {
		return configuration, ErrLoadConfigurationFailure
	}

	return configuration, nil
}

// Open sets up any services and starts the webserver.
func (b *Backend) Open() (err error) {
	b.StartTime = time.Now().UTC()
	b.Logger.Info().Msgf("Starting sandwich. Version %s", VERSION)

	go b.PublishSimpleWebhook(b.EmptySession, "Starting backend", "", "Version "+VERSION, EmbedColourSandwich)

	// Setup Prometheus
	go b.SetupPrometheus()

	b.Logger.Info().Msgf("Serving http at %s", b.host)

	err = b.Route.Run(b.host)
	if err != nil {
		return err
	}

	return nil
}

// Close gracefully closes the backend.
func (b *Backend) Close() (err error) {
	return
}

// SetupPrometheus sets up prometheus.
func (b *Backend) SetupPrometheus() (err error) {
	b.Logger.Info().Msgf("Serving prometheus at %s", b.prometheusAddress)

	err = http.ListenAndServe(b.prometheusAddress, nil)
	if err != nil {
		b.Logger.Error().Str("host", b.prometheusAddress).Err(err).Msg("Failed to serve prometheus server")

		return fmt.Errorf("Failed to serve prometheus: %w", err)
	}

	return nil
}

// PrepareGin prepares gin routes and middleware.
func (b *Backend) PrepareGin() (g *gin.Engine) {
	g = gin.Default()

	registerStaticRoutes(g)

	return
}
