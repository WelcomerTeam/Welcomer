package main

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/WelcomerTeam/Discord/discord"
	"github.com/WelcomerTeam/Sandwich-Daemon/pkg/syncmap"
	sandwich_protobuf "github.com/WelcomerTeam/Sandwich-Daemon/proto"
	subway "github.com/WelcomerTeam/Subway/subway"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	interactions "github.com/WelcomerTeam/Welcomer/welcomer-interactions"
	_ "github.com/joho/godotenv/autoload"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	PermissionsDefault = 0o744
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

	sandwichManagerName := flag.String("sandwichManagerName", os.Getenv("SANDWICH_MANAGER_NAME"), "Sandwich manager identifier name")

	sandwichGRPCHost := flag.String("sandwichGRPCHost", os.Getenv("SANDWICH_GRPC_HOST"), "GRPC Address for the Sandwich Daemon service")

	proxyAddress := flag.String("proxyAddress", os.Getenv("PROXY_ADDRESS"), "Address to proxy requests through. This can be 'https://discord.com', if one is not setup")

	proxyDebug := flag.Bool("proxyDebug", false, "Enable debugging requests to the proxy")

	prometheusAddress := flag.String("prometheusAddress", os.Getenv("INTERACTIONS_PROMETHEUS_ADDRESS"), "Prometheus address")

	postgresURL := flag.String("postgresURL", os.Getenv("POSTGRES_URL"), "Postgres connection URL")

	host := flag.String("host", os.Getenv("INTERACTIONS_HOST"), "Host to serve interactions from")

	publicKeys := flag.String("publicKey", os.Getenv("INTERACTIONS_PUBLIC_KEY"), "Public key(s) for signature validation. Comma delimited.")

	dryRun := flag.Bool("dryRun", false, "When true, will close after setting up the app")

	syncCommands := flag.Bool("syncCommands", false, "If true, will update commands")

	flag.Parse()

	var err error

	ctx, cancel := context.WithCancel(context.Background())

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

	// Setup app.

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

	mux := http.NewServeMux()

	mux.HandleFunc("/internal/sync-commands", func(w http.ResponseWriter, r *http.Request) {
		body := struct {
			Token         string            `json:"token"`
			ApplicationID discord.Snowflake `json:"application_id"`
		}{}

		defer r.Body.Close()

		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			welcomer.Logger.Error().Err(err).Msg("Failed to decode request body")
			http.Error(w, "Invalid request body", http.StatusBadRequest)

			return
		}

		now := time.Now()

		applicationCommands := app.Commands.MapApplicationCommands()

		session := discord.NewSession("Bot "+body.Token, welcomer.RESTInterface)

		applicationCommands, err := discord.BulkOverwriteGlobalApplicationCommands(ctx, session, body.ApplicationID, applicationCommands)
		if err != nil {
			welcomer.Logger.Error().Err(err).Msg("Failed to sync commands")
			http.Error(w, fmt.Sprintf("Failed to sync commands: %v", err), http.StatusInternalServerError)

			return
		}

		// Copy of code from jobs/sync-interaction-commands.go

		_, err = welcomer.Queries.ClearInteractionCommands(ctx, int64(body.ApplicationID))
		if err != nil {
			welcomer.Logger.Error().Err(err).Msg("Failed to clear interaction commands")
			http.Error(w, fmt.Sprintf("Failed to clear interaction commands: %v", err), http.StatusInternalServerError)

			return
		}

		manyInteractionCommands := make([]database.CreateManyInteractionCommandsParams, 0)

		for _, command := range applicationCommands {
			manyInteractionCommands = append(manyInteractionCommands, database.CreateManyInteractionCommandsParams{
				ApplicationID: int64(body.ApplicationID),
				Command:       command.Name,
				InteractionID: int64(*command.ID),
				CreatedAt:     now,
			})
		}

		_, err = welcomer.Queries.CreateManyInteractionCommands(ctx, manyInteractionCommands)
		if err != nil {
			welcomer.Logger.Error().Err(err).Msg("Failed to create many interaction commands")
			http.Error(w, fmt.Sprintf("Failed to create many interaction commands: %v", err), http.StatusInternalServerError)

			return
		}

		w.WriteHeader(http.StatusNoContent)
	})

	mux.HandleFunc("/internal/fetch-public-keys", func(w http.ResponseWriter, _ *http.Request) {
		err = FetchPublicKeys(ctx)
		if err != nil {
			welcomer.Logger.Error().Err(err).Msg("Failed to update public keys")

			http.Error(w, fmt.Sprintf("Failed to update public keys: %v", err), http.StatusInternalServerError)

			return
		}

		w.WriteHeader(http.StatusNoContent)
	})

	if err != nil {
		welcomer.Logger.Panic().Err(err).Msg("Exception creating app")
	}

	if *syncCommands {
		configurationsPb, err := welcomer.SandwichClient.FetchApplication(ctx, &sandwich_protobuf.ApplicationIdentifier{
			ApplicationIdentifier: *sandwichManagerName,
		})
		if err != nil {
			panic(fmt.Errorf(`failed to sync command: grpcInterface.FetchConsumerConfiguration(): %w`, err))
		}

		configuration, ok := configurationsPb.GetApplications()[*sandwichManagerName]

		if !ok {
			panic(fmt.Errorf(`failed to sync command: could not find manager matching "%s"`, *sandwichManagerName))
		}

		err = app.SyncCommands(ctx, "Bot "+configuration.BotToken, discord.Snowflake(configuration.UserId))
		if err != nil {
			panic(fmt.Errorf(`failed to sync commands. app.SyncCommands(): %w`, err))
		}

		welcomer.Logger.Info().Msg("Successfully synced commands")

	}

	// We return if it a dry run. Any issues loading up the bot would've already caused a panic.

	if *dryRun {
		return
	}

	if err = app.ListenAndServe("", *host, mux); err != nil {
		welcomer.Logger.Panic().Str("host", *host).Err(err).Msg("Failed to serve interactions server")
	}

	cancel()

	if err = welcomer.GRPCConnection.Close(); err != nil {
		welcomer.Logger.Warn().Err(err).Msg("Exception whilst closing grpc client")
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
