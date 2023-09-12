package backend

import (
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
)

type GuildSettingsTempChannels struct {
	ToggleEnabled    bool  `json:"enabled"`
	ToggleAutopurge  bool  `json:"toggle_autopurge"`
	ChannelLobby     int64 `json:"channel_lobby"`
	ChannelCategory  int64 `json:"channel_category"`
	DefaultUserCount int32 `json:"default_user_count"`
}

func GuildSettingsTempChannelsSettingsToPartial(
	tempChannels *database.GuildSettingsTempchannels,
) *GuildSettingsTempChannels {
	partial := &GuildSettingsTempChannels{
		ToggleEnabled:    tempChannels.ToggleEnabled,
		ToggleAutopurge:  tempChannels.ToggleAutopurge,
		ChannelLobby:     tempChannels.ChannelLobby,
		ChannelCategory:  tempChannels.ChannelCategory,
		DefaultUserCount: tempChannels.DefaultUserCount,
	}

	return partial
}

func PartialToGuildSettingsTempChannelsSettings(guildID int64, guildSettings *GuildSettingsTempChannels) *database.GuildSettingsTempchannels {
	return &database.GuildSettingsTempchannels{
		GuildID:          guildID,
		ToggleEnabled:    guildSettings.ToggleEnabled,
		ToggleAutopurge:  guildSettings.ToggleAutopurge,
		ChannelLobby:     guildSettings.ChannelLobby,
		ChannelCategory:  guildSettings.ChannelCategory,
		DefaultUserCount: guildSettings.DefaultUserCount,
	}
}
