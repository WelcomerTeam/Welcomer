package backend

import (
	"context"
	"fmt"
	"time"

	discord "github.com/WelcomerTeam/Discord/discord"
	welcomer "github.com/WelcomerTeam/Welcomer/welcomer-core"
)

func hasWelcomerPresence(ctx context.Context, guildID discord.Snowflake, returnBotGuildMembers bool) (ok bool, guild discord.Guild, guildMembers []discord.GuildMember, err error) {
	guild, err = welcomer.GRPCInterface.FetchGuildByID(backend.GetBasicEventContext(ctx).ToGRPCContext(), guildID)
	if err != nil {
		welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to get welcomer presence")

		return false, guild, nil, fmt.Errorf("failed to get welcomer presence: %w", err)
	}

	if guild.ID.IsNil() {
		return false, guild, nil, nil
	}

	if returnBotGuildMembers {
		guildMembers, err = welcomer.GetWelcomerPresence(ctx, guildID)
		if err != nil {
			welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to get bot users for guild")
		}
	}

	return true, guild, guildMembers, nil
}

func fetchManagersForGuild(ctx context.Context, guildID discord.Snowflake) (managers []string, err error) {
	// Find out what managers can see this guild
	locations, err := welcomer.GRPCInterface.WhereIsGuild(backend.GetBasicEventContext(ctx).ToGRPCContext(), guildID)
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
	memberships, err := welcomer.Queries.GetValidUserMembershipsByGuildID(ctx, guildID, time.Now())
	if err != nil {
		welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to get welcomer memberships")

		return false, false, err
	}

	hasWelcomerPro, hasCustomBackgrounds = welcomer.CheckGuildMemberships(memberships)

	return
}
