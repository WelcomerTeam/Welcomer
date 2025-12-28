package service

import (
	"html"
	"strings"

	"github.com/WelcomerTeam/Discord/discord"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
)

func (is *ImageService) GenerateCanvas(ctx *ImageGenerationContext) strings.Builder {
	builder := strings.Builder{}

	builder.WriteString(`<!DOCTYPE html><html lang="en"><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width, initial-scale=1.0"></head>`)

	builder.WriteString(`<body><div id="canvas" style="`)
	is.getCanvasStyle(ctx, ctx.CustomWelcomerImage).Build(&builder)
	builder.WriteString(`">`)

	functions := welcomer.GatherFunctions(ctx.NumberLocale)
	variables := welcomer.GatherVariables(nil, &discord.GuildMember{User: &ctx.User}, welcomer.GuildVariables{
		Guild:         &ctx.Guild,
		MembersJoined: ctx.MembersJoined,
		NumberLocale:  ctx.NumberLocale,
	}, ctx.Invite, nil)

	for index, layer := range ctx.CustomWelcomerImage.Layers {
		builder.WriteString(`<div style="`)
		getObjectStyleBase(layer, len(ctx.CustomWelcomerImage.Layers), 0).Build(&builder)
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
			is.getObjectStyle(ctx, layer, len(ctx.CustomWelcomerImage.Layers), index).Build(&builder)
			builder.WriteString(`"><span>`)

			builder.WriteString(html.EscapeString(markdownValue))

			builder.WriteString(`</span></div>`)
		case welcomer.CustomWelcomerImageLayerTypeImage:
			formattedValue, err := welcomer.FormatString(functions, variables, layer.Value)
			if err != nil {
				welcomer.Logger.Error().Err(err).Int("layer_idx", index).Msg("failed to format string for custom welcomer image layer")

				continue
			}

			builder.WriteString("<img style=\"")
			is.getObjectStyle(ctx, layer, len(ctx.CustomWelcomerImage.Layers), index).Build(&builder)
			builder.WriteString(`" src="`)

			builder.WriteString(html.EscapeString(formattedValue))

			builder.WriteString(`"><img>`)
		case welcomer.CustomWelcomerImageLayerTypeShapeRectangle, welcomer.CustomWelcomerImageLayerTypeShapeCircle:
			builder.WriteString("<div style=\"")
			is.getObjectStyle(ctx, layer, len(ctx.CustomWelcomerImage.Layers), index).Build(&builder)
			builder.WriteString(`"></div>`)
		}

		builder.WriteString(`</div>`)
	}

	builder.WriteString(`</div></body>`)

	return builder
}

func FetchFonts(builder *strings.Builder, ctx *ImageGenerationContext) {
	fontsLoaded := map[string]bool{}

	for _, layer := range ctx.CustomWelcomerImage.Layers {
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
