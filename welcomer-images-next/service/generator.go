package service

import (
	"html"
	"strings"

	"github.com/WelcomerTeam/Welcomer/welcomer-core"
)

func GenerateCanvas(customImage welcomer.CustomWelcomerImage) strings.Builder {
	builder := strings.Builder{}

	builder.WriteString(`<!DOCTYPE html><html lang="en"><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width, initial-scale=1.0">`)

	// load font stylesheets
	FetchFonts(&builder, customImage)

	builder.WriteString(`</head>`)

	builder.WriteString(`<body><div id="canvas" style="`)
	getCanvasStyle(customImage).Build(&builder)
	builder.WriteString(`">`)

	for index, layer := range customImage.Layers {
		builder.WriteString(`<div style="`)
		getObjectStyleBase(layer, len(customImage.Layers), 0).Build(&builder)
		builder.WriteString(`">`)

		switch layer.Type {
		case welcomer.CustomWelcomerImageLayerTypeText:
			builder.WriteString("<div style=\"")
			getObjectStyle(layer, len(customImage.Layers), index).Build(&builder)
			builder.WriteString(`"><span>`)

			// TODO: this needs to output as markdown
			// TODO: this needs to be formatted.
			builder.WriteString(html.EscapeString(layer.Value))

			builder.WriteString(`</span></div>`)
		case welcomer.CustomWelcomerImageLayerTypeImage:
			builder.WriteString("<img style=\"")
			getObjectStyle(layer, len(customImage.Layers), index).Build(&builder)
			builder.WriteString(`" src="`)

			// TODO: this needs to be formatted.
			builder.WriteString(html.EscapeString(layer.Value))

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
