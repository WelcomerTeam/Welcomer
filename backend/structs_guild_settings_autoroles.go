package backend

import (
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
)

type GuildSettingsAutoRoles struct {
	ToggleEnabled bool    `json:"toggle_enabled"`
	Roles         []int64 `json:"roles"`
}

func GuildSettingsAutoRolesSettingsToPartial(
	freeRoles *database.GuildSettingsAutoroles,
) *GuildSettingsAutoRoles {
	partial := &GuildSettingsAutoRoles{
		ToggleEnabled: freeRoles.ToggleEnabled,
		Roles:         freeRoles.Roles,
	}

	if len(partial.Roles) == 0 {
		partial.Roles = make([]int64, 0)
	}

	return partial
}

func PartialToGuildSettingsAutoRolesSettings(guildID int64, guildSettings *GuildSettingsAutoRoles) *database.GuildSettingsAutoroles {
	return &database.GuildSettingsAutoroles{
		GuildID:       guildID,
		ToggleEnabled: guildSettings.ToggleEnabled,
		Roles:         guildSettings.Roles,
	}
}
