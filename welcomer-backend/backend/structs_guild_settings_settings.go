package backend

import (
	"database/sql"
	"fmt"

	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
)

type GuildSettingsSettings struct {
	Name             string `json:"name"`
	SiteSplashUrl    string `json:"site_splash_url"`
	EmbedColour      int32  `json:"embed_colour"`
	SiteStaffVisible bool   `json:"site_staff_visible"`
	SiteGuildVisible bool   `json:"site_guild_visible"`
	SiteAllowInvites bool   `json:"site_allow_invites"`
	MemberCount      int32  `json:"member_count"`
	NumberLocale     string `json:"number_locale"`
}

func MustParseNumberLocale(locale string) database.NumberLocale {
	parsed, err := database.ParseNumberLocale(locale)
	if err != nil {
		panic(fmt.Sprintf("failed to parse number locale: %v", err))
	}

	return parsed
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
		MemberCount:      guildSettings.MemberCount,
		NumberLocale:     database.NumberLocale(guildSettings.NumberLocale.Int32).String(),
	}

	return partial
}

func MustParseGuildSettingsSettings(partial *GuildSettingsSettings) error {
	if _, err := database.ParseNumberLocale(partial.NumberLocale); err != nil {
		return fmt.Errorf("invalid number locale: %w", err)
	}

	return nil
}

func PartialToGuildSettings(guildID int64, guildSettings *GuildSettingsSettings) *database.UpdateGuildParams {
	return &database.UpdateGuildParams{
		GuildID:          guildID,
		EmbedColour:      guildSettings.EmbedColour,
		SiteSplashUrl:    guildSettings.SiteSplashUrl,
		SiteStaffVisible: guildSettings.SiteStaffVisible,
		SiteGuildVisible: guildSettings.SiteGuildVisible,
		SiteAllowInvites: guildSettings.SiteAllowInvites,
		NumberLocale: sql.NullInt32{
			Int32: int32(MustParseNumberLocale(guildSettings.NumberLocale)),
			Valid: true,
		},
	}
}
