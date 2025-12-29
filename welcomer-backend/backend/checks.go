package backend

import (
	"context"
	"fmt"

	discord "github.com/WelcomerTeam/Discord/discord"
	sandwich_protobuf "github.com/WelcomerTeam/Sandwich-Daemon/proto"
	welcomer "github.com/WelcomerTeam/Welcomer/welcomer-core"
)

func hasWelcomerPresence(ctx context.Context, guildID discord.Snowflake, returnBotGuildMembers bool) (ok bool, guild *discord.Guild, guildMembers []*discord.GuildMember, err error) {
	guildsPb, err := welcomer.SandwichClient.FetchGuild(ctx, &sandwich_protobuf.FetchGuildRequest{
		GuildIds: []int64{int64(guildID)},
	})
	if err != nil {
		welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to get welcomer presence")

		return false, guild, nil, fmt.Errorf("failed to get welcomer presence: %w", err)
	}

	guildPb, ok := guildsPb.GetGuilds()[int64(guildID)]
	if !ok {
		return false, nil, nil, nil
	}

	guild = sandwich_protobuf.PBToGuild(guildPb)
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

func fetchApplicationsForGuild(ctx context.Context, guildID discord.Snowflake) (applicationIdentifiers []string, err error) {
	// Find out what applications can see this guild
	locationsPb, err := welcomer.SandwichClient.WhereIsGuild(ctx, &sandwich_protobuf.WhereIsGuildRequest{
		GuildId: int64(guildID),
	})
	if err != nil {
		welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to do guild lookup")

		return nil, fmt.Errorf("failed to do guild lookup: %w", err)
	}

	locations := locationsPb.GetLocations()
	applicationIdentifiers = make([]string, 0, len(locations))

	for _, location := range locations {
		applicationIdentifiers = append(applicationIdentifiers, location.Identifier)
	}

	return applicationIdentifiers, nil
}
