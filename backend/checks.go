package backend

import (
	"time"

	discord "github.com/WelcomerTeam/Discord/discord"
)

func hasWelcomerPresence(guildID discord.Snowflake) (ok bool, guild *discord.Guild, err error) {
	guild, err = backend.GRPCInterface.FetchGuildByID(backend.GetBasicEventContext(), guildID)

	if err != nil {
		backend.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to get welcomer presence")

		return false, nil, nil
	}

	if guild == nil {
		return false, nil, nil
	}

	return true, guild, nil
}

func hasWelcomerMembership(guildID discord.Snowflake) (bool, error) {
	memberships, err := backend.Database.GetValidUserMembershipsByGuildID(backend.ctx, guildID, time.Now())

	if err != nil {
		backend.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to get welcomer memberships")

		return false, nil
	}

	return len(memberships) > 0, nil
}
