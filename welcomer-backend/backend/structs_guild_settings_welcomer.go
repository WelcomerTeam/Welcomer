package backend

import (
	core "github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
)

type GuildSettingsWelcomer struct {
	Text   *GuildSettingsWelcomerText   `json:"text"`
	Images *GuildSettingsWelcomerImages `json:"images"`
	DMs    *GuildSettingsWelcomerDms    `json:"dms"`
	Custom *GuildSettingsWelcomerCustom `json:"custom,omitempty"`
}

type GuildSettingsWelcomerText struct {
	Channel       *string `json:"channel"`
	MessageFormat string  `json:"message_json"`
	ToggleEnabled bool    `json:"enabled"`
}

type GuildSettingsWelcomerImages struct {
	BackgroundName         string `json:"background"`
	ColourText             string `json:"text_colour"`
	ColourTextBorder       string `json:"text_colour_border"`
	ColourImageBorder      string `json:"border_colour"`
	ColourProfileBorder    string `json:"profile_border_colour"`
	ImageAlignment         string `json:"image_alignment"`
	ImageTheme             string `json:"image_theme"`
	ImageMessage           string `json:"message"`
	ImageProfileBorderType string `json:"profile_border_type"`
	ToggleEnabled          bool   `json:"enabled"`
	ToggleImageBorder      bool   `json:"enable_border"`
}

type GuildSettingsWelcomerDms struct {
	MessageFormat       string `json:"message_json"`
	ToggleEnabled       bool   `json:"enabled"`
	ToggleUseTextFormat bool   `json:"reuse_message"`
	ToggleIncludeImage  bool   `json:"include_image"`
}

type GuildSettingsWelcomerCustom struct {
	CustomBackgroundIDs []string `json:"custom_ids"`
}

func GuildSettingsWelcomerSettingsToPartial(
	text *database.GuildSettingsWelcomerText,
	images *database.GuildSettingsWelcomerImages,
	dms *database.GuildSettingsWelcomerDms,
	Custom *GuildSettingsWelcomerCustom,
) *GuildSettingsWelcomer {
	partial := &GuildSettingsWelcomer{
		Text: &GuildSettingsWelcomerText{
			ToggleEnabled: text.ToggleEnabled,
			Channel:       Int64ToStringPointer(text.Channel),
			MessageFormat: JSONBToString(text.MessageFormat),
		},
		Images: &GuildSettingsWelcomerImages{
			ToggleEnabled:          images.ToggleEnabled,
			ToggleImageBorder:      images.ToggleImageBorder,
			BackgroundName:         images.BackgroundName,
			ColourText:             images.ColourText,
			ColourTextBorder:       images.ColourTextBorder,
			ColourImageBorder:      images.ColourImageBorder,
			ColourProfileBorder:    images.ColourProfileBorder,
			ImageAlignment:         core.ImageAlignment(images.ImageAlignment).String(),
			ImageTheme:             core.ImageTheme(images.ImageTheme).String(),
			ImageMessage:           images.ImageMessage,
			ImageProfileBorderType: core.ImageProfileBorderType(images.ImageProfileBorderType).String(),
		},
		DMs: &GuildSettingsWelcomerDms{
			ToggleEnabled:       dms.ToggleEnabled,
			ToggleUseTextFormat: dms.ToggleUseTextFormat,
			ToggleIncludeImage:  dms.ToggleIncludeImage,
			MessageFormat:       JSONBToString(dms.MessageFormat),
		},
		Custom: Custom,
	}

	return partial
}

func PartialToGuildSettingsWelcomerSettings(guildID int64, guildSettings *GuildSettingsWelcomer) (*database.GuildSettingsWelcomerText, *database.GuildSettingsWelcomerImages, *database.GuildSettingsWelcomerDms) {
	return &database.GuildSettingsWelcomerText{
			GuildID:       guildID,
			ToggleEnabled: guildSettings.Text.ToggleEnabled,
			Channel:       StringPointerToInt64(guildSettings.Text.Channel),
			MessageFormat: StringToJSONB(guildSettings.Text.MessageFormat),
		}, &database.GuildSettingsWelcomerImages{
			GuildID:                guildID,
			ToggleEnabled:          guildSettings.Images.ToggleEnabled,
			ToggleImageBorder:      guildSettings.Images.ToggleImageBorder,
			BackgroundName:         guildSettings.Images.BackgroundName,
			ColourText:             guildSettings.Images.ColourText,
			ColourTextBorder:       guildSettings.Images.ColourTextBorder,
			ColourImageBorder:      guildSettings.Images.ColourImageBorder,
			ColourProfileBorder:    guildSettings.Images.ColourProfileBorder,
			ImageAlignment:         int32(ParseImageAlignment(guildSettings.Images.ImageAlignment)),
			ImageTheme:             int32(ParseImageTheme(guildSettings.Images.ImageTheme)),
			ImageMessage:           guildSettings.Images.ImageMessage,
			ImageProfileBorderType: int32(ParseImageProfileBorderType(guildSettings.Images.ImageProfileBorderType)),
		}, &database.GuildSettingsWelcomerDms{
			GuildID:             guildID,
			ToggleEnabled:       guildSettings.DMs.ToggleEnabled,
			ToggleUseTextFormat: guildSettings.DMs.ToggleUseTextFormat,
			ToggleIncludeImage:  guildSettings.DMs.ToggleIncludeImage,
			MessageFormat:       StringToJSONB(guildSettings.DMs.MessageFormat),
		}
}

func ParseImageAlignment(value string) core.ImageAlignment {
	imageAlignment, _ := core.ParseImageAlignment(value)

	return imageAlignment
}

func ParseImageTheme(value string) core.ImageTheme {
	imageTheme, _ := core.ParseImageTheme(value)

	return imageTheme
}

func ParseImageProfileBorderType(value string) core.ImageProfileBorderType {
	imageProfileBorderType, _ := core.ParseImageProfileBorderType(value)

	return imageProfileBorderType
}
