package backend

import (
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
)

type GuildSettingsBorderwall struct {
	ToggleEnabled   bool    `json:"toggle_enabled"`
	MessageVerify   string  `json:"message_verify"`
	MessageVerified string  `json:"message_verified"`
	RolesOnJoin     []int64 `json:"roles_on_join"`
	RolesOnVerify   []int64 `json:"roles_on_verify"`
}

func GuildSettingsBorderwallSettingsToPartial(
	borderwall *database.GuildSettingsBorderwall,
) *GuildSettingsBorderwall {
	partial := &GuildSettingsBorderwall{
		ToggleEnabled:   borderwall.ToggleEnabled,
		MessageVerify:   JSONBToString(borderwall.MessageVerify),
		MessageVerified: JSONBToString(borderwall.MessageVerified),
		RolesOnJoin:     borderwall.RolesOnJoin,
		RolesOnVerify:   borderwall.RolesOnVerify,
	}

	if len(partial.RolesOnJoin) == 0 {
		partial.RolesOnJoin = make([]int64, 0)
	}

	if len(partial.RolesOnVerify) == 0 {
		partial.RolesOnVerify = make([]int64, 0)
	}

	return partial
}

func PartialToGuildSettingsBorderwallSettings(guildID int64, guildSettings *GuildSettingsBorderwall) *database.GuildSettingsBorderwall {
	return &database.GuildSettingsBorderwall{
		GuildID:         guildID,
		ToggleEnabled:   guildSettings.ToggleEnabled,
		MessageVerify:   StringToJSONB(guildSettings.MessageVerify),
		MessageVerified: StringToJSONB(guildSettings.MessageVerified),
		RolesOnJoin:     guildSettings.RolesOnJoin,
		RolesOnVerify:   guildSettings.RolesOnVerify,
	}
}
