package service

import (
	"strings"

	"github.com/WelcomerTeam/Welcomer/welcomer-core"
)

type Styling map[string]string

func (s Styling) Add(key, value string) {
	s[key] = value
}

func (s Styling) Build(builder *strings.Builder) {
	for key, value := range s {
		if key == "" || value == "" {
			continue
		}

		builder.WriteString(key + ":" + value + ";")
	}
}

func getCanvasStyle(customImage welcomer.CustomWelcomerImage) Styling {
	styling := Styling{}

	if len(customImage.Dimensions) == 2 {
		if customImage.Dimensions[0] == 0 {
			customImage.Dimensions[0] = 1000
		}

		if customImage.Dimensions[1] == 0 {
			customImage.Dimensions[1] = 300
		}

		styling.Add("width", welcomer.Itoa(int64(customImage.Dimensions[0]))+"px")
		styling.Add("height", welcomer.Itoa(int64(customImage.Dimensions[1]))+"px")
	}

	styling.Add("background", getFillAsCSS(customImage.Fill, "transparent"))
	styling.Add("box-sizing", "border-box")
	styling.Add("overflow", "hidden")
	styling.Add("position", "relative")

	if customImage.Stroke != nil && customImage.Stroke.Width > 0 {
		styling.Add("border", welcomer.Itoa(int64(customImage.Stroke.Width))+"px solid "+getFillAsCSS(customImage.Stroke.Color, "#000000"))
	}

	return styling
}

func getObjectStyleBase(layer welcomer.CustomWelcomerImageLayer, layer_count int, index int) Styling {
	styling := Styling{}

	styling.Add("position", "absolute")
	styling.Add("z-index", welcomer.Itoa(int64(layer_count-index)))
	styling.Add("box-sizing", "border-box")

	if len(layer.Dimensions) == 2 {
		styling.Add("width", welcomer.Itoa(int64(layer.Dimensions[0]))+"px")
		styling.Add("height", welcomer.Itoa(int64(layer.Dimensions[1]))+"px")
	} else {
		styling.Add("width", "auto")
		styling.Add("height", "auto")
	}

	if len(layer.Position) == 2 {
		styling.Add("left", welcomer.Itoa(int64(layer.Position[0]))+"px")
		styling.Add("top", welcomer.Itoa(int64(layer.Position[1]))+"px")
	} else {
		styling.Add("left", "0px")
		styling.Add("top", "0px")
	}

	styling.Add("border", "red 1px solid")

	styling.Add(
		"transform",
		welcomer.If(layer.Rotation != 0, "rotate("+welcomer.Itoa(int64(layer.Rotation))+"deg)", "")+
			"scale("+welcomer.If(layer.InvertedX, "-1", "1")+","+welcomer.If(layer.InvertedY, "-1", "1")+")",
	)

	return styling
}

func getObjectStyle(layer welcomer.CustomWelcomerImageLayer, layer_count int, index int) Styling {
	styling := getObjectStyleBase(layer, layer_count, index)

	styling.Add("z-index", "0")
	styling.Add("transform", "")
	styling.Add("left", "0px")
	styling.Add("top", "0px")

	if layer.Type == welcomer.CustomWelcomerImageLayerTypeShapeCircle {
		styling.Add("border-radius", "100%")
	} else if len(layer.BorderRadius) == 4 {
		styling.Add("border-radius",
			normalizeBorderRadius(layer.BorderRadius[0])+" "+
				normalizeBorderRadius(layer.BorderRadius[1])+" "+
				normalizeBorderRadius(layer.BorderRadius[2])+" "+
				normalizeBorderRadius(layer.BorderRadius[3]),
		)
	}

	if layer.Type == welcomer.CustomWelcomerImageLayerTypeText {
		styling.Add("background", "transparent")
		styling.Add("color", getFillAsCSS(layer.Fill, "inherit"))
	} else {
		styling.Add("background", getFillAsCSS(layer.Fill, "transparent"))
		styling.Add("color", "inherit")
	}

	if layer.Type == welcomer.CustomWelcomerImageLayerTypeText {
		if layer.Typography != nil {
			font, ok := Fonts[layer.Typography.FontFamily]
			if ok {
				if font.websafe {
					styling.Add("font-family", layer.Typography.FontFamily)
					styling.Add("font-weight", font.GetWeight(layer.Typography.FontWeight))
				} else {
					styling.Add("font-family", "&quot;"+font.name+"&quot;, sans-serif")
					styling.Add("font-weight", font.GetWeight(layer.Typography.FontWeight))
				}
			}

			styling.Add("font-size", welcomer.Itoa(int64(layer.Typography.FontSize))+"px")

			if layer.Typography.LineHeight != 0 {
				styling.Add("line-height", welcomer.Itoa(int64(layer.Typography.LineHeight))+"em")
			}

			if layer.Typography.LetterSpacing != 0 {
				styling.Add("letter-spacing", welcomer.Itoa(int64(layer.Typography.LetterSpacing))+"px")
			}

			if layer.Stroke.Width != 0 {
				styling.Add("text-shadow", generateTextShadow(layer.Stroke.Width, getFillAsCSS(layer.Stroke.Color, "#000000")))
			}

			styling.Add("justify-content", string(layer.Typography.HorizontalAlignment))
			styling.Add("align-items", string(layer.Typography.VerticalAlignment))
		}

		styling.Add("display", "flex")
		styling.Add("white-space", "pre-wrap")
	} else {
		if layer.Stroke.Width > 0 {
			styling.Add("border", welcomer.Itoa(int64(layer.Stroke.Width))+"px solid "+getFillAsCSS(layer.Stroke.Color, "#000000"))
		}
	}

	return styling
}

func getFillAsCSS(value string, defaultValue string) string {
	if len(value) == 0 {
		return defaultValue
	}

	if value[0] == '#' {
		return value
	}

	if value == "solid:profile" {
		// TODO: hex from avatar
		return defaultValue
	}

	if value[:4] == "ref:" {
		// TODO: artifact serve
		return defaultValue
	}

	if _, ok := backgrounds[value]; ok {
		// TODO: default image serve
		return defaultValue
	}

	return defaultValue
}

func normalizeBorderRadius(value string) string {
	if value[len(value)-1] == '%' {
		return value
	}

	if len(value) > 0 {
		return value + "px"
	}

	return "0px"
}

func generateTextShadow(width int, color string) string {
	var builder strings.Builder

	p := width * width
	first := true

	for dx := -width; dx <= width; dx++ {
		for dy := -width; dy <= width; dy++ {
			if dx*dx+dy*dy <= p {
				if !first {
					builder.WriteString(", ")
				}
				builder.WriteString(welcomer.Itoa(int64(dx)) + "px " + welcomer.Itoa(int64(dy)) + "px 0 " + color)
				first = false
			}
		}
	}

	return builder.String()
}
