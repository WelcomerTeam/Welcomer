package backend

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"sync"
	"time"

	discord "github.com/WelcomerTeam/Discord/discord"
	protobuf "github.com/WelcomerTeam/Sandwich-Daemon/protobuf"
	sandwich "github.com/WelcomerTeam/Sandwich/sandwich"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	"github.com/gin-contrib/sessions"
	limits "github.com/gin-contrib/size"
	"github.com/jackc/pgx/v4/pgxpool"
	ginprometheus "github.com/zsais/go-gin-prometheus"

	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/logger"
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

	RequestSizeLimit = 100000000
)

var backend *Backend

type Backend struct {
	sync.Mutex

	ConfigurationLocation string `json:"configuration_location"`

	ctx    context.Context
	cancel func()

	Logger    zerolog.Logger `json:"-"`
	StartTime time.Time      `json:"start_time" yaml:"start_time"`

	configurationMu sync.RWMutex
	Configuration   *Configuration `json:"configuration" yaml:"configuration"`

	RESTInterface discord.RESTInterface

	SandwichClient protobuf.SandwichClient
	GRPCInterface  sandwich.GRPC

	PrometheusHandler *ginprometheus.Prometheus

	Route *gin.Engine

	Pool  *pgxpool.Pool
	Store Store

	Database *database.Queries

	EmptySession *discord.Session

	botToken   string
	BotSession *discord.Session

	donatorBotToken   string
	DonatorBotSession *discord.Session

	cdnCustomBackgroundsPath string
	customBackgroundsPath    string

	// Environment Variables.
	host              string
	prometheusAddress string
	postgresAddress   string
	nginxAddress      string
}

// Configuration represents the configuration file.
type Configuration struct {
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
func NewBackend(conn grpc.ClientConnInterface, restInterface discord.RESTInterface, logger io.Writer, isReleaseMode bool, configurationLocation, host, botToken, donatorBotToken, prometheusAddress, postgresAddress, nginxAddress, clientId, clientSecret, redirectURL, cdnCustomBackgroundsPath, cdnBackgroundsPath string) (b *Backend, err error) {
	if backend != nil {
		return backend, ErrBackendAlreadyExists
	}

	b = &Backend{
		Logger: zerolog.New(logger).With().Timestamp().Logger(),

		ConfigurationLocation: configurationLocation,

		configurationMu: sync.RWMutex{},
		Configuration:   &Configuration{},

		RESTInterface: restInterface,

		SandwichClient: protobuf.NewSandwichClient(conn),
		GRPCInterface:  sandwich.NewDefaultGRPCClient(),

		PrometheusHandler: ginprometheus.NewPrometheus("gin"),

		cdnCustomBackgroundsPath: cdnCustomBackgroundsPath,
		customBackgroundsPath:    cdnBackgroundsPath,

		host:              host,
		botToken:          botToken,
		donatorBotToken:   donatorBotToken,
		prometheusAddress: prometheusAddress,
		postgresAddress:   postgresAddress,
		nginxAddress:      nginxAddress,
	}

	b.Lock()
	defer b.Unlock()

	b.ctx, b.cancel = context.WithCancel(context.Background())

	if isReleaseMode {
		gin.SetMode(gin.ReleaseMode)
	}

	// Setup OAuth2

	OAuth2Config.ClientID = clientId
	OAuth2Config.ClientSecret = clientSecret
	OAuth2Config.RedirectURL = redirectURL

	// Setup sessions
	b.EmptySession = discord.NewSession(b.ctx, "", b.RESTInterface, b.Logger)
	b.BotSession = discord.NewSession(b.ctx, b.botToken, b.RESTInterface, b.Logger)
	b.DonatorBotSession = discord.NewSession(b.ctx, b.donatorBotToken, b.RESTInterface, b.Logger)

	if nginxAddress != "" {
		err = b.Route.SetTrustedProxies([]string{nginxAddress})
		if err != nil {
			return nil, fmt.Errorf("failed to set trusted proxies: %w", err)
		}
	}

	// Setup postgres pool.
	pool, err := pgxpool.Connect(b.ctx, b.postgresAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	b.Pool = pool

	b.Database = database.New(b.Pool)

	// Setup session store.
	store, err := NewStore(b.Pool, []byte("Testing"))
	if err != nil {
		return nil, err
	}

	b.Store = store

	// Load configuration.
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

	// Setup gin router.
	b.Route = b.PrepareGin()

	backend = b

	return b, nil
}

// GetEventContext.
func (b *Backend) GetBasicEventContext() (client *sandwich.EventContext) {
	return &sandwich.EventContext{
		Context: b.ctx,
		Logger:  b.Logger,
		Sandwich: &sandwich.Sandwich{
			SandwichClient: b.SandwichClient,
		},
	}
}

// LoadConfiguration handles loading the configuration file.
func (b *Backend) LoadConfiguration(path string) (configuration *Configuration, err error) {
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

	configuration = &Configuration{}

	err = yaml.Unmarshal(file, configuration)
	if err != nil {
		return configuration, ErrLoadConfigurationFailure
	}

	return configuration, nil
}

// Open sets up any services and starts the webserver.
func (b *Backend) Open() error {
	b.StartTime = time.Now().UTC()
	b.Logger.Info().Msgf("Starting sandwich. Version %s", VERSION)

	go b.PublishSimpleWebhook(b.EmptySession, "Starting backend", "", "Version "+VERSION, EmbedColourSandwich)

	// Setup Prometheus
	go b.SetupPrometheus()

	b.Logger.Info().Msgf("Serving http at %s", b.host)

	err := b.Route.Run(b.host)
	if err != nil {
		return fmt.Errorf("failed to run gin: %w", err)
	}

	return nil
}

// Close gracefully closes the backend.
func (b *Backend) Close() error {
	// TODO

	return nil
}

// SetupPrometheus sets up prometheus.
func (b *Backend) SetupPrometheus() error {
	b.Logger.Info().Msgf("Serving prometheus at %s", b.prometheusAddress)

	b.PrometheusHandler.SetListenAddress(b.prometheusAddress)
	b.PrometheusHandler.SetMetricsPath(nil)

	return nil
}

// PrepareGin prepares gin routes and middleware.
func (b *Backend) PrepareGin() *gin.Engine {
	router := gin.New()
	router.TrustedPlatform = gin.PlatformCloudflare

	_ = router.SetTrustedProxies(nil)

	router.Use(logger.SetLogger())
	router.Use(b.PrometheusHandler.HandlerFunc())
	router.Use(limits.RequestSizeLimiter(RequestSizeLimit))
	router.Use(sessions.Sessions("session", b.Store))
	router.Use(gzip.Gzip(gzip.DefaultCompression))

	router.Use(gin.Recovery())

	registerExampleRoutes(router)

	registerSessionRoutes(router)
	registerUserRoutes(router)

	registerGuildRoutes(router)
	registerGuildSettingsRoutes(router)

	registerGuildSettingsAutoRolesRoutes(router)
	registerGuildSettingsBorderwallRoutes(router)
	registerGuildSettingsFreeRolesRoutes(router)
	registerGuildSettingsLeaverRoutes(router)
	registerGuildSettingsRulesRoutes(router)
	registerGuildSettingsTempChannelsRoutes(router)
	registerGuildSettingsTimeRolesRoutes(router)
	registerGuildSettingsWelcomerRoutes(router)

	return router
}
