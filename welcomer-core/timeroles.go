package welcomer

import (
	"encoding/json"

	"github.com/WelcomerTeam/Discord/discord"
)

type GuildSettingsTimeRolesRole struct {
	Role    discord.Snowflake `json:"role_id"`
	Seconds int               `json:"seconds"`
}

func UnmarshalTimeRolesJSON(rolesJSON []byte) (roles []GuildSettingsTimeRolesRole) {
	_ = json.Unmarshal(rolesJSON, &roles)

	return
}

func MarshalTimeRolesJSON(roles []GuildSettingsTimeRolesRole) (rolesJSON []byte) {
	rolesJSON, _ = json.Marshal(roles)

	return
}
