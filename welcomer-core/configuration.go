package welcomer

import (
	"os"

	sandwich_daemon "github.com/WelcomerTeam/Sandwich-Daemon"
)

func GatherConfiguration() sandwich_daemon.ConfigProviderFromPath {
	return sandwich_daemon.NewConfigProviderFromPath(os.Getenv("CONFIGURATION_LOCATION"))
}
