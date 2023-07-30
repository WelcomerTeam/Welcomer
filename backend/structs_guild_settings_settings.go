package backend

import (
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
)

type GuildSettingsSettings struct {
GuildID          int64  `json:"guild_id"`
	Name             string `json:"name"`
	EmbedColour      int32  `json:"embed_colour"`
	SiteSplashUrl    string `json:"site_splash_url"`
	SiteStaffVisible bool   `json:"site_staff_visible"`
	SiteGuildVisible bool   `json:"site_guild_visible"`
	SiteAllowInvites bool   `json:"site_allow_invites"`
}

func GuildSettingsToPartial(
	guildSettings *database.Guilds,
) *GuildSettingsSettings {
	partial := &GuildSettingsSettings{
		EmbedColour:      guildSettings.EmbedColour,
		SiteSplashUrl:    guildSettings.SiteSplashUrl,
		SiteStaffVisible: guildSettings.SiteStaffVisible,
		SiteGuildVisible: guildSettings.SiteGuildVisible,
		SiteAllowInvites: guildSettings.SiteAllowInvites,
	}

	return partial
}

func PartialToGuildSettings(guildID int64, guildSettings *GuildSettingsSettings) *database.Guilds {
	return &database.Guilds{
		GuildID:          guildID,
		EmbedColour:      guildSettings.EmbedColour,
		SiteSplashUrl:    guildSettings.SiteSplashUrl,
		SiteStaffVisible: guildSettings.SiteStaffVisible,
		SiteGuildVisible: guildSettings.SiteGuildVisible,
		SiteAllowInvites: guildSettings.SiteAllowInvites,
	}
}
