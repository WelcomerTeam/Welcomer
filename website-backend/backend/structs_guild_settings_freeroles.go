package backend

import (
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
)

type GuildSettingsFreeRoles struct {
	ToggleEnabled bool     `json:"enabled"`
	Roles         []string `json:"roles"`
}

func GuildSettingsFreeRolesSettingsToPartial(
	freeRoles *database.GuildSettingsFreeroles,
) *GuildSettingsFreeRoles {
	partial := &GuildSettingsFreeRoles{
		ToggleEnabled: freeRoles.ToggleEnabled,
		Roles:         Int64SliceToString(freeRoles.Roles),
	}

	if len(partial.Roles) == 0 {
		partial.Roles = make([]string, 0)
	}

	return partial
}

func PartialToGuildSettingsFreeRolesSettings(guildID int64, guildSettings *GuildSettingsFreeRoles) *database.GuildSettingsFreeroles {
	return &database.GuildSettingsFreeroles{
		GuildID:       guildID,
		ToggleEnabled: guildSettings.ToggleEnabled,
		Roles:         StringSliceToInt64(guildSettings.Roles),
	}
}
