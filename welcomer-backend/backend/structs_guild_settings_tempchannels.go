package backend

import (
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
)

type GuildSettingsTempChannels struct {
	ChannelLobby     *string `json:"channel_lobby"`
	ChannelCategory  *string `json:"channel_category"`
	DefaultUserCount int32   `json:"default_user_count"`
	ToggleEnabled    bool    `json:"enabled"`
	ToggleAutopurge  bool    `json:"autopurge"`
}

func GuildSettingsTempChannelsSettingsToPartial(
	tempChannels *database.GuildSettingsTempchannels,
) *GuildSettingsTempChannels {
	partial := &GuildSettingsTempChannels{
		ToggleEnabled:    tempChannels.ToggleEnabled,
		ToggleAutopurge:  tempChannels.ToggleAutopurge,
		ChannelLobby:     welcomer.Int64ToStringPointer(tempChannels.ChannelLobby),
		ChannelCategory:  welcomer.Int64ToStringPointer(tempChannels.ChannelCategory),
		DefaultUserCount: tempChannels.DefaultUserCount,
	}

	return partial
}

func PartialToGuildSettingsTempChannelsSettings(guildID int64, guildSettings *GuildSettingsTempChannels) *database.GuildSettingsTempchannels {
	return &database.GuildSettingsTempchannels{
		GuildID:          guildID,
		ToggleEnabled:    guildSettings.ToggleEnabled,
		ToggleAutopurge:  guildSettings.ToggleAutopurge,
		ChannelLobby:     welcomer.StringPointerToInt64(guildSettings.ChannelLobby),
		ChannelCategory:  welcomer.StringPointerToInt64(guildSettings.ChannelCategory),
		DefaultUserCount: guildSettings.DefaultUserCount,
	}
}
