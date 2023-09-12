package backend

import (
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
)

type GuildSettingsFreeRoles struct {
	ToggleEnabled bool    `json:"enabled"`
	Roles         []int64 `json:"roles"`
}

func GuildSettingsFreeRolesSettingsToPartial(
	freeRoles *database.GuildSettingsFreeroles,
) *GuildSettingsFreeRoles {
	partial := &GuildSettingsFreeRoles{
		ToggleEnabled: freeRoles.ToggleEnabled,
		Roles:         freeRoles.Roles,
	}

	if len(partial.Roles) == 0 {
		partial.Roles = make([]int64, 0)
	}

	return partial
}

func PartialToGuildSettingsFreeRolesSettings(guildID int64, guildSettings *GuildSettingsFreeRoles) *database.GuildSettingsFreeroles {
	return &database.GuildSettingsFreeroles{
		GuildID:       guildID,
		ToggleEnabled: guildSettings.ToggleEnabled,
		Roles:         guildSettings.Roles,
	}
}
