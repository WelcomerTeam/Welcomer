package welcomer

import (
	"encoding/json"
	"fmt"

	"github.com/WelcomerTeam/Discord/discord"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	"github.com/jackc/pgtype"
)

func MustConvertToJSONB(v any) pgtype.JSONB {
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

var DefaultAutoroles database.GuildSettingsAutoroles = database.GuildSettingsAutoroles{
	ToggleEnabled: false,
	Roles:         []int64{},
}

var DefaultBorderwall database.GuildSettingsBorderwall = database.GuildSettingsBorderwall{
	ToggleEnabled: false,
	ToggleSendDm:  true,
	Channel:       0,
	MessageVerify: MustConvertToJSONB(discord.MessageParams{
		Embeds: []discord.Embed{
			{
				Description: "This server is protected by Borderwall. Please verify at {{Borderwall.Link}}.",
				Color:       EmbedColourInfo,
			},
		},
	}),
	MessageVerified: MustConvertToJSONB(discord.MessageParams{
		Embeds: []discord.Embed{
			{
				Description: "Thank you for verifying! You now have access to the server.",
				Color:       EmbedColourSuccess,
			},
		},
	}),
	RolesOnJoin:   []int64{},
	RolesOnVerify: []int64{},
}

var DefaultFreeRoles database.GuildSettingsFreeroles = database.GuildSettingsFreeroles{
	ToggleEnabled: false,
	Roles:         []int64{},
}

var DefaultLeaver database.GuildSettingsLeaver = database.GuildSettingsLeaver{
	ToggleEnabled: false,
	Channel:       0,
	MessageFormat: MustConvertToJSONB(discord.MessageParams{
		Content: "{{User.Name}} has left the server. We now have {{Guild.Members}} members.",
	}),
	AutoDeleteLeaverMessages: false,
	LeaverMessageLifetime:    0,
}

var DefaultRules database.GuildSettingsRules = database.GuildSettingsRules{
	ToggleEnabled:    false,
	ToggleDmsEnabled: true,
	Rules:            []string{},
}

var DefaultTempChannels database.GuildSettingsTempchannels = database.GuildSettingsTempchannels{
	ToggleEnabled:    false,
	ToggleAutopurge:  true,
	ChannelLobby:     0,
	ChannelCategory:  0,
	DefaultUserCount: 0,
}

var DefaultTimeRoles database.GuildSettingsTimeroles = database.GuildSettingsTimeroles{
	ToggleEnabled: false,
	Timeroles:     pgtype.JSONB{Status: pgtype.Null},
}

var DefaultWelcomerText database.GuildSettingsWelcomerText = database.GuildSettingsWelcomerText{
	ToggleEnabled: false,
	Channel:       0,
	MessageFormat: MustConvertToJSONB(discord.MessageParams{
		Content: "Welcome {{User.Mention}} to **{{Guild.Name}}**! You are the {{Ordinal(Guild.Members)}} member!",
	}),
}

var DefaultWelcomer database.GuildSettingsWelcomer = database.GuildSettingsWelcomer{
	AutoDeleteWelcomeMessages:        false,
	WelcomeMessageLifetime:           0,
	AutoDeleteWelcomeMessagesOnLeave: false,
}

var DefaultWelcomerImages database.GuildSettingsWelcomerImages = database.GuildSettingsWelcomerImages{
	ToggleEnabled:          false,
	ToggleImageBorder:      true,
	ToggleShowAvatar:       true,
	BackgroundName:         "solid:profile",
	ColourText:             "#FFFFFF",
	ColourTextBorder:       "#000000",
	ColourImageBorder:      "#FFFFFF",
	ColourProfileBorder:    "#FFFFFF",
	ImageAlignment:         int32(ImageAlignmentLeft),
	ImageTheme:             int32(ImageThemeDefault),
	ImageMessage:           "Welcome {{User.Name}}\nto {{Guild.Name}}you are the {{Ordinal(Guild.Members)}} member!",
	ImageProfileBorderType: int32(ImageProfileBorderTypeCircular),
	UseCustomBuilder:       false,
	CustomBuilderData: pgtype.JSONB{
		Status: pgtype.Present,
		Bytes:  []byte("{}"),
	},
}

var DefaultWelcomerDms database.GuildSettingsWelcomerDms = database.GuildSettingsWelcomerDms{
	ToggleEnabled:       false,
	ToggleUseTextFormat: true,
	ToggleIncludeImage:  true,
	MessageFormat: MustConvertToJSONB(discord.MessageParams{
		Content: "Welcome {{User.Mention}} to **{{Guild.Name}}**! You are the {{Ordinal(Guild.Members)}} member!",
	}),
}

var DefaultGuild database.Guilds = database.Guilds{
	EmbedColour:      EmbedColourInfo,
	SiteSplashUrl:    "",
	SiteStaffVisible: false,
	SiteGuildVisible: false,
	SiteAllowInvites: false,
}
