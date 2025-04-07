package service

import (
	"image"

	"github.com/WelcomerTeam/Welcomer/welcomer-core"
)

var themes = make(map[welcomer.ImageTheme]func(*ImageService, GenerateImageArguments) (*GenerateThemeResponse, error))

func registerThemeFunc(theme welcomer.ImageTheme, f func(*ImageService, GenerateImageArguments) (*GenerateThemeResponse, error)) {
	themes[theme] = f
}

type GenerateImageArguments struct {
	ImageOptions GenerateImageOptions

	// Avatar with mask and background pre-applied
	Avatar image.Image
}

type GenerateThemeResponse struct {
	// Overlay
	Overlay image.Image

	// The target size of entire image
	TargetImageSize                     image.Rectangle
	TargetImageWidth, TargetImageHeight int

	// The target size of backgrounds. This is
	// equal to TargetImage however changes if
	// there is a border.
	TargetBackgroundSize                 image.Rectangle
	TargetBackgroundW, TargetBackgroundH int

	// Point to move from (0,0) when
	// rendering the backgrounds
	BackgroundAnchor image.Point

	// Point to move from (0,0) when
	// rendering the overlay
	OverlayAnchor image.Point
}
