package backend

import (
	"time"

	discord "github.com/WelcomerTeam/Discord/discord"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
)

func hasWelcomerPresence(guildID discord.Snowflake, returnBotGuildMembers bool) (ok bool, guild *discord.Guild, guildMembers []*discord.GuildMember, err error) {
	guild, err = backend.GRPCInterface.FetchGuildByID(backend.GetBasicEventContext().ToGRPCContext(), guildID)

	if err != nil {
		backend.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to get welcomer presence")

		return false, nil, nil, err
	}

	if guild == nil {
		return false, nil, nil, nil
	}

	if returnBotGuildMembers {
		guildMembers, err = fetchBotUsersForGuild(guildID)

		if err != nil {
			backend.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to get bot users for guild")
		}
	}

	return true, guild, guildMembers, nil
}

func fetchBotUsersForGuild(guildID discord.Snowflake) (guildMembers []*discord.GuildMember, err error) {
	// Find out what managers can see this guild
	locations, err := backend.GRPCInterface.WhereIsGuild(backend.GetBasicEventContext().ToGRPCContext(), guildID)

	if err != nil {
		backend.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to do guild lookup")

		return nil, err
	}

	guildMembers = make([]*discord.GuildMember, 0, len(locations))

	for _, location := range locations {
		if location.GuildMember != nil {
			guildMembers = append(guildMembers, location.GuildMember)
		}
	}

	return guildMembers, nil
}

func getGuildMembership(guildID discord.Snowflake) (hasWelcomerPro bool, hasCustomBackgrounds bool, err error) {
	memberships, err := backend.Database.GetValidUserMembershipsByGuildID(backend.ctx, guildID, time.Now())

	if err != nil {
		backend.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to get welcomer memberships")

		return false, false, err
	}

	hasWelcomerPro, hasCustomBackgrounds = welcomer.CheckGuildMemberships(memberships)

	return
}

func CalculateRoleValues(roles []*MinimalRole, guildMembers []*discord.GuildMember) (convertedRoles []*MinimalRole) {
	roleMap := MinimalRolesToMap(roles)

	highestRolePosition := int32(0)

	for _, guildMember := range guildMembers {
		highestRolePositionForMember := getHighestRoleForGuildMember(roleMap, guildMember)
		if highestRolePositionForMember > highestRolePosition {
			highestRolePosition = highestRolePositionForMember
		}
	}

	convertedRoles = make([]*MinimalRole, len(roles))

	for i, role := range roles {
		role.IsAssignable = (!role.managed) && (role.Position < highestRolePosition)
		role.IsElevated = false // TODO: Check for permissions

		convertedRoles[i] = role
	}

	return
}

func getHighestRoleForGuildMember(roleMap map[discord.Snowflake]*MinimalRole, guildMember *discord.GuildMember) int32 {
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
