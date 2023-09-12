package backend

import (
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
)

type GuildSettingsLeaver struct {
	ToggleEnabled bool   `json:"enabled"`
	Channel       int64  `json:"channel"`
	MessageFormat string `json:"message_json"`
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
