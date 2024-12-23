package backend

import (
	"context"
	"encoding/hex"
	"fmt"
	discord "github.com/WelcomerTeam/Discord/discord"
	protobuf "github.com/WelcomerTeam/Sandwich-Daemon/protobuf"
	sandwich "github.com/WelcomerTeam/Sandwich/sandwich"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	utils "github.com/WelcomerTeam/Welcomer/welcomer-utils"
	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/logger"
	"github.com/gin-contrib/sessions"
	limits "github.com/gin-contrib/size"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/plutov/paypal/v4"
	"github.com/rs/zerolog"
	gin_prometheus "github.com/zsais/go-gin-prometheus"
	"google.golang.org/grpc"
	"net/http"
	"os"
	"strings"
	"time"
)

const VERSION = "0.1"

const (
	PermissionsDefault = 0o744
	PermissionWrite    = 0o600

	RequestSizeLimit = 20_000_000

	MaxAge = time.Hour * 24 * 7 * 4 // 4 weeks
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

	PrometheusHandler *gin_prometheus.Prometheus

	Route *gin.Engine

	Store Store

	Database *database.Queries
	Pool     *pgxpool.Pool

	IPChecker utils.IPChecker

	EmptySession      *discord.Session
	BotSession        *discord.Session
	DonatorBotSession *discord.Session

	PaypalClient *paypal.Client
}

// BackendOptions represents any options passable when creating the backend service.
type BackendOptions struct {
	Domain            string
	Host              string
	KeyPairs          string
	NginxAddress      string
	PostgresAddress   string
	PrometheusAddress string

	Conn          grpc.ClientConnInterface
	RESTInterface discord.RESTInterface
	Pool          *pgxpool.Pool

	BotToken        string
	DonatorBotToken string

	DiscordClientID     string
	DiscordClientSecret string
	DiscordRedirectURL  string

	PatreonClientID     string
	PatreonClientSecret string
	PatreonRedirectURL  string

	PaypalClientID     string
	PaypalClientSecret string
	PaypalIsLive       bool
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

		PrometheusHandler: gin_prometheus.NewPrometheus("gin"),

		Database: database.New(options.Pool),
		Pool:     options.Pool,

		IPChecker: utils.NewLRUIPChecker(logger, 1024),
	}

	// Setup Discord OAuth2
	DiscordOAuth2Config.ClientID = options.DiscordClientID
	DiscordOAuth2Config.ClientSecret = options.DiscordClientSecret
	DiscordOAuth2Config.RedirectURL = options.DiscordRedirectURL

	// Setup Patreon OAuth2
	PatreonOAuth2Config.ClientID = options.PatreonClientID
	PatreonOAuth2Config.ClientSecret = options.PatreonClientSecret
	PatreonOAuth2Config.RedirectURL = options.PatreonRedirectURL

	// Setup sessions
	b.EmptySession = discord.NewSession(b.ctx, "", b.RESTInterface)
	b.BotSession = discord.NewSession(b.ctx, b.Options.BotToken, b.RESTInterface)
	b.DonatorBotSession = discord.NewSession(b.ctx, b.Options.DonatorBotToken, b.RESTInterface)

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
		return nil, fmt.Errorf("failed to create session store: %w", err)
	}

	store.Options(sessions.Options{
		Path:     "/",
		Domain:   options.Domain,
		MaxAge:   int(MaxAge.Seconds()),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	b.Store = store

	paypalClient, err := paypal.NewClient(options.PaypalClientID, options.PaypalClientSecret, utils.If(options.PaypalIsLive, paypal.APIBaseLive, paypal.APIBaseSandBox))
	if err != nil {
		return nil, fmt.Errorf("failed to create paypal client: %w", err)
	}

	b.PaypalClient = paypalClient

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

// Open sets up any services and starts the web server.
func (b *Backend) Open() error {
	b.StartTime = time.Now()
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

	b.PrometheusHandler.SetListenAddress(b.Options.PrometheusAddress)
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
	router.Use(sessions.Sessions(os.Getenv("SESSION_COOKIE_NAME"), b.Store))
	router.Use(gzip.Gzip(gzip.DefaultCompression))

	router.Use(gin.Recovery())

	registerExampleRoutes(router)

	registerMetaRoutes(router)

	registerSessionRoutes(router)
	registerUserRoutes(router)

	registerBillingRoutes(router)
	registerMembershipsRoutes(router)
	registerPatreonRoutes(router)

	registerBorderwallRoutes(router)

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
