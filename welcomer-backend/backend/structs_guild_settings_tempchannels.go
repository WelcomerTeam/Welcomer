package backend

import (
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
)

type GuildSettingsTempChannels struct {
	ToggleEnabled    bool    `json:"enabled"`
	ToggleAutopurge  bool    `json:"autopurge"`
	ChannelLobby     *string `json:"channel_lobby"`
	ChannelCategory  *string `json:"channel_category"`
	DefaultUserCount int32   `json:"default_user_count"`
}

func GuildSettingsTempChannelsSettingsToPartial(
	tempChannels *database.GuildSettingsTempchannels,
) *GuildSettingsTempChannels {
	partial := &GuildSettingsTempChannels{
		ToggleEnabled:    tempChannels.ToggleEnabled,
		ToggleAutopurge:  tempChannels.ToggleAutopurge,
		ChannelLobby:     Int64ToStringPointer(tempChannels.ChannelLobby),
		ChannelCategory:  Int64ToStringPointer(tempChannels.ChannelCategory),
		DefaultUserCount: tempChannels.DefaultUserCount,
	}

	return partial
}

func PartialToGuildSettingsTempChannelsSettings(guildID int64, guildSettings *GuildSettingsTempChannels) *database.GuildSettingsTempchannels {
	return &database.GuildSettingsTempchannels{
		GuildID:          guildID,
		ToggleEnabled:    guildSettings.ToggleEnabled,
		ToggleAutopurge:  guildSettings.ToggleAutopurge,
		ChannelLobby:     StringPointerToInt64(guildSettings.ChannelLobby),
		ChannelCategory:  StringPointerToInt64(guildSettings.ChannelCategory),
		DefaultUserCount: guildSettings.DefaultUserCount,
	}
}
