package backend

import (
	"time"

	discord "github.com/WelcomerTeam/Discord/discord"
	"github.com/WelcomerTeam/Welcomer/welcomer/database"
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

func getGuildMembership(guildID discord.Snowflake) (hasWelcomerPro bool, hasCustomBackgrounds bool, err error) {
	memberships, err := backend.Database.GetValidUserMembershipsByGuildID(backend.ctx, guildID, time.Now())

	if err != nil {
		backend.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to get welcomer memberships")

		return false, false, err
	}

	for _, membership := range memberships {
		switch database.MembershipType(membership.MembershipType) {
		case database.MembershipTypeLegacyCustomBackgrounds,
			database.MembershipTypeCustomBackgrounds:
			hasCustomBackgrounds = true
		case database.MembershipTypeLegacyWelcomerPro1,
			database.MembershipTypeLegacyWelcomerPro3,
			database.MembershipTypeLegacyWelcomerPro5,
			database.MembershipTypeWelcomerPro:
			hasWelcomerPro = true
		}
	}

	return
}
