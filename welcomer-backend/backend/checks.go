package backend

import (
	"context"
	"fmt"
	"time"

	discord "github.com/WelcomerTeam/Discord/discord"
	welcomer "github.com/WelcomerTeam/Welcomer/welcomer-core"
)

func hasWelcomerPresence(ctx context.Context, guildID discord.Snowflake, returnBotGuildMembers bool) (ok bool, guild discord.Guild, guildMembers []discord.GuildMember, err error) {
	guild, err = backend.GRPCInterface.FetchGuildByID(backend.GetBasicEventContext(ctx).ToGRPCContext(), guildID)
	if err != nil {
		welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to get welcomer presence")

		return false, guild, nil, fmt.Errorf("failed to get welcomer presence: %w", err)
	}

	if guild.ID.IsNil() {
		return false, guild, nil, nil
	}

	if returnBotGuildMembers {
		guildMembers, err = fetchBotUsersForGuild(ctx, guildID)
		if err != nil {
			welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to get bot users for guild")
		}
	}

	return true, guild, guildMembers, nil
}

func fetchBotUsersForGuild(ctx context.Context, guildID discord.Snowflake) (guildMembers []discord.GuildMember, err error) {
	// Find out what managers can see this guild
	locations, err := backend.GRPCInterface.WhereIsGuild(backend.GetBasicEventContext(ctx).ToGRPCContext(), guildID)
	if err != nil {
		welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to do guild lookup")

		return nil, fmt.Errorf("failed to do guild lookup: %w", err)
	}

	guildMembers = make([]discord.GuildMember, 0, len(locations))

	for _, location := range locations {
		guildMembers = append(guildMembers, location.GuildMember)
	}

	return guildMembers, nil
}

func fetchManagersForGuild(ctx context.Context, guildID discord.Snowflake) (managers []string, err error) {
	// Find out what managers can see this guild
	locations, err := backend.GRPCInterface.WhereIsGuild(backend.GetBasicEventContext(ctx).ToGRPCContext(), guildID)
	if err != nil {
		welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to do guild lookup")

		return nil, fmt.Errorf("failed to do guild lookup: %w", err)
	}

	managers = make([]string, 0, len(locations))

	for _, location := range locations {
		managers = append(managers, location.Manager)
	}

	return managers, nil
}

func getGuildMembership(ctx context.Context, guildID discord.Snowflake) (hasWelcomerPro, hasCustomBackgrounds bool, err error) {
	memberships, err := backend.Database.GetValidUserMembershipsByGuildID(ctx, guildID, time.Now())
	if err != nil {
		welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to get welcomer memberships")

		return false, false, err
	}

	hasWelcomerPro, hasCustomBackgrounds = welcomer.CheckGuildMemberships(memberships)

	return
}

func CalculateRoleValues(roles []MinimalRole, guildMembers []discord.GuildMember) (convertedRoles []MinimalRole) {
	roleMap := MinimalRolesToMap(roles)

	highestRolePosition := int32(0)

	for _, guildMember := range guildMembers {
		highestRolePositionForMember := getHighestRoleForGuildMember(roleMap, guildMember)
		if highestRolePositionForMember > highestRolePosition {
			highestRolePosition = highestRolePositionForMember
		}
	}

	convertedRoles = make([]MinimalRole, len(roles))

	for i, role := range roles {
		role.IsAssignable = (!role.managed) && (role.Position < highestRolePosition)
		role.IsElevated = false // TODO: Check for permissions

		convertedRoles[i] = role
	}

	return
}

func getHighestRoleForGuildMember(roleMap map[discord.Snowflake]MinimalRole, guildMember discord.GuildMember) int32 {
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
