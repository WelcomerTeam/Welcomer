package backend

import (
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
)

type GuildSettingsBorderwall struct {
	Channel         *string  `json:"channel"`
	MessageVerify   string   `json:"message_verify"`
	MessageVerified string   `json:"message_verified"`
	RolesOnJoin     []string `json:"roles_on_join"`
	RolesOnVerify   []string `json:"roles_on_verify"`
	ToggleEnabled   bool     `json:"enabled"`
	ToggleSendDm    bool     `json:"send_dm"`
}

func GuildSettingsBorderwallSettingsToPartial(borderwall database.GuildSettingsBorderwall) *GuildSettingsBorderwall {
	partial := &GuildSettingsBorderwall{
		ToggleEnabled:   borderwall.ToggleEnabled,
		ToggleSendDm:    borderwall.ToggleSendDm,
		Channel:         welcomer.Int64ToStringPointer(borderwall.Channel),
		MessageVerify:   welcomer.JSONBToString(borderwall.MessageVerify),
		MessageVerified: welcomer.JSONBToString(borderwall.MessageVerified),
		RolesOnJoin:     welcomer.Int64SliceToString(borderwall.RolesOnJoin),
		RolesOnVerify:   welcomer.Int64SliceToString(borderwall.RolesOnVerify),
	}

	if len(partial.RolesOnJoin) == 0 {
		partial.RolesOnJoin = make([]string, 0)
	}

	if len(partial.RolesOnVerify) == 0 {
		partial.RolesOnVerify = make([]string, 0)
	}

	return partial
}

func PartialToGuildSettingsBorderwallSettings(guildID int64, guildSettings *GuildSettingsBorderwall) *database.GuildSettingsBorderwall {
	return &database.GuildSettingsBorderwall{
		GuildID:         guildID,
		ToggleEnabled:   guildSettings.ToggleEnabled,
		ToggleSendDm:    guildSettings.ToggleSendDm,
		Channel:         welcomer.StringPointerToInt64(guildSettings.Channel),
		MessageVerify:   welcomer.StringToJSONB(guildSettings.MessageVerify),
		MessageVerified: welcomer.StringToJSONB(guildSettings.MessageVerified),
		RolesOnJoin:     welcomer.StringSliceToInt64(guildSettings.RolesOnJoin),
		RolesOnVerify:   welcomer.StringSliceToInt64(guildSettings.RolesOnVerify),
	}
}
