package welcomer

import (
	"github.com/WelcomerTeam/Discord/discord"
	jsoniter "github.com/json-iterator/go"
)

type GuildSettingsTimeRolesRole struct {
	Role    discord.Snowflake `json:"role_id"`
	Seconds int               `json:"seconds"`
}

func UnmarshalTimeRolesJSON(rolesJSON []byte) (roles []GuildSettingsTimeRolesRole) {
	_ = jsoniter.Unmarshal(rolesJSON, &roles)

	return
}

func MarshalTimeRolesJSON(roles []GuildSettingsTimeRolesRole) (rolesJSON []byte) {
	rolesJSON, _ = jsoniter.Marshal(roles)

	return
}
