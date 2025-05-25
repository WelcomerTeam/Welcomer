package welcomer

import (
	"context"
	"encoding/json"
	"slices"

	"github.com/WelcomerTeam/Discord/discord"
	pb "github.com/WelcomerTeam/Sandwich-Daemon/proto"
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

func FilterAssignableTimeRoles(ctx context.Context, sandwichClient pb.SandwichClient, guildID, applicationID int64, timeRoles []GuildSettingsTimeRolesRole) (out []GuildSettingsTimeRolesRole, err error) {
	roleIDs := make([]int64, len(timeRoles))
	for i, timeRole := range timeRoles {
		roleIDs[i] = int64(timeRole.Role)
	}

	assignableRoleIDs, err := FilterAssignableRolesAsSnowflakes(ctx, sandwichClient, guildID, applicationID, roleIDs)
	if err != nil {
		return nil, err
	}

	for _, timeRole := range timeRoles {
		if slices.Contains(assignableRoleIDs, timeRole.Role) {
			out = append(out, timeRole)

			continue
		}
	}

	return out, nil
}
