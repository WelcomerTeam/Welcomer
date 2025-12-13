package main

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/WelcomerTeam/Discord/discord"
	"github.com/WelcomerTeam/Sandwich-Daemon/pkg/syncmap"
	subway "github.com/WelcomerTeam/Subway/subway"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	interactions "github.com/WelcomerTeam/Welcomer/welcomer-interactions"
	_ "github.com/joho/godotenv/autoload"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type PublicKeyHandler struct {
	PublicKeys  syncmap.Map[string, ed25519.PublicKey]
	DefaultKeys []ed25519.PublicKey
}

func NewPublicKeyHandler() *PublicKeyHandler {
	return &PublicKeyHandler{
		PublicKeys:  syncmap.Map[string, ed25519.PublicKey]{},
		DefaultKeys: make([]ed25519.PublicKey, 0),
	}
}

func stringToPublicKey(key string) (ed25519.PublicKey, error) {
	if !welcomer.IsValidPublicKey(key) {
		return nil, fmt.Errorf("invalid public key format: %s", key)
	}

	hex, err := hex.DecodeString(key)
	if err != nil {
		return nil, fmt.Errorf("failed to decode public key %s: %w", key, err)
	}

	return ed25519.PublicKey(hex), nil
}

func (h *PublicKeyHandler) SetDefaultPublicKeys(keys []string) error {
	h.DefaultKeys = make([]ed25519.PublicKey, 0, len(keys))

	for _, key := range keys {
		publicKey, err := stringToPublicKey(key)
		if err != nil {
			return fmt.Errorf("failed to set default public key %s: %w", key, err)
		}

		h.DefaultKeys = append(h.DefaultKeys, publicKey)
	}

	return nil
}

func (h *PublicKeyHandler) SetPublicKey(manager, key string) error {
	if manager == "" {
		return fmt.Errorf("manager name cannot be empty")
	}

	publicKey, err := stringToPublicKey(key)
	if err != nil {
		return fmt.Errorf("failed to set public key for manager %s: %w", manager, err)
	}

	h.PublicKeys.Store(manager, publicKey)

	return nil
}

func (h *PublicKeyHandler) GetPublicKeys(r *http.Request) []ed25519.PublicKey {
	keys := make([]ed25519.PublicKey, 0)
	keys = append(keys, h.DefaultKeys...)

	manager := r.URL.Query().Get("manager")
	if manager == "" {
		welcomer.Logger.Warn().Msg("No manager specified in request path")

		return keys
	}

	key, ok := h.PublicKeys.Load(manager)
	if ok {
		keys = append(keys, key)
	} else {
		welcomer.Logger.Warn().Str("manager", manager).Msg("No public key found for manager")
	}

	return keys
}

var PublicKeyHandlerInstance = NewPublicKeyHandler()

func main() {
	loggingLevel := flag.String("level", os.Getenv("LOGGING_LEVEL"), "Logging level")

	sandwichGRPCHost := flag.String("sandwichGRPCHost", os.Getenv("SANDWICH_GRPC_HOST"), "GRPC Address for the Sandwich Daemon service")

	proxyAddress := flag.String("proxyAddress", os.Getenv("PROXY_ADDRESS"), "Address to proxy requests through. This can be 'https://discord.com', if one is not setup")

	proxyDebug := flag.Bool("proxyDebug", false, "Enable debugging requests to the proxy")

	prometheusAddress := flag.String("prometheusAddress", os.Getenv("INTERACTIONS_PROMETHEUS_ADDRESS"), "Prometheus address")

	postgresURL := flag.String("postgresURL", os.Getenv("POSTGRES_URL"), "Postgres connection URL")

	publicKeys := flag.String("publicKey", os.Getenv("INTERACTIONS_PUBLIC_KEY"), "Public key(s) for signature validation. Comma delimited.")

	flag.Parse()

	var err error

	ctx, _ := context.WithCancel(context.Background())

	restInterface := welcomer.NewTwilightProxy(*proxyAddress)
	restInterface.SetDebug(*proxyDebug)

	welcomer.SetupLogger(*loggingLevel)
	welcomer.SetupGRPCConnection(*sandwichGRPCHost,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(1024*1024*1024)), // Set max message size to 1GB
	)
	welcomer.SetupRESTInterface(restInterface)
	welcomer.SetupSandwichClient()
	welcomer.SetupDatabase(ctx, *postgresURL)

	app := interactions.NewWelcomer(ctx, subway.SubwayOptions{
		SandwichClient:    welcomer.SandwichClient,
		RESTInterface:     welcomer.RESTInterface,
		Logger:            slog.Default(),
		PublicKeysHandler: PublicKeyHandlerInstance,
		PrometheusAddress: *prometheusAddress,
	})

	err = PublicKeyHandlerInstance.SetDefaultPublicKeys(strings.Split(*publicKeys, ","))
	if err != nil {
		panic(fmt.Errorf("failed to set default public keys: %w", err))
	}

	err = FetchPublicKeys(ctx)
	if err != nil {
		panic(fmt.Errorf("failed to fetch public keys: %w", err))
	}

	configurationGatherer := welcomer.GetConfigurationGatherer(ctx)

	configuration, err := configurationGatherer.GetConfig(ctx)
	if err != nil {
		panic(fmt.Errorf("failed to gather configuration: %w", err))
	}

	applicationCommands := app.Commands.MapApplicationCommands()

	for _, application := range configuration.Applications {
		currentUser, err := discord.GetCurrentUser(ctx, discord.NewSession("Bot "+application.BotToken, restInterface))
		if err != nil {
			slog.Error("Failed to get current user", "application", application.DisplayName, "error", err)
			continue
		}

		session := discord.NewSession("Bot "+application.BotToken, restInterface)

		now := time.Now()

		applicationCommands, err := discord.BulkOverwriteGlobalApplicationCommands(ctx, session, currentUser.ID, applicationCommands)
		if err != nil {
			slog.Error("Failed to sync commands", "applicationID", currentUser.ID, "application", application.DisplayName, "error", err)
			continue
		}

		slog.Info("Successfully synced commands", "applicationID", currentUser.ID, "application", application.DisplayName)

		_, err = welcomer.Queries.ClearInteractionCommands(ctx, int64(currentUser.ID))
		if err != nil {
			slog.Error("Failed to clear interaction commands", "applicationID", currentUser.ID, "application", application.DisplayName, "error", err)
		}

		manyInteractionCommands := make([]database.CreateManyInteractionCommandsParams, 0)

		for _, command := range applicationCommands {
			manyInteractionCommands = append(manyInteractionCommands, database.CreateManyInteractionCommandsParams{
				ApplicationID: int64(currentUser.ID),
				Command:       command.Name,
				InteractionID: int64(*command.ID),
				CreatedAt:     now,
			})
		}

		_, err = welcomer.Queries.CreateManyInteractionCommands(ctx, manyInteractionCommands)
		if err != nil {
			slog.Error("Failed to create many interaction commands", "applicationID", currentUser.ID, "application", application.DisplayName, "error", err)
			continue
		}

		slog.Info("Successfully created interaction commands", "applicationID", currentUser.ID, "application", application.DisplayName)
	}
}

func FetchPublicKeys(ctx context.Context) error {
	customBots, err := welcomer.Queries.GetAllCustomBotsWithToken(ctx, welcomer.GetCustomBotEnvironmentType())
	if err != nil {
		return fmt.Errorf("failed to fetch custom bots: %w", err)
	}

	for _, bot := range customBots {
		if bot.PublicKey != "" {
			if welcomer.IsValidPublicKey(bot.PublicKey) {
				err = PublicKeyHandlerInstance.SetPublicKey(welcomer.GetCustomBotKey(bot.CustomBotUuid), bot.PublicKey)
				if err != nil {
					welcomer.Logger.Error().Err(err).Str("applicationName", bot.ApplicationName).Msg("Failed to set public key for custom bot")
				}
			} else {
				welcomer.Logger.Warn().Str("publicKey", bot.PublicKey).Msg("Invalid public key format for custom bot")
			}
		}
	}

	return nil
}
