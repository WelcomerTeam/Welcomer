package welcomer

import (
	"context"
	"encoding/json"

	"github.com/WelcomerTeam/Discord/discord"
	pb "github.com/WelcomerTeam/Sandwich-Daemon/protobuf"
	utils "github.com/WelcomerTeam/Welcomer/welcomer-utils"
	"github.com/rs/zerolog"
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

func FilterAssignableRoles(ctx context.Context, sandwichClient pb.SandwichClient, logger zerolog.Logger, guildID int64, applicationID int64, roleIDs []int64) (out []discord.Snowflake, err error) {
	guildRoles, err := sandwichClient.FetchGuildRoles(ctx, &pb.FetchGuildRolesRequest{
		GuildID: int64(guildID),
	})
	if err != nil {
		logger.Error().Err(err).
			Int64("guild_id", int64(guildID)).
			Msg("Failed to fetch guild roles.")

		return nil, err
	}

	guildMember, err := sandwichClient.FetchGuildMembers(ctx, &pb.FetchGuildMembersRequest{
		GuildID: int64(guildID),
		UserIDs: []int64{int64(applicationID)},
	})
	if err != nil {
		logger.Error().Err(err).
			Int64("guild_id", int64(guildID)).
			Int64("user_id", int64(applicationID)).
			Msg("Failed to fetch application guild member.")
	}

	// Get the guild member of the application.
	applicationUser, ok := guildMember.GuildMembers[int64(applicationID)]
	if !ok {
		logger.Error().Err(err).
			Int64("guild_id", int64(guildID)).
			Int64("user_id", int64(applicationID)).
			Msg("Application guild member not present in response.")

		return nil, utils.ErrMissingApplicationUser
	}

	// Get the top role position of the application user.
	var applicationUserTopRolePosition int32

	for _, roleID := range applicationUser.Roles {
		role, ok := guildRoles.GuildRoles[roleID]
		if ok && role.Position > applicationUserTopRolePosition {
			applicationUserTopRolePosition = role.Position
		}
	}

	// Filter out any roles that are not in cache or are above the application user's top role position.
	for _, roleID := range roleIDs {
		role, ok := guildRoles.GuildRoles[roleID]
		if ok {
			if role.Position < applicationUserTopRolePosition {
				out = append(out, discord.Snowflake(roleID))
			}
		}
	}

	return out, nil
}

func FilterAssignableTimeRoles(ctx context.Context, sandwichClient pb.SandwichClient, logger zerolog.Logger, guildID int64, applicationID int64, timeRoles []GuildSettingsTimeRolesRole) (out []GuildSettingsTimeRolesRole, err error) {
	roleIDs := make([]int64, len(timeRoles))
	for i, timeRole := range timeRoles {
		roleIDs[i] = int64(timeRole.Role)
	}

	assignableRoleIDs, err := FilterAssignableRoles(ctx, sandwichClient, logger, guildID, applicationID, roleIDs)
	if err != nil {
		return nil, err
	}

	for _, timeRole := range timeRoles {
		for _, assignableRoleID := range assignableRoleIDs {
			if timeRole.Role == assignableRoleID {
				out = append(out, timeRole)

				break
			}
		}
	}

	return out, nil
}
