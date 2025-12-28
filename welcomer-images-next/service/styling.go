package service

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"net/http"
	"net/url"
	"strings"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/WelcomerTeam/Welcomer/welcomer-core"
)

const SolidProfileLuminance = 0.7

type ImageGenerationContext struct {
	context.Context

	GenerateRequest
}

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

func (is *ImageService) getCanvasStyle(ctx *ImageGenerationContext, customImage welcomer.CustomWelcomerImage) Styling {
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

	styling.Add("background", is.getFillAsCSS(ctx, customImage.Fill, "transparent"))
	styling.Add("box-sizing", "border-box")
	styling.Add("overflow", "hidden")
	styling.Add("position", "relative")

	if customImage.Stroke != nil && customImage.Stroke.Width > 0 {
		styling.Add("border", welcomer.Itoa(int64(customImage.Stroke.Width))+"px solid "+is.getFillAsCSS(ctx, customImage.Stroke.Color, "#000000"))
	}

	return styling
}

func getObjectStyleBase(layer welcomer.CustomWelcomerImageLayer, layer_count, index int) Styling {
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

	styling.Add(
		"transform",
		welcomer.If(layer.Rotation != 0, "rotate("+welcomer.Itoa(int64(layer.Rotation))+"deg)", "")+
			"scale("+welcomer.If(layer.InvertedX, "-1", "1")+","+welcomer.If(layer.InvertedY, "-1", "1")+")",
	)

	return styling
}

func (is *ImageService) getObjectStyle(ctx *ImageGenerationContext, layer welcomer.CustomWelcomerImageLayer, layer_count, index int) Styling {
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
		styling.Add("color", is.getFillAsCSS(ctx, layer.Fill, "inherit"))
	} else {
		styling.Add("background", is.getFillAsCSS(ctx, layer.Fill, "transparent"))
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
				styling.Add("text-shadow", generateTextShadow(layer.Stroke.Width, is.getFillAsCSS(ctx, layer.Stroke.Color, "#000000")))
			}

			styling.Add("justify-content", string(layer.Typography.HorizontalAlignment))
			styling.Add("align-items", string(layer.Typography.VerticalAlignment))
		}

		styling.Add("display", "flex")
		styling.Add("white-space", "pre-wrap")
	} else {
		if layer.Stroke.Width > 0 {
			styling.Add("border", welcomer.Itoa(int64(layer.Stroke.Width))+"px solid "+is.getFillAsCSS(ctx, layer.Stroke.Color, "#000000"))
		}
	}

	return styling
}

func (is *ImageService) getFillAsCSS(ctx *ImageGenerationContext, value, defaultValue string) string {
	if len(value) == 0 {
		return defaultValue
	}

	if value[0] == '#' {
		return value
	}

	if value == "solid:profile" {
		if ctx.User.Avatar != "" {
			if src, err := is.fetchImageFromURL(ctx.Context, welcomer.GetUserAvatar(&ctx.User)); err == nil {
				return getImageLumainceAsHex(src)
			}
		}

		return defaultValue
	}

	if value[:4] == "ref:" {
		return "url(https://www.welcomer.gg/api/guild/" + ctx.Guild.ID.String() + "/welcomer/artifact/" + url.QueryEscape(value[4:]) + ")"
	}

	if _, ok := backgrounds[value]; ok {
		return "url(https://www.welcomer.gg/assets/backgrounds/" + value + ".webp)"
	}

	return defaultValue
}

func (is *ImageService) fetchImageFromURL(ctx context.Context, avatarURL string) (image.Image, error) {
	parsedURL, isValidURL := welcomer.IsValidURL(avatarURL)
	if parsedURL == nil || !isValidURL {
		return nil, welcomer.ErrInvalidURL
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, avatarURL, nil)
	if err != nil {
		welcomer.Logger.Error().Err(err).Str("url", avatarURL).Msg("Failed to create new request for avatar")

		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := is.Client.Do(req)
	if err != nil || resp == nil {
		welcomer.Logger.Error().Err(err).Str("url", avatarURL).Msg("Failed to fetch profile picture for avatar")

		return nil, fmt.Errorf("failed to fetch image: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		welcomer.Logger.Error().Err(err).Int("status", resp.StatusCode).Str("url", avatarURL).Msg("Failed to fetch profile picture for avatar")

		return nil, fmt.Errorf("failed to fetch image, status code %d", resp.StatusCode)
	}

	im, _, err := image.Decode(resp.Body)
	if err != nil || im == nil {
		welcomer.Logger.Error().Err(err).Msg("Failed to decode profile picture of user")

		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	return im, nil
}

func getImageLumainceAsHex(src image.Image) string {
	// Initial threshold for solid profile luminance
	threshold := SolidProfileLuminance

	var colour color.RGBA
	var ok bool

	// Iterate through thresholds to find the primary color
	for threshold > 0 {
		// Attempt to identify the most common light color based on the current threshold
		colour, ok = getCommonLuminance(src, threshold)

		// If not found, decrease the threshold and try again
		if !ok {
			threshold -= 0.1
		} else {
			break // Exit the loop if a color is found
		}
	}

	return fmt.Sprintf("#%02x%02x%02x", colour.R, colour.G, colour.B)
}

// getCommonLuminance calculates the most common light color in the image.
// It takes an image.Image 'src' and a 'threshold' for defining lightness.
// It returns the most common light color as a color.RGBA value and a boolean flag ('ok') indicating success.
func getCommonLuminance(src image.Image, threshold float64) (colour color.RGBA, ok bool) {
	// Map to store the count of each color with its occurrence
	colorCount := make(map[color.RGBA]int)

	// Traverse through each pixel in the image
	bounds := src.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			// Extract RGB values of the pixel
			r, g, b, a := src.At(x, y).RGBA()

			// Check if the color is considered light based on the provided threshold
			if a > 0 && isLightColorLuminance(r, g, b, threshold) {
				// Store the color in the map and count its occurrences
				colorCount[color.RGBA{uint8(r), uint8(g), uint8(b), 255}]++
			}
		}
	}

	// If no light color found, return default black color and set 'ok' flag to false
	if len(colorCount) == 0 {
		return colour, false
	}

	// Find the most common light color by counting occurrences
	maxOccurrences := 0
	for color, count := range colorCount {
		if count > maxOccurrences {
			colour = color
			maxOccurrences = count
		}
	}

	return colour, true
}

// isLightColorLuminance determines if a color is considered 'light' based on its luminance and a threshold value.
// It takes the red (r), green (g), blue (b) values and a threshold for luminance.
// It returns true if the color is light, false otherwise.
func isLightColorLuminance(r, g, b uint32, threshold float64) bool {
	// Calculate luminance of the color using RGB values
	lum := 0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)

	// Check if the luminance exceeds the threshold
	return lum > threshold
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
