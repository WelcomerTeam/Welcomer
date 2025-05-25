package backend

import (
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	discord "github.com/WelcomerTeam/Discord/discord"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/logger"
	"github.com/gin-contrib/sessions"
	limits "github.com/gin-contrib/size"
	"github.com/gin-gonic/gin"
	"github.com/plutov/paypal/v4"
	gin_prometheus "github.com/zsais/go-gin-prometheus"
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
	StartTime time.Time

	Options Options

	PrometheusHandler *gin_prometheus.Prometheus

	Route *gin.Engine

	Store Store

	IPChecker welcomer.IPChecker

	EmptySession      *discord.Session
	BotSession        *discord.Session
	DonatorBotSession *discord.Session

	PaypalClient *paypal.Client
}

// Options represents any options passable when creating the backend service.
type Options struct {
	Domain            string
	Host              string
	KeyPairs          string
	NginxAddress      string
	PostgresAddress   string
	PrometheusAddress string

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
func NewBackend(options Options) (*Backend, error) {
	if backend != nil {
		return backend, ErrBackendAlreadyExists
	}

	b := &Backend{
		Options:           options,
		PrometheusHandler: gin_prometheus.NewPrometheus("gin"),
		IPChecker:         welcomer.NewLRUIPChecker(1024),
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
	b.EmptySession = discord.NewSession("", welcomer.RESTInterface)
	b.BotSession = discord.NewSession(b.Options.BotToken, welcomer.RESTInterface)
	b.DonatorBotSession = discord.NewSession(b.Options.DonatorBotToken, welcomer.RESTInterface)

	println(welcomer.RESTInterface)

	if options.NginxAddress != "" {
		err := b.Route.SetTrustedProxies([]string{options.NginxAddress})
		if err != nil {
			return nil, fmt.Errorf("failed to set trusted proxies: %w", err)
		}
	}

	keyPairs := [][]byte{}

	keyPairStrings := strings.Split(options.KeyPairs, ",")
	for _, keyPairString := range keyPairStrings {
		byteSlice, err := hex.DecodeString(keyPairString)
		if err != nil {
			return nil, fmt.Errorf("failed to parse keyPairString %s to hex: %w", keyPairString, err)
		}

		keyPairs = append(keyPairs, byteSlice)
	}

	// Setup session store.
	store, err := NewStore(welcomer.Pool, keyPairs...)
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

	paypalClient, err := paypal.NewClient(options.PaypalClientID, options.PaypalClientSecret, welcomer.If(options.PaypalIsLive, paypal.APIBaseLive, paypal.APIBaseSandBox))
	if err != nil {
		return nil, fmt.Errorf("failed to create paypal client: %w", err)
	}

	b.PaypalClient = paypalClient

	// Setup gin router.
	b.Route = b.PrepareGin()

	backend = b

	return b, nil
}

// Open sets up any services and starts the web server.
func (b *Backend) Open() error {
	b.StartTime = time.Now()
	welcomer.Logger.Info().Msgf("Starting backend. Version %s", VERSION)

	// Setup Prometheus
	go b.SetupPrometheus()

	welcomer.Logger.Info().Msgf("Serving http at %s", b.Options.Host)

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
	welcomer.Logger.Info().Msgf("Serving prometheus at %s", b.Options.PrometheusAddress)

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
