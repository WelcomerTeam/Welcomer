package welcomer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/WelcomerTeam/Discord/discord"
	sandwich_daemon "github.com/WelcomerTeam/Sandwich-Daemon"
	sandwich_protobuf "github.com/WelcomerTeam/Sandwich-Daemon/proto"
	"github.com/gofrs/uuid"
	"google.golang.org/grpc/status"
)

func GetCustomBotKey(v uuid.UUID) string {
	return "custom_bot_" + v.String()
}

func GetGuildCustomBotLimit(ctx context.Context, guildID discord.Snowflake) int {
	hasWelcomerPro, _, _, _ := CheckGuildMemberships(ctx, guildID)
	if hasWelcomerPro {
		return 2
	}

	return 0
}

func GetCustomBotConfiguration(customBotUUID uuid.UUID, botToken string, autoStart bool, guildID discord.Snowflake) sandwich_daemon.ApplicationConfiguration {
	return sandwich_daemon.ApplicationConfiguration{
		ApplicationIdentifier: GetCustomBotKey(customBotUUID),
		ProducerIdentifier:    os.Getenv("CUSTOM_BOT_PRODUCER_IDENTIFIER"),
		DisplayName:           fmt.Sprintf("Custom Bot %s", customBotUUID.String()),
		ClientName:            GetCustomBotKey(customBotUUID),
		BotToken:              botToken,
		AutoStart:             autoStart,
		Intents:               32511,
		AutoSharded:           true,
		Values: map[string]any{
			"public":     false,
			"custom_bot": true,
			"bot_uuid":   customBotUUID,
			"guild_id":   guildID,
		},

		IncludeRandomSuffix: false,
		DefaultPresence:     discord.UpdateStatus{},
		ChunkGuildsOnStart:  false,
		EventBlacklist:      []string{},
		ProduceBlacklist:    []string{},
		ShardCount:          0,
		ShardIDs:            "",
	}
}

func StartCustomBot(ctx context.Context, customBotUUID uuid.UUID, botToken string, guildID discord.Snowflake, blocking bool) error {
	identifier := GetCustomBotKey(customBotUUID)

	applications, err := SandwichClient.FetchApplication(ctx, &sandwich_protobuf.ApplicationIdentifier{
		ApplicationIdentifier: identifier,
	})

	application, ok := applications.GetApplications()[identifier]

	if err != nil || !ok || application.GetBotToken() != botToken {
		configuration := GetCustomBotConfiguration(customBotUUID, botToken, true, guildID)

		values, err := json.Marshal(configuration.Values)
		if err != nil {
			return fmt.Errorf("failed to marshal custom bot values: %w", err)
		}

		application, err = SandwichClient.CreateApplication(ctx, &sandwich_protobuf.CreateApplicationRequest{
			SaveConfig:            false,
			ApplicationIdentifier: configuration.ApplicationIdentifier,
			ProducerIdentifier:    configuration.ProducerIdentifier,
			DisplayName:           configuration.DisplayName,
			ClientName:            configuration.ClientName,
			BotToken:              configuration.BotToken,
			AutoStart:             configuration.AutoStart,
			Intents:               configuration.Intents,
			AutoSharded:           configuration.AutoSharded,
			Values:                values,
		})
	}

	resp, err := SandwichClient.StartApplication(ctx, &sandwich_protobuf.ApplicationIdentifierWithBlocking{
		ApplicationIdentifier: identifier,
		Blocking:              blocking,
	})

	// TODO: Add status codes to sandwich daemon for better error handling
	if err != nil && status.Convert(err).Message() == "application already running" {
		return nil // Application is already running, so do not need to do anything
	}

	if err != nil {
		return fmt.Errorf("failed to start custom bot %s: %w", identifier, err)
	}

	if !resp.GetOk() {
		return fmt.Errorf("failed to start custom bot %s: %s", identifier, resp.GetError())
	}

	return nil
}

func StopCustomBot(ctx context.Context, customBotUUID uuid.UUID, blocking bool) error {
	identifier := GetCustomBotKey(customBotUUID)

	resp, err := SandwichClient.StopApplication(ctx, &sandwich_protobuf.ApplicationIdentifierWithBlocking{
		ApplicationIdentifier: identifier,
		Blocking:              blocking,
	})

	// TODO: Add status codes to sandwich daemon for better error handling
	if err != nil && status.Convert(err).Message() == "application not found" {
		// Application is not in sandwich daemon, so do not need to do anything
		return nil
	}

	if err != nil {
		return fmt.Errorf("failed to stop custom bot: %w", err)
	}

	if !resp.GetOk() {
		return fmt.Errorf("failed to stop custom bot: %s", resp.GetError())
	}

	// Cleanup the application if it was stopped successfully.

	resp, err = SandwichClient.DeleteApplication(ctx, &sandwich_protobuf.ApplicationIdentifier{
		ApplicationIdentifier: identifier,
	})

	if err != nil && status.Convert(err).Message() == "application not found" {
		// Application is not in sandwich daemon, so do not need to do anything
		return nil
	}

	if err != nil {
		return fmt.Errorf("failed to delete custom bot %s: %w", identifier, err)
	}

	if !resp.GetOk() {
		return fmt.Errorf("failed to delete custom bot %s: %s", identifier, resp.GetError())
	}

	return nil
}
