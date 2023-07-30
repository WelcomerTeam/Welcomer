package backend

import (
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	jsoniter "github.com/json-iterator/go"
)

type GuildSettingsTimeRoles struct {
	ToggleEnabled bool                         `json:"toggle_enabled"`
	Roles         []GuildSettingsTimeRolesRole `json:"roles"`
}

type GuildSettingsTimeRolesRole struct {
	Role    string `json:"role_id"`
	Seconds int    `json:"seconds"`
}

func GuildSettingsTimeRolesSettingsToPartial(
	timeroles *database.GuildSettingsTimeroles,
) *GuildSettingsTimeRoles {
	partial := &GuildSettingsTimeRoles{
		ToggleEnabled: timeroles.ToggleEnabled,
		Roles:         UnmarshalTimeRolesJSON(JSONBToBytes(timeroles.Timeroles)),
	}

	if len(partial.Roles) == 0 {
		partial.Roles = make([]GuildSettingsTimeRolesRole, 0)
	}

	return partial
}

func UnmarshalTimeRolesJSON(rolesJSON []byte) (roles []GuildSettingsTimeRolesRole) {
	_ = jsoniter.Unmarshal(rolesJSON, &roles)

	return
}

func MarshalTimeRolesJSON(roles []GuildSettingsTimeRolesRole) (rolesJSON []byte) {
	rolesJSON, _ = jsoniter.Marshal(roles)

	return
}

func PartialToGuildSettingsTimeRolesSettings(guildID int64, guildSettings *GuildSettingsTimeRoles) *database.GuildSettingsTimeroles {
	return &database.GuildSettingsTimeroles{
		GuildID:       guildID,
		ToggleEnabled: guildSettings.ToggleEnabled,
		Timeroles:     BytesToJSONB(MarshalTimeRolesJSON(guildSettings.Roles)),
	}
}
