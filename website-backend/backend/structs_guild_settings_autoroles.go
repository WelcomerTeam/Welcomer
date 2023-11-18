package backend

import (
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
)

type GuildSettingsAutoRoles struct {
	ToggleEnabled bool     `json:"enabled"`
	Roles         []string `json:"roles"`
}

func GuildSettingsAutoRolesSettingsToPartial(
	autoRoles *database.GuildSettingsAutoroles,
) *GuildSettingsAutoRoles {
	partial := &GuildSettingsAutoRoles{
		ToggleEnabled: autoRoles.ToggleEnabled,
		Roles:         Int64SliceToString(autoRoles.Roles),
	}

	if len(partial.Roles) == 0 {
		partial.Roles = make([]string, 0)
	}

	return partial
}

func PartialToGuildSettingsAutoRolesSettings(guildID int64, guildSettings *GuildSettingsAutoRoles) *database.GuildSettingsAutoroles {
	return &database.GuildSettingsAutoroles{
		GuildID:       guildID,
		ToggleEnabled: guildSettings.ToggleEnabled,
		Roles:         StringSliceToInt64(guildSettings.Roles),
	}
}
