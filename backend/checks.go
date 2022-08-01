package backend

import (
	"database/sql"

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

func hasWelcomerMembership(guildID discord.Snowflake) (ok bool, err error) {
	var sqlGuildID sql.NullInt64

	sqlGuildID.Int64 = int64(guildID)
	sqlGuildID.Valid = true

	memberships, err := backend.Database.GetUserMembershipsByGuildID(backend.ctx, sqlGuildID)

	if err != nil {
		backend.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to get welcomer memberships")

		return false, nil
	}

	if len(memberships) == 0 {
		return false, nil
	}

	return true, nil
}
