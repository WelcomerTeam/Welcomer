package backend

import (
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
)

type GuildSettingsBorderwall struct {
	ToggleEnabled   bool    `json:"toggle_enabled"`
	ToggleSendDm    bool    `json:"toggle_send_dm"`
	Channel         *string `json:"channel"`
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
		ToggleSendDm:    borderwall.ToggleSendDm,
		Channel:         Int64ToStringPointer(borderwall.Channel),
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
		ToggleSendDm:    guildSettings.ToggleSendDm,
		Channel:         StringPointerToInt64(guildSettings.Channel),
		MessageVerify:   StringToJSONB(guildSettings.MessageVerify),
		MessageVerified: StringToJSONB(guildSettings.MessageVerified),
		RolesOnJoin:     guildSettings.RolesOnJoin,
		RolesOnVerify:   guildSettings.RolesOnVerify,
	}
}
