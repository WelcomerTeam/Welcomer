package welcomer

import (
	"context"

	"github.com/WelcomerTeam/Discord/discord"
	sandwich "github.com/WelcomerTeam/Sandwich-Daemon/protobuf"
	internal "github.com/WelcomerTeam/Sandwich/sandwich"
)

type AssignableRole struct {
	discord.Role

	IsAssignable bool `json:"is_assignable"`
	IsElevated   bool `json:"is_elevated"`
}

func GetWelcomerPresence(ctx context.Context, guildID discord.Snowflake) (guildMembers []discord.GuildMember, err error) {
	locations, err := GRPCInterface.WhereIsGuild(&internal.GRPCContext{
		Context:        ctx,
		Logger:         Logger,
		SandwichClient: SandwichClient,
		GRPCInterface:  GRPCInterface,
	}, guildID)
	if err != nil {
		Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to do guild lookup")

		return nil, err
	}

	guildMembers = make([]discord.GuildMember, 0, len(locations))

	mustFetchPermissions := false

	for _, location := range locations {
		guildMembers = append(guildMembers, location.GuildMember)

		mustFetchPermissions = mustFetchPermissions || location.GuildMember.Permissions == nil
	}

	if mustFetchPermissions {
		roles, err := GRPCInterface.FetchRolesByName(&internal.GRPCContext{
			Context:        ctx,
			Logger:         Logger,
			SandwichClient: SandwichClient,
			GRPCInterface:  GRPCInterface,
		}, guildID, "")
		if err != nil {
			Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to fetch guild roles")

			return nil, err
		}

		for i, guildMember := range guildMembers {
			permissions := int64(0)

			for _, roleID := range guildMember.Roles {
				for _, role := range roles {
					if role.ID == roleID {
						permissions |= int64(role.Permissions)
					}
				}
			}

			guildMembers[i].Permissions = ToPointer(discord.Int64(permissions))
		}
	}

	return guildMembers, err
}

func MinimalRolesToMap(roles []discord.Role) map[discord.Snowflake]AssignableRole {
	roleMap := map[discord.Snowflake]AssignableRole{}

	for _, role := range roles {
		roleMap[role.ID] = AssignableRole{role, false, false}
	}

	return roleMap
}

func Accelerator_CanAssignRole(ctx context.Context, guildID discord.Snowflake, role discord.Role) (canAssignRoles, isRoleAssignable bool, err error) {

	// Fetch guild roles so we can check if the role is assignable.
	guildRolesPb, err := SandwichClient.FetchGuildRoles(ctx, &sandwich.FetchGuildRolesRequest{
		GuildID: int64(guildID),
	})
	if err != nil {
		Logger.Error().Err(err).
			Int64("guild_id", int64(guildID)).
			Msg("Failed to fetch guild roles")

		return false, false, err
	}

	guildRoles := make([]discord.Role, 0, len(guildRolesPb.GetGuildRoles()))

	for _, rolePb := range guildRolesPb.GetGuildRoles() {
		role, err := sandwich.GRPCToRole(rolePb)
		if err != nil {
			continue
		}

		guildRoles = append(guildRoles, role)
	}

	// Check welcomer presence on the current server.
	welcomerPresence, err := GetWelcomerPresence(ctx, guildID)
	if err != nil {
		Logger.Error().Err(err).
			Int64("guild_id", int64(guildID)).
			Msg("Failed to get welcomer presence")

		return false, false, err
	}

	// Check if welcomer can assign roles to users.
	for _, guildMember := range welcomerPresence {
		canAssignRoles = canAssignRoles || GuildMemberCanAssignRoles(guildMember)
	}

	if !canAssignRoles {
		return canAssignRoles, false, nil
	}

	// Check if the role is assignable by welcomer using the guild roles and roles Welcomer has.
	isRoleAssignable = CanAssignRole(role, guildRoles, welcomerPresence)
	if !isRoleAssignable {
		return canAssignRoles, false, nil
	}

	return true, true, nil
}

func CanAssignRole(role discord.Role, guildRoles []discord.Role, guildMembers []discord.GuildMember) bool {
	if role.Managed {
		return false
	}

	assignableRoles := CalculateRoleValues(guildRoles, guildMembers)

	for _, assignableRole := range assignableRoles {
		if assignableRole.ID == role.ID {
			return assignableRole.IsAssignable
		}
	}

	return false
}

func GuildMemberCanAssignRoles(guildMember discord.GuildMember) bool {
	if guildMember.Permissions != nil {
		permission := *guildMember.Permissions

		return (permission & discord.PermissionManageRoles) != 0
	}

	return false
}

func CalculateRoleValues(roles []discord.Role, guildMembers []discord.GuildMember) (convertedRoles []AssignableRole) {
	roleMap := MinimalRolesToMap(roles)

	convertedRoles = make([]AssignableRole, len(roles))

	roleIndex := 0

	for _, role := range roles {
		convertedRoles[roleIndex] = AssignableRole{role, false, false}
		roleIndex++
	}

	for _, guildMember := range guildMembers {
		highestRolePositionForMember := GetHighestRoleForGuildMember(roleMap, guildMember)

		for i, role := range convertedRoles {
			convertedRoles[i].IsAssignable = role.IsAssignable || ((!role.Managed) &&
				(role.Position < highestRolePositionForMember) &&
				GuildMemberCanAssignRoles(guildMember))

			convertedRoles[i].IsElevated = (role.Permissions & discord.PermissionElevated) != 0
		}
	}

	return convertedRoles
}

func GetHighestRoleForGuildMember(roleMap map[discord.Snowflake]AssignableRole, guildMember discord.GuildMember) int32 {
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
