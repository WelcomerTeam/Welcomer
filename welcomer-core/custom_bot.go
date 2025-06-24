package welcomer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/WelcomerTeam/Discord/discord"
	sandwich_daemon "github.com/WelcomerTeam/Sandwich-Daemon"
	sandwich_protobuf "github.com/WelcomerTeam/Sandwich-Daemon/proto"
	"github.com/gofrs/uuid"
)

func GetCustomBotKey(v uuid.UUID) string {
	return "custom_bot_" + v.String()
}

func GetCustomBotConfiguration(customBotUUID uuid.UUID, botToken string) sandwich_daemon.ApplicationConfiguration {
	return sandwich_daemon.ApplicationConfiguration{
		ApplicationIdentifier: GetCustomBotKey(customBotUUID),
		ProducerIdentifier:    "welcomer",
		DisplayName:           fmt.Sprintf("Custom Bot %s", customBotUUID.String()),
		ClientName:            GetCustomBotKey(customBotUUID),
		BotToken:              botToken,
		AutoStart:             true,
		Intents:               32511,
		AutoSharded:           true,
		Values: map[string]any{
			"public":     false,
			"custom_bot": true,
			"bot_uuid":   customBotUUID,
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

func StartCustomBot(ctx context.Context, customBotUUID uuid.UUID, botToken string, blocking bool) error {
	identifier := GetCustomBotKey(customBotUUID)

	applications, err := SandwichClient.FetchApplication(ctx, &sandwich_protobuf.ApplicationIdentifier{
		ApplicationIdentifier: identifier,
	})

	application, ok := applications.GetApplications()[identifier]

	if err != nil || !ok || application.GetBotToken() != botToken {
		configuration := GetCustomBotConfiguration(customBotUUID, botToken)

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
	if err != nil {
		return fmt.Errorf("failed to start custom bot: %w", err)
	}

	if !resp.GetOk() {
		return fmt.Errorf("failed to start custom bot: %s", resp.GetError())
	}

	return nil
}

func StopCustomBot(ctx context.Context, customBotUUID uuid.UUID, blocking bool) error {
	identifier := GetCustomBotKey(customBotUUID)

	resp, err := SandwichClient.StopApplication(ctx, &sandwich_protobuf.ApplicationIdentifierWithBlocking{
		ApplicationIdentifier: identifier,
		Blocking:              blocking,
	})
	if err != nil {
		return fmt.Errorf("failed to stop custom bot: %w", err)
	}

	if !resp.GetOk() {
		return fmt.Errorf("failed to stop custom bot: %s", resp.GetError())
	}

	return nil
}
