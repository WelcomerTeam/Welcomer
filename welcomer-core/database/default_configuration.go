package database

import (
	"encoding/json"
	"fmt"

	"github.com/WelcomerTeam/Discord/discord"
	utils "github.com/WelcomerTeam/Welcomer/welcomer-utils"
	"github.com/jackc/pgtype"
)

func MustConvertToJSONB(v interface{}) pgtype.JSONB {
	jsonValue, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("MustConvertToJSONB(%v): %v", v, err))
	}

	jb := pgtype.JSONB{}

	err = jb.Set(jsonValue)
	if err != nil {
		panic(fmt.Sprintf("MustConvertToJSONB(%v): %v", v, err))
	}

	return jb
}

var DefaultAutoroles GuildSettingsAutoroles = GuildSettingsAutoroles{
	ToggleEnabled: false,
	Roles:         []int64{},
}

var DefaultBorderwall GuildSettingsBorderwall = GuildSettingsBorderwall{
	ToggleEnabled: false,
	ToggleSendDm:  true,
	Channel:       0,
	MessageVerify: MustConvertToJSONB(discord.MessageParams{
		Content: "Hello world!",
		Embeds: []*discord.Embed{
			{
				Title: "Welcome to the server!",
			},
		},
	}),
	MessageVerified: pgtype.JSONB{Status: pgtype.Null},
	RolesOnJoin:     []int64{},
	RolesOnVerify:   []int64{},
}

var DefaultFreeRoles GuildSettingsFreeroles = GuildSettingsFreeroles{
	ToggleEnabled: false,
	Roles:         []int64{},
}

var DefaultLeaver GuildSettingsLeaver = GuildSettingsLeaver{
	ToggleEnabled: false,
	Channel:       0,
	MessageFormat: MustConvertToJSONB(discord.MessageParams{
		Content: "{{User.Name}} has left the server. We now have {{Guild.Members}} members.",
	}),
}

var DefaultRules GuildSettingsRules = GuildSettingsRules{
	ToggleEnabled:    false,
	ToggleDmsEnabled: true,
	Rules:            []string{},
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
	MessageFormat: MustConvertToJSONB(discord.MessageParams{
		Content: "Welcome {{User.Mention}} to **{{Guild.Name}}**! You are the {{Guild.Members}}{{Ordinal(Guild.Members)}} member!",
	}),
}

var DefaultWelcomerImages GuildSettingsWelcomerImages = GuildSettingsWelcomerImages{
	ToggleEnabled:          false,
	ToggleImageBorder:      true,
	BackgroundName:         "solid:profile",
	ColourText:             "FFFFFF",
	ColourTextBorder:       "000000",
	ColourImageBorder:      "FFFFFF",
	ColourProfileBorder:    "FFFFFF",
	ImageAlignment:         int32(utils.ImageAlignmentLeft),
	ImageTheme:             int32(utils.ImageThemeDefault),
	ImageMessage:           "Welcome {{User.Name}}\nto {{Guild.Name}}you are the {{Guild.Members}}{{Ordinal(Guild.Members)}} member!",
	ImageProfileBorderType: int32(utils.ImageProfileBorderTypeCircular),
}

var DefaultWelcomerDms GuildSettingsWelcomerDms = GuildSettingsWelcomerDms{
	ToggleEnabled:       false,
	ToggleUseTextFormat: true,
	ToggleIncludeImage:  true,
	MessageFormat:       pgtype.JSONB{Status: pgtype.Null},
}

var DefaultGuild Guilds = Guilds{
	EmbedColour:      0x2F80ED,
	SiteSplashUrl:    "",
	SiteStaffVisible: false,
	SiteGuildVisible: false,
	SiteAllowInvites: false,
}
