package welcomer

import (
	"context"
	"fmt"
	"strings"

	"github.com/WelcomerTeam/Discord/discord"
	pb "github.com/WelcomerTeam/Sandwich-Daemon/proto"
	sandwich_protobuf "github.com/WelcomerTeam/Sandwich-Daemon/proto"
)

type AssignableRole struct {
	*discord.Role

	IsAssignable bool `json:"is_assignable"`
	IsElevated   bool `json:"is_elevated"`
}

func GetWelcomerPresence(ctx context.Context, guildID discord.Snowflake) (guildMembers []*discord.GuildMember, err error) {
	pbLocations, err := SandwichClient.WhereIsGuild(ctx, &sandwich_protobuf.WhereIsGuildRequest{
		GuildId: int64(guildID),
	})
	if err != nil {
		Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to do guild lookup")

		return nil, err
	}

	guildMembers = make([]*discord.GuildMember, 0, len(pbLocations.GetLocations()))

	mustFetchPermissions := false

	for _, location := range pbLocations.GetLocations() {
		guildMember := sandwich_protobuf.PBToGuildMember(location.GuildMember)
		guildMembers = append(guildMembers, guildMember)
		mustFetchPermissions = mustFetchPermissions || guildMember.Permissions == nil
	}

	if mustFetchPermissions {
		pbRoles, err := SandwichClient.FetchGuildRole(ctx, &sandwich_protobuf.FetchGuildRoleRequest{
			GuildId: int64(guildID),
		})
		if err != nil {
			Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to fetch guild roles")

			return nil, err
		}

		for i, guildMember := range guildMembers {
			permissions := int64(0)

			for _, roleID := range guildMember.Roles {
				for _, userPb := range pbRoles.GetRoles() {
					role := sandwich_protobuf.PBToRole(userPb)
					if role.ID == roleID {
						permissions |= int64(userPb.Permissions)
					}
				}
			}

			guildMembers[i].Permissions = ToPointer(discord.Int64(permissions))
		}
	}

	return guildMembers, err
}

func Accelerator_CanAssignRole(ctx context.Context, guildID discord.Snowflake, role *discord.Role) (canAssignRoles, isRoleAssignable, isRoleElevated bool, err error) {
	// Fetch guild roles so we can check if the role is assignable.
	guildRolesPb, err := SandwichClient.FetchGuildRole(ctx, &sandwich_protobuf.FetchGuildRoleRequest{
		GuildId: int64(guildID),
	})
	if err != nil {
		Logger.Error().Err(err).
			Int64("guild_id", int64(guildID)).
			Msg("Failed to fetch guild roles")

		return false, false, false, err
	}

	guildRoles := make([]*discord.Role, 0, len(guildRolesPb.GetRoles()))

	for _, rolePb := range guildRolesPb.GetRoles() {
		guildRoles = append(guildRoles, sandwich_protobuf.PBToRole(rolePb))
	}

	// Check welcomer presence on the current server.
	welcomerPresence, err := GetWelcomerPresence(ctx, guildID)
	if err != nil {
		Logger.Error().Err(err).
			Int64("guild_id", int64(guildID)).
			Msg("Failed to get welcomer presence")

		return false, false, false, err
	}

	// Check if welcomer can assign roles to users.
	for _, guildMember := range welcomerPresence {
		canAssignRoles = canAssignRoles || GuildMemberCanAssignRoles(guildMember)
	}

	if !canAssignRoles {
		return canAssignRoles, false, false, nil
	}

	// Check if the role is assignable by welcomer using the guild roles and roles Welcomer has.
	isRoleAssignable, isRoleElevated = CanAssignRole(role, guildRoles, welcomerPresence)
	if !isRoleAssignable {
		return canAssignRoles, false, false, nil
	}

	return canAssignRoles, isRoleAssignable, isRoleElevated, nil
}

func CanAssignRole(role *discord.Role, guildRoles []*discord.Role, guildMembers []*discord.GuildMember) (isAssignable, isElevated bool) {
	if role.Managed {
		return false, false
	}

	assignableRoles := CalculateRoleValues(guildRoles, guildMembers)

	for _, assignableRole := range assignableRoles {
		if assignableRole.ID == role.ID {
			return assignableRole.IsAssignable, assignableRole.IsElevated
		}
	}

	return false, false
}

func GuildMemberCanAssignRoles(guildMember *discord.GuildMember) bool {
	if guildMember.Permissions != nil {
		permission := *guildMember.Permissions

		return (permission & discord.PermissionManageRoles) != 0
	}

	return false
}

func CalculateRoleValues(roles []*discord.Role, guildMembers []*discord.GuildMember) (convertedRoles []*AssignableRole) {
	roleMap := map[discord.Snowflake]AssignableRole{}

	for _, role := range roles {
		roleMap[role.ID] = AssignableRole{role, false, false}
	}

	convertedRoles = make([]*AssignableRole, len(roles))

	roleIndex := 0

	for _, role := range roles {
		convertedRoles[roleIndex] = &AssignableRole{role, false, false}
		roleIndex++
	}

	for _, guildMember := range guildMembers {
		highestRolePositionForMember := GetHighestRoleForGuildMember(roleMap, guildMember)

		for i, role := range convertedRoles {
			convertedRoles[i].IsAssignable = role.IsAssignable || ((!role.Managed) &&
				(role.Position < highestRolePositionForMember) &&
				GuildMemberCanAssignRoles(guildMember))

			convertedRoles[i].IsElevated = (role.Permissions & PermissionElevated) != 0
		}
	}

	return convertedRoles
}

func GetHighestRoleForGuildMember(roleMap map[discord.Snowflake]AssignableRole, guildMember *discord.GuildMember) int32 {
	highestRolePosition := int32(0)

	for _, roleID := range guildMember.Roles {
		role, ok := roleMap[roleID]
		if ok {
			if role.Position > highestRolePosition {
				highestRolePosition = role.Position
			}
		}
	}

	return highestRolePosition
}

func FilterAssignableRolesAsSnowflakes(ctx context.Context, sandwichClient pb.SandwichClient, guildID, applicationID int64, roleIDs []int64) (out []discord.Snowflake, err error) {
	assignableRoles, err := FilterAssignableRoles(ctx, sandwichClient, guildID, applicationID, roleIDs)
	if err != nil {
		return nil, err
	}

	out = make([]discord.Snowflake, len(assignableRoles))
	for i, role := range assignableRoles {
		out[i] = discord.Snowflake(role.ID)
	}

	return out, nil
}

func FilterAssignableRoles(ctx context.Context, sandwichClient pb.SandwichClient, guildID, applicationID int64, roleIDs []int64) (out []discord.Role, err error) {
	guildRolesPb, err := sandwichClient.FetchGuildRole(ctx, &pb.FetchGuildRoleRequest{
		GuildId: int64(guildID),
	})
	if err != nil {
		Logger.Error().Err(err).
			Int64("guild_id", int64(guildID)).
			Msg("Failed to fetch guild roles")

		return nil, fmt.Errorf("failed to fetch guild roles: %v", err)
	}

	guildMembersPb, err := sandwichClient.FetchGuildMember(ctx, &pb.FetchGuildMemberRequest{
		GuildId: int64(guildID),
		UserIds: []int64{int64(applicationID)},
	})
	if err != nil || guildMembersPb == nil {
		Logger.Error().Err(err).
			Int64("guild_id", int64(guildID)).
			Int64("user_id", int64(applicationID)).
			Msg("Failed to fetch application guild member")

		return nil, ErrMissingApplicationUser
	}

	guildRoles := guildRolesPb.GetRoles()
	guildMembers := guildMembersPb.GetGuildMembers()

	// Get the guild member of the application.
	applicationUser, ok := guildMembers[int64(applicationID)]
	if !ok || applicationUser == nil {
		Logger.Error().Err(err).
			Int64("guild_id", int64(guildID)).
			Int64("user_id", int64(applicationID)).
			Msg("Application guild member not present in response")

		return nil, ErrMissingApplicationUser
	}

	// Get the top role position of the application user.
	var applicationUserTopRolePosition int32

	for _, roleID := range applicationUser.GetRoles() {
		role, ok := guildRoles[roleID]
		if ok && role.GetPosition() > applicationUserTopRolePosition {
			applicationUserTopRolePosition = role.GetPosition()
		}
	}

	// Filter out any roles that are not in cache or are above the application user's top role position.
	for _, roleID := range roleIDs {
		rolePb, ok := guildRoles[roleID]
		if ok {
			if rolePb.GetPosition() < applicationUserTopRolePosition {
				out = append(out, *sandwich_protobuf.PBToRole(rolePb))
			}
		}
	}

	return out, nil
}

var nameMap = map[int]string{
	discord.PermissionKickMembers:     "Kick Members",
	discord.PermissionBanMembers:      "Ban Members",
	discord.PermissionAdministrator:   "Administrator",
	discord.PermissionManageChannels:  "Manage Channels",
	discord.PermissionManageServer:    "Manage Server",
	discord.PermissionManageMessages:  "Manage Messages",
	discord.PermissionManageRoles:     "Manage Roles",
	discord.PermissionManageWebhooks:  "Manage Webhooks",
	discord.PermissionManageEmojis:    "Manage Emojis",
	discord.PermissionManageThreads:   "Manage Threads",
	discord.PermissionModerateMembers: "Moderate Members",
}

func GetRolePermissionList(permissions int) []string {
	roleNames := make([]string, 0)

	for permission, name := range nameMap {
		if permissions&permission != 0 {
			roleNames = append(roleNames, name)
		}
	}

	return roleNames
}

func GetRolePermissionListAsString(permissions int) string {
	roleNames := GetRolePermissionList(permissions)

	if len(roleNames) == 0 {
		return "None"
	}

	var builder strings.Builder

	for i, name := range roleNames {
		if i > 0 {
			builder.WriteString(", ")
		}

		builder.WriteString("`" + name + "`")
	}

	return builder.String()
}
