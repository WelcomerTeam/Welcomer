package welcomer

import (
	"fmt"
	"os"

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

func GatherConfiguration() sandwich_daemon.ConfigProvider {
	environmentType := getEnvironmentType()

	switch environmentType {
	case EnvironmentTypeProduction:
		return sandwich_daemon.NewConfigProviderFromPath(os.Getenv("CONFIGURATION_LOCATION"))
	case EnvironmentTypeProductionOutsideDocker:
		return sandwich_daemon.NewConfigProviderFromPath(os.Getenv("CONFIGURATION_LOCATION_PRODUCTION"))
	case EnvironmentTypeDevelopment:
		return sandwich_daemon.NewConfigProviderFromPath(os.Getenv("CONFIGURATION_LOCATION"))
	}

	return nil
}
