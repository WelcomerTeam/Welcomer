package welcomer

import (
	"context"
	"fmt"
	"os"

	"github.com/WelcomerTeam/Discord/discord"
	sandwich_daemon "github.com/WelcomerTeam/Sandwich-Daemon"
)

type EnvironmentType int

const (
	EnvironmentTypeDevelopment EnvironmentType = iota
	EnvironmentTypeProduction
	EnvironmentTypeProductionOutsideDocker
)

func getEnvironmentType() EnvironmentType {
	environmentEnv := os.Getenv("ENVIRONMENT")

	switch environmentEnv {
	case "development":
		return EnvironmentTypeDevelopment
	case "production":
		return EnvironmentTypeProduction
	case "production-outside-docker":
		return EnvironmentTypeProductionOutsideDocker
	default:
		panic(fmt.Sprintf("invalid environment type: %s", environmentEnv))
	}
}

type WelcomerDatabaseConfigProvider struct {
	DefaultConfigProvider sandwich_daemon.ConfigProvider
}

func GetConfigurationGatherer(ctx context.Context) sandwich_daemon.ConfigProvider {
	environmentType := getEnvironmentType()

	var fileConfigProvider sandwich_daemon.ConfigProvider

	switch environmentType {
	case EnvironmentTypeProduction:
		fileConfigProvider = sandwich_daemon.NewConfigProviderFromPath(os.Getenv("CONFIGURATION_LOCATION"))
	case EnvironmentTypeProductionOutsideDocker:
		fileConfigProvider = sandwich_daemon.NewConfigProviderFromPath(os.Getenv("CONFIGURATION_LOCATION_PRODUCTION"))
	case EnvironmentTypeDevelopment:
		fileConfigProvider = sandwich_daemon.NewConfigProviderFromPath(os.Getenv("CONFIGURATION_LOCATION"))
	}

	return &WelcomerDatabaseConfigProvider{fileConfigProvider}
}

func (p *WelcomerDatabaseConfigProvider) GetConfig(ctx context.Context) (*sandwich_daemon.Configuration, error) {
	config, err := p.DefaultConfigProvider.GetConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get configuration: %w", err)
	}

	customBots, err := Queries.GetAllCustomBotsWithToken(ctx)
	if err != nil {
		panic(fmt.Errorf("failed to get custom bots: %w", err))
	}

	for _, customBot := range customBots {
		if customBot.Token == "" {
			Logger.Error().
				Str("application_name", customBot.ApplicationName).
				Int64("application_id", customBot.ApplicationID).
				Msg("Custom bot token is empty, skipping configuration for this bot")

			continue
		}

		decryptedBotToken, err := DecryptBotToken(customBot.Token, customBot.CustomBotUuid)
		if err != nil {
			Logger.Error().
				Err(err).
				Str("application_name", customBot.ApplicationName).
				Int64("application_id", customBot.ApplicationID).
				Msgf("Failed to decrypt bot token for application")

			continue
		}

		applicationConfiguration := GetCustomBotConfiguration(customBot.CustomBotUuid, decryptedBotToken, customBot.IsActive, discord.Snowflake(customBot.GuildID))
		config.Applications = append(config.Applications, &applicationConfiguration)
	}

	return config, nil
}

// Not implemented, as the configuration is read-only in this context.
func (p *WelcomerDatabaseConfigProvider) SaveConfig(ctx context.Context, config *sandwich_daemon.Configuration) error {
	return nil
}
