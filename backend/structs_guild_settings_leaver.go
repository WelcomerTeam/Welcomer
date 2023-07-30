package backend

import (
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
)

type GuildSettingsLeaver struct {
	GuildID       int64  `json:"guild_id"`
	ToggleEnabled bool   `json:"toggle_enabled"`
	Channel       int64  `json:"channel"`
	MessageFormat string `json:"message_format"`
}

func GuildSettingsLeaverSettingsToPartial(
	leaver *database.GuildSettingsLeaver,
) *GuildSettingsLeaver {
	partial := &GuildSettingsLeaver{
		ToggleEnabled: leaver.ToggleEnabled,
		Channel:       leaver.Channel,
		MessageFormat: JSONBToString(leaver.MessageFormat),
	}

	return partial
}

func PartialToGuildSettingsLeaverSettings(guildID int64, guildSettings *GuildSettingsLeaver) *database.GuildSettingsLeaver {
	return &database.GuildSettingsLeaver{
		GuildID:       guildID,
		ToggleEnabled: guildSettings.ToggleEnabled,
		Channel:       guildSettings.Channel,
		MessageFormat: StringToJSONB(guildSettings.MessageFormat),
	}
}
