package backend

import (
	"context"
	"fmt"
	"io"
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
	"google.golang.org/grpc"
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

	ctx context.Context

	Logger    zerolog.Logger `json:"-"`
	StartTime time.Time      `json:"start_time" yaml:"start_time"`

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

	// Environment Variables.
	host              string
	prometheusAddress string
	postgresAddress   string
	nginxAddress      string
}

// NewBackend creates a new backend.
func NewBackend(ctx context.Context, conn grpc.ClientConnInterface, restInterface discord.RESTInterface, logger io.Writer, pool *pgxpool.Pool, host, botToken, donatorBotToken, prometheusAddress, postgresAddress, nginxAddress, clientId, clientSecret, redirectURL string) (b *Backend, err error) {
	if backend != nil {
		return backend, ErrBackendAlreadyExists
	}

	b = &Backend{
		Logger: zerolog.New(logger).With().Timestamp().Logger(),

		RESTInterface: restInterface,

		SandwichClient: protobuf.NewSandwichClient(conn),
		GRPCInterface:  sandwich.NewDefaultGRPCClient(),

		PrometheusHandler: ginprometheus.NewPrometheus("gin"),

		host:              host,
		botToken:          botToken,
		donatorBotToken:   donatorBotToken,
		prometheusAddress: prometheusAddress,
		postgresAddress:   postgresAddress,
		nginxAddress:      nginxAddress,
	}

	b.Lock()
	defer b.Unlock()

	b.ctx = ctx

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

	b.Pool = pool
	b.Database = database.New(b.Pool)

	// Setup session store.
	store, err := NewStore(b.Pool, []byte("Testing"))
	if err != nil {
		return nil, err
	}

	b.Store = store

	b.Logger = zerolog.New(logger).With().Timestamp().Logger()
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

// Open sets up any services and starts the webserver.
func (b *Backend) Open() error {
	b.StartTime = time.Now().UTC()
	b.Logger.Info().Msgf("Starting backend. Version %s", VERSION)

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
