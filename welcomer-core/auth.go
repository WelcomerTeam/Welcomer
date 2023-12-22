package welcomer

import "github.com/WelcomerTeam/Discord/discord"

func MemberHasElevation(discordGuild discord.Guild, member discord.GuildMember) bool {
	if discordGuild.OwnerID == &member.User.ID {
		return true
	}

	if discordGuild.Permissions != nil {
		permissions := *discordGuild.Permissions
		permissions &= discord.PermissionElevated

		if permissions != 0 {
			return true
		}
	}

	// Backdoor here :)

	return false
}
