package backend

import (
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
)

type GuildSettingsLeaver struct {
	ToggleEnabled bool   `json:"enabled"`
	Channel       *string  `json:"channel"`
	MessageFormat string `json:"message_json"`
}

func GuildSettingsLeaverSettingsToPartial(
	leaver *database.GuildSettingsLeaver,
) *GuildSettingsLeaver {
	partial := &GuildSettingsLeaver{
		ToggleEnabled: leaver.ToggleEnabled,
		Channel:       Int64ToStringPointer(leaver.Channel),
		MessageFormat: JSONBToString(leaver.MessageFormat),
	}

	return partial
}

func PartialToGuildSettingsLeaverSettings(guildID int64, guildSettings *GuildSettingsLeaver) *database.GuildSettingsLeaver {
	return &database.GuildSettingsLeaver{
		GuildID:       guildID,
		ToggleEnabled: guildSettings.ToggleEnabled,
		Channel:       StringPointerToInt64(guildSettings.Channel),
		MessageFormat: StringToJSONB(guildSettings.MessageFormat),
	}
}
