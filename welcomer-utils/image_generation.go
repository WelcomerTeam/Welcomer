package utils

import (
	"regexp"
	"strings"

	"github.com/WelcomerTeam/Discord/discord"
)

var (
	CustomBackgroundPrefix = "custom:"
	SolidColourPrefix      = "solid:"
	SolidColourPrefixBased = "profile"
	UnsplashPrefix         = "unsplash:"

	RGBAPrefix = "rgba"
	RGBPrefix  = "rgb"

	fallbackColour = "#FFFFFF"

	RGBRegex  = regexp.MustCompile(`^rgb\(([0-9]+)(\w+)?, ([0-9]+)(\w+)?, ([0-9]+)(\w+)?\)$`)
	RGBARegex = regexp.MustCompile(`^rgba\(([0-9]+)(\w+)?, ([0-9]+)(\w+)?, ([0-9].+)(\w+)?\)$`)

	unsplashRegex = regexp.MustCompile(`^[a-zA-Z_-]+$`)
)

type GenerateImageOptionsRaw struct {
	ShowAvatar         bool
	AvatarURL          string
	Background         string
	Text               string
	TextFont           string
	TextColor          int64
	UserID             int64
	ProfileBorderColor int64
	GuildID            int64
	ImageBorderColor   int64
	TextStrokeColor    int64
	Theme              int32
	TextAlign          int32
	ImageBorderWidth   int32
	ProfileFloat       int32
	ProfileBorderWidth int32
	ProfileBorderCurve int32
	TextStroke         bool
	AllowAnimated      bool
}

//go:generate go-enum -f=$GOFILE --marshal

// ENUM(left, center, right, topLeft, topCenter, topRight, bottomLeft, bottomCenter, bottomRight)
type ImageAlignment int32

// ENUM(default, vertical, card)
type ImageTheme int32

// ENUM(circular, rounded, squared, hexagonal)
type ImageProfileBorderType int32

// ENUM(default, welcomer, solid, solidProfile, unsplash, url)
type BackgroundType int32

// ENUM(unknown, image/png, image/jpeg, image/gif, image/webp)
type ImageFileType int32

func (i ImageFileType) GetExtension() string {
	switch i {
	case ImageFileTypeImagePng:
		return "png"
	case ImageFileTypeImageJpeg:
		return "jpeg"
	case ImageFileTypeImageGif:
		return "gif"
	case ImageFileTypeImageWebp:
		return "webp"
	case ImageFileTypeUnknown:
		fallthrough
	default:
		return "png"
	}
}

type Background struct {
	Value string         `json:"value"`
	Type  BackgroundType `json:"type"`
}

type UserProvidedEmbed struct {
	Content string          `json:"content"`
	Embeds  []discord.Embed `json:"embeds"`
}

// ParseBackground parses a background string provided by user.
// Expected formats:
// solid:FFAAAA - Solid colour with HEX code.
// solid:profile - Solid colour based on user profile picture.
// unsplash:Bnr_ZSmqbDY - Unsplash along with Id.
// custom:018c186a-4ce5-74c7-b2d1-b0639c2f4686 - per-guild utils.background
func ParseBackground(str string) (Background, error) {
	str = strings.TrimSpace(str)

	switch {
	case strings.HasPrefix(str, SolidColourPrefix):
		// extract value
		value := strings.TrimPrefix(str, SolidColourPrefix)

		if value == SolidColourPrefixBased {
			return Background{
				Type:  BackgroundTypeSolidProfile,
				Value: "",
			}, nil
		} else if IsValidColour(value) {
			return Background{
				Type:  BackgroundTypeSolid,
				Value: value,
			}, nil
		}
	case strings.HasPrefix(str, UnsplashPrefix):
		// extract value
		value := strings.TrimPrefix(str, UnsplashPrefix)

		if IsValidUnsplashID(value) {
			return Background{
				Type:  BackgroundTypeUnsplash,
				Value: value,
			}, nil
		}
	case strings.HasPrefix(str, CustomBackgroundPrefix):
		// extract value
		value := strings.TrimPrefix(str, CustomBackgroundPrefix)

		return Background{
			Type:  BackgroundTypeWelcomer,
			Value: value,
		}, nil
	default:
		return Background{
			Type:  BackgroundTypeDefault,
			Value: str,
		}, nil
	}

	return Background{}, ErrInvalidBackground
}
