package backend

import (
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
)

type GuildSettingsFreeRoles struct {
	Roles         []string `json:"roles"`
	ToggleEnabled bool     `json:"enabled"`
}

func GuildSettingsFreeRolesSettingsToPartial(
	freeRoles *database.GuildSettingsFreeroles,
) *GuildSettingsFreeRoles {
	partial := &GuildSettingsFreeRoles{
		ToggleEnabled: freeRoles.ToggleEnabled,
		Roles:         welcomer.Int64SliceToString(freeRoles.Roles),
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
		Roles:         welcomer.StringSliceToInt64(guildSettings.Roles),
	}
}
