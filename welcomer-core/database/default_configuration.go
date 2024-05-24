package database

import "github.com/jackc/pgtype"

var DefaultAutoroles GuildSettingsAutoroles = GuildSettingsAutoroles{
	ToggleEnabled: false,
	Roles:         []int64{},
}

var DefaultBorderwall GuildSettingsBorderwall = GuildSettingsBorderwall{
	ToggleEnabled:   false,
	ToggleSendDm:    true,
	Channel:         0,
	MessageVerify:   pgtype.JSONB{Status: pgtype.Null},
	MessageVerified: pgtype.JSONB{Status: pgtype.Null},
	RolesOnJoin:     make([]int64, 0),
	RolesOnVerify:   make([]int64, 0),
}

var DefaultFreeRoles GuildSettingsFreeroles = GuildSettingsFreeroles{
	ToggleEnabled: false,
	Roles:         make([]int64, 0),
}

var DefaultLeaver GuildSettingsLeaver = GuildSettingsLeaver{
	ToggleEnabled: false,
	Channel:       0,
	MessageFormat: pgtype.JSONB{Status: pgtype.Null},
}

var DefaultRules GuildSettingsRules = GuildSettingsRules{
	ToggleEnabled:    false,
	ToggleDmsEnabled: true,
	Rules:            make([]string, 0),
}

var DefaultTempChannels GuildSettingsTempchannels = GuildSettingsTempchannels{
	ToggleEnabled:    false,
	ToggleAutopurge:  true,
	ChannelLobby:     0,
	ChannelCategory:  0,
	DefaultUserCount: 0,
}

var DefaultTimeRoles GuildSettingsTimeroles = GuildSettingsTimeroles{
	ToggleEnabled: false,
	Timeroles:     pgtype.JSONB{Status: pgtype.Null},
}

var DefaultWelcomerText GuildSettingsWelcomerText = GuildSettingsWelcomerText{
	ToggleEnabled: false,
	Channel:       0,
	MessageFormat: pgtype.JSONB{Status: pgtype.Null},
}

var DefaultWelcomerImages GuildSettingsWelcomerImages = GuildSettingsWelcomerImages{
	ToggleEnabled:          false,
	ToggleImageBorder:      true,
	BackgroundName:         "",
	ColourText:             "FFFFFF",
	ColourTextBorder:       "000000",
	ColourImageBorder:      "FFFFFF",
	ColourProfileBorder:    "",
	ImageAlignment:         0,
	ImageTheme:             0,
	ImageMessage:           "",
	ImageProfileBorderType: 0,
}

var DefaultWelcomerDms GuildSettingsWelcomerDms = GuildSettingsWelcomerDms{
	ToggleEnabled:       false,
	ToggleUseTextFormat: false,
	ToggleIncludeImage:  false,
	MessageFormat:       pgtype.JSONB{Status: pgtype.Null},
}

var DefaultGuild Guilds = Guilds{
	EmbedColour:      0,
	SiteSplashUrl:    "",
	SiteStaffVisible: false,
	SiteGuildVisible: false,
	SiteAllowInvites: false,
}
