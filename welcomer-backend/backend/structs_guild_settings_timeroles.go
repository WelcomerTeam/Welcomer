package backend

import (
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
)

type GuildSettingsTimeRoles struct {
	ToggleEnabled bool                                  `json:"enabled"`
	Roles         []welcomer.GuildSettingsTimeRolesRole `json:"roles"`
}

func GuildSettingsTimeRolesSettingsToPartial(
	timeroles *database.GuildSettingsTimeroles,
) *GuildSettingsTimeRoles {
	partial := &GuildSettingsTimeRoles{
		ToggleEnabled: timeroles.ToggleEnabled,
		Roles:         welcomer.UnmarshalTimeRolesJSON(JSONBToBytes(timeroles.Timeroles)),
	}

	if len(partial.Roles) == 0 {
		partial.Roles = make([]welcomer.GuildSettingsTimeRolesRole, 0)
	}

	return partial
}

func PartialToGuildSettingsTimeRolesSettings(guildID int64, guildSettings *GuildSettingsTimeRoles) *database.GuildSettingsTimeroles {
	return &database.GuildSettingsTimeroles{
		GuildID:       guildID,
		ToggleEnabled: guildSettings.ToggleEnabled,
		Timeroles:     BytesToJSONB(welcomer.MarshalTimeRolesJSON(guildSettings.Roles)),
	}
}
