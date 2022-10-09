package welcomer

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/WelcomerTeam/Discord/discord"
)

var (
	ErrInvalidJSON       = fmt.Errorf("invalid json")
	ErrInvalidColour     = fmt.Errorf("colour format is not recognised")
	ErrInvalidBackground = fmt.Errorf("invalid background")

	CustomBackgroundPrefix = "custom:"
	SolidColourPrefix      = "solid:"
	SolidColourPrefixBased = "profile"
	UnsplashPrefix         = "unsplash:"

	RGBAPrefix = "rgba"
	RGBPrefix  = "rgb"

	fallbackColour = "#FFFFFF"

	RGBRegex  = regexp.MustCompile(`^rgb\(([0-9]+)\w+?, ([0-9]+)\w+?, ([0-9]+)\w+?\)$`)
	RGBARegex = regexp.MustCompile(`^rgba\(([0-9]+)\w+?, ([0-9]+)\w+?, ([0-9].+)\w+?\)$`)

	unsplashRegex = regexp.MustCompile(`^[a-zA-Z_-]+$`)
)

//go:generate go-enum -f=$GOFILE --marshal

// ENUM(left, center, right, topLeft, topCenter, topRight, bottomLeft, bottomCenter, bottomRight)
type ImageAlignment int32

// ENUM(default, vertical, card)
type ImageTheme int32

// ENUM(circular, rounded, squared, hexagonal)
type ImageProfileBorderType int32

// ENUM(default, welcomer, solid, solidProfile, unsplash, url)
type BackgroundType int32

type Background struct {
	Type BackgroundType `json:"type"`
	// Background specific values.
	Value string `json:"value"`
}

type UserProvidedEmbed struct {
	Content string          `json:"content"`
	Embeds  []discord.Embed `json:"embeds"`
}

type Colour struct {
	R uint8
	G uint8
	B uint8
	A uint8
}

// ParseBackground parses a background string provided by user.
// Expected formats:
// solid:FFAAAA - Solid colour with HEX code.
// solid:profile - Solid colour based on user profile picture.
// unsplash:Bnr_ZSmqbDY - Unsplash along with Id.
// custom_341685098468343822 - per-guild welcomer background
func ParseBackground(str string) (*Background, error) {
	str = strings.TrimSpace(str)

	switch {
	case strings.HasPrefix(str, SolidColourPrefix):
		// extract value
		value := strings.TrimPrefix(str, SolidColourPrefix)

		if value == SolidColourPrefixBased {
			return &Background{
				Type:  BackgroundTypeSolidProfile,
				Value: "",
			}, nil
		} else if isValidColour(value) {
			return &Background{
				Type:  BackgroundTypeSolid,
				Value: value,
			}, nil
		}
	case strings.HasPrefix(str, UnsplashPrefix):
		// extract value
		value := strings.TrimPrefix(str, UnsplashPrefix)

		if isValidUnsplashID(value) {
			return &Background{
				Type:  BackgroundTypeUnsplash,
				Value: value,
			}, nil
		}
	case strings.HasPrefix(str, CustomBackgroundPrefix):
		// extract value
		value := strings.TrimPrefix(str, CustomBackgroundPrefix)

		return &Background{
			Type:  BackgroundTypeWelcomer,
			Value: value,
		}, nil
	default:
		return &Background{
			Type:  BackgroundTypeDefault,
			Value: "",
		}, nil
	}

	return nil, ErrInvalidBackground
}

// ParseColour parses a colour and returns RGBA.
// Expected formats:
// #FFAAAA
// #FFAAAAFF
// RGBA(255, 255, 255, 0.1)
// RGB(255, 255, 255)
func ParseColour(str string, defaultValue string) (*Colour, error) {
	str = strings.TrimSpace(str)

	if str == "" {
		if defaultValue == "" {
			str = fallbackColour
		} else {
			str = defaultValue
		}
	}

	switch {
	case strings.HasPrefix(str, RGBPrefix):
		if strings.HasPrefix(str, RGBAPrefix) {
			// Validate against RGBA regex
			if RGBARegex.MatchString(str) {
				str = strings.TrimPrefix(str, RGBAPrefix) // (255, 255, 255, 0.1)
				str = strings.TrimPrefix(str, "(")        // 255, 255, 255, 0.1)
				str = strings.TrimSuffix(str, ")")        // 255, 255, 255, 0.1
				values := strings.Split(str, ",")         // ["255", " 255", " 255, " 0.1"]

				alphaString := strings.TrimSpace(values[3]) // 0.1
				alphaFloat, err := strconv.ParseFloat(alphaString, int64BitSize)

				if err == nil {
					var alphaInt uint8

					// If our float is above one, we will assume alpha max is 255 instead of 1.
					// We store all values as 255 so we multiply the float by 255 if it is not max 1.
					if alphaFloat > 1 {
						alphaInt = uint8(alphaFloat)
					} else {
						alphaInt = uint8(alphaFloat * 255)
					}

					// We can assume these values are ints due to our regex only allowing numerical values.
					colourR, _ := strconv.ParseInt(strings.TrimSpace(values[0]), int64Base, int64BitSize)
					colourG, _ := strconv.ParseInt(strings.TrimSpace(values[1]), int64Base, int64BitSize)
					colourB, _ := strconv.ParseInt(strings.TrimSpace(values[2]), int64Base, int64BitSize)

					if colourR > 255 || colourR < 0 {
						return nil, ErrInvalidColour
					}

					if colourG > 255 || colourG < 0 {
						return nil, ErrInvalidColour
					}

					if colourB > 255 || colourB < 0 {
						return nil, ErrInvalidColour
					}

					return &Colour{uint8(colourR), uint8(colourG), uint8(colourB), alphaInt}, nil
				}
			}
		} else {
			// Validate against RGB regex
			if RGBRegex.MatchString(str) {
				str = strings.TrimPrefix(str, RGBPrefix) // (255, 255, 255)
				str = strings.TrimPrefix(str, "(")       // 255, 255, 255)
				str = strings.TrimSuffix(str, ")")       // 255, 255, 255
				values := strings.Split(str, ",")        // ["255", " 255", " 255]

				// We can assume these values are ints due to our regex only allowing numerical values.
				colourR, _ := strconv.ParseInt(strings.TrimSpace(values[0]), int64Base, int64BitSize)
				colourG, _ := strconv.ParseInt(strings.TrimSpace(values[1]), int64Base, int64BitSize)
				colourB, _ := strconv.ParseInt(strings.TrimSpace(values[2]), int64Base, int64BitSize)

				if colourR > 255 || colourR < 0 {
					return nil, ErrInvalidColour
				}

				if colourG > 255 || colourG < 0 {
					return nil, ErrInvalidColour
				}

				if colourB > 255 || colourB < 0 {
					return nil, ErrInvalidColour
				}

				return &Colour{uint8(colourR), uint8(colourG), uint8(colourB), 255}, nil
			}
		}
	default:
		str = strings.TrimPrefix(str, "#")
		if isValidHex(str, true) {
			// We can assume these values are ints due to isValidHex.
			colourR, _ := strconv.ParseInt(strings.TrimSpace(str[0:1]), hexBase, int64BitSize)
			colourG, _ := strconv.ParseInt(strings.TrimSpace(str[2:3]), hexBase, int64BitSize)
			colourB, _ := strconv.ParseInt(strings.TrimSpace(str[4:5]), hexBase, int64BitSize)

			var colourA int64

			if len(str) == 8 {
				colourA, _ = strconv.ParseInt(strings.TrimSpace(str[6:7]), hexBase, int64BitSize)
			} else {
				colourA = 255
			}

			if colourR > 255 || colourR < 0 {
				return nil, ErrInvalidColour
			}

			if colourG > 255 || colourG < 0 {
				return nil, ErrInvalidColour
			}

			if colourB > 255 || colourB < 0 {
				return nil, ErrInvalidColour
			}

			return &Colour{uint8(colourR), uint8(colourG), uint8(colourB), uint8(colourA)}, nil
		}
	}

	return nil, ErrInvalidColour
}
