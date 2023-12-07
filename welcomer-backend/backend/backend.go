package backend

import (
	"context"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
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

	MaxAge = time.Hour * 24 * 7
)

var backend *Backend

type Backend struct {
	ctx context.Context

	Logger    zerolog.Logger
	StartTime time.Time

	Options BackendOptions

	RESTInterface discord.RESTInterface

	SandwichClient protobuf.SandwichClient
	GRPCInterface  sandwich.GRPC

	PrometheusHandler *ginprometheus.Prometheus

	Route *gin.Engine

	Store Store

	Database *database.Queries

	EmptySession      *discord.Session
	BotSession        *discord.Session
	DonatorBotSession *discord.Session
}

// BackendOptions represents any options passable when creating
// the backend service
type BackendOptions struct {
	BotToken          string
	ClientId          string
	ClientSecret      string
	Conn              grpc.ClientConnInterface
	DonatorBotToken   string
	Host              string
	NginxAddress      string
	Pool              *pgxpool.Pool
	PostgresAddress   string
	PrometheusAddress string
	RedirectURL       string
	RESTInterface     discord.RESTInterface
	KeyPairs          string
}

// NewBackend creates a new backend.
func NewBackend(ctx context.Context, logger zerolog.Logger, options BackendOptions) (b *Backend, err error) {
	if backend != nil {
		return backend, ErrBackendAlreadyExists
	}

	b = &Backend{
		ctx: ctx,

		Logger: logger,

		Options: options,

		RESTInterface: options.RESTInterface,

		SandwichClient: protobuf.NewSandwichClient(options.Conn),
		GRPCInterface:  sandwich.NewDefaultGRPCClient(),

		PrometheusHandler: ginprometheus.NewPrometheus("gin"),

		Database: database.New(options.Pool),
	}

	// Setup OAuth2
	OAuth2Config.ClientID = options.ClientId
	OAuth2Config.ClientSecret = options.ClientSecret
	OAuth2Config.RedirectURL = options.RedirectURL

	// Setup sessions
	b.EmptySession = discord.NewSession(b.ctx, "", b.RESTInterface, b.Logger)
	b.BotSession = discord.NewSession(b.ctx, b.Options.BotToken, b.RESTInterface, b.Logger)
	b.DonatorBotSession = discord.NewSession(b.ctx, b.Options.DonatorBotToken, b.RESTInterface, b.Logger)

	if options.NginxAddress != "" {
		err = b.Route.SetTrustedProxies([]string{options.NginxAddress})
		if err != nil {
			return nil, fmt.Errorf("failed to set trusted proxies: %w", err)
		}
	}

	keyPairs := [][]byte{}

	keyPairStrings := strings.Split(options.KeyPairs, ",")
	for _, keyPairString := range keyPairStrings {
		byteSlice, err := hex.DecodeString(keyPairString)
		if err != nil {
			return nil, fmt.Errorf("failed to parse keyPairString %s to hex: %v", keyPairString, err.Error())
		}

		keyPairs = append(keyPairs, byteSlice)
	}

	// Setup session store.
	store, err := NewStore(options.Pool, keyPairs...)
	if err != nil {
		return nil, err
	}

	store.Options(sessions.Options{
		Path:     "",
		Domain:   "",
		MaxAge:   int(MaxAge),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	b.Store = store

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

	b.Logger.Info().Msgf("Serving http at %s", b.Options.Host)

	err := b.Route.Run(b.Options.Host)
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
	b.Logger.Info().Msgf("Serving prometheus at %s", b.Options.PrometheusAddress)

	b.PrometheusHandler.SetListenAddress(b.Options.PostgresAddress)
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
