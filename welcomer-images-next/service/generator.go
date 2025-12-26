package service

import (
	"html"
	"strings"
	"time"

	"github.com/WelcomerTeam/Discord/discord"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
)

func GenerateCanvas(customImage welcomer.CustomWelcomerImage) strings.Builder {
	builder := strings.Builder{}

	builder.WriteString(`<!DOCTYPE html><html lang="en"><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width, initial-scale=1.0"></head>`)

	builder.WriteString(`<body><div id="canvas" style="`)
	getCanvasStyle(customImage).Build(&builder)
	builder.WriteString(`">`)

	functions := welcomer.GatherFunctions(database.NumberLocale(database.NumberLocaleDots))
	variables := welcomer.GatherVariables(nil, &discord.GuildMember{
		JoinedAt:                   time.Time{},
		CommunicationDisabledUntil: &time.Time{},
		PremiumSince:               &time.Time{},
		User: &discord.User{
			Banner:        "",
			GlobalName:    "",
			Avatar:        "",
			Username:      "",
			Discriminator: "",
			Locale:        "",
			Email:         "",
			ID:            0,
			PremiumType:   0,
			Flags:         0,
			AccentColor:   0,
			PublicFlags:   0,
			MFAEnabled:    false,
			Verified:      false,
			Bot:           false,
			System:        false,
		},
		Nick:    "",
		Avatar:  "",
		Roles:   discord.SnowflakeList{},
		Deaf:    false,
		Mute:    false,
		Pending: false,
	}, welcomer.GuildVariables{
		Guild:         &discord.Guild{},
		MembersJoined: 123,
		NumberLocale:  database.NumberLocale(database.NumberLocaleDots),
	}, nil, nil)

	for index, layer := range customImage.Layers {
		builder.WriteString(`<div style="`)
		getObjectStyleBase(layer, len(customImage.Layers), 0).Build(&builder)
		builder.WriteString(`">`)

		switch layer.Type {
		case welcomer.CustomWelcomerImageLayerTypeText:
			formattedValue, err := welcomer.FormatString(functions, variables, layer.Value)
			if err != nil {
				welcomer.Logger.Error().Err(err).Int("layer_idx", index).Msg("failed to format string for custom welcomer image layer")

				continue
			}

			markdownValue, err := Render(formattedValue)
			if err != nil {
				welcomer.Logger.Error().Err(err).Int("layer_idx", index).Msg("failed to render markdown for custom welcomer image layer")

				continue
			}

			builder.WriteString("<div style=\"")
			getObjectStyle(layer, len(customImage.Layers), index).Build(&builder)
			builder.WriteString(`"><span>`)

			builder.WriteString(html.EscapeString(markdownValue))

			builder.WriteString(`</span></div>`)
		case welcomer.CustomWelcomerImageLayerTypeImage:
			formattedValue, err := welcomer.FormatString(functions, variables, layer.Value)
			if err != nil {
				welcomer.Logger.Error().Err(err).Int("layer_idx", index).Msg("failed to format string for custom welcomer image layer")

				continue
			}

			markdownValue, err := Render(formattedValue)
			if err != nil {
				welcomer.Logger.Error().Err(err).Int("layer_idx", index).Msg("failed to render markdown for custom welcomer image layer")

				continue
			}

			builder.WriteString("<img style=\"")
			getObjectStyle(layer, len(customImage.Layers), index).Build(&builder)
			builder.WriteString(`" src="`)

			builder.WriteString(html.EscapeString(markdownValue))

			builder.WriteString(`"><img>`)
		case welcomer.CustomWelcomerImageLayerTypeShapeRectangle, welcomer.CustomWelcomerImageLayerTypeShapeCircle:
			builder.WriteString("<div style=\"")
			getObjectStyle(layer, len(customImage.Layers), index).Build(&builder)
			builder.WriteString(`"></div>`)
		}

		builder.WriteString(`</div>`)
	}

	builder.WriteString(`</div></body>`)

	return builder
}

func FetchFonts(builder *strings.Builder, customImage welcomer.CustomWelcomerImage) {
	fontsLoaded := map[string]bool{}

	for _, layer := range customImage.Layers {
		if layer.Type != welcomer.CustomWelcomerImageLayerTypeText {
			continue
		}

		if layer.Typography == nil {
			continue
		}

		fontFamily := layer.Typography.FontFamily
		if fontFamily == "" {
			continue
		}

		font, ok := Fonts[fontFamily]
		if !ok {
			continue
		}

		if font.websafe {
			continue
		}

		weight := layer.Typography.FontWeight
		if _, ok := font.weights[weight]; !ok {
			weight = font.defaultWeight
		}

		if fontsLoaded[fontFamily+"-"+weight] {
			continue
		}

		fontsLoaded[fontFamily+"-"+weight] = true

		builder.WriteString(`<link href="https://fonts.googleapis.com/css2?family=` + html.EscapeString(fontFamily) + `:wght@` + html.EscapeString(font.weights[weight]) + `&display=block" rel="stylesheet">`)
	}
}
