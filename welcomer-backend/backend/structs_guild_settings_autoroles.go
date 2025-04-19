package backend

import (
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
)

type GuildSettingsAutoRoles struct {
	Roles         []string `json:"roles"`
	ToggleEnabled bool     `json:"enabled"`
}

func GuildSettingsAutoRolesSettingsToPartial(
	autoRoles *database.GuildSettingsAutoroles,
) *GuildSettingsAutoRoles {
	partial := &GuildSettingsAutoRoles{
		ToggleEnabled: autoRoles.ToggleEnabled,
		Roles:         welcomer.Int64SliceToString(autoRoles.Roles),
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
		Roles:         welcomer.StringSliceToInt64(guildSettings.Roles),
	}
}
