package backend

import (
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
)

type GuildSettingsLeaver struct {
	Channel       *string `json:"channel"`
	MessageFormat string  `json:"message_json"`
	ToggleEnabled bool    `json:"enabled"`
}

func GuildSettingsLeaverSettingsToPartial(leaver database.GuildSettingsLeaver) *GuildSettingsLeaver {
	partial := &GuildSettingsLeaver{
		ToggleEnabled: leaver.ToggleEnabled,
		Channel:       welcomer.Int64ToStringPointer(leaver.Channel),
		MessageFormat: welcomer.JSONBToString(leaver.MessageFormat),
	}

	return partial
}

func PartialToGuildSettingsLeaverSettings(guildID int64, guildSettings *GuildSettingsLeaver) *database.GuildSettingsLeaver {
	return &database.GuildSettingsLeaver{
		GuildID:       guildID,
		ToggleEnabled: guildSettings.ToggleEnabled,
		Channel:       welcomer.StringPointerToInt64(guildSettings.Channel),
		MessageFormat: welcomer.StringToJSONB(guildSettings.MessageFormat),
	}
}
