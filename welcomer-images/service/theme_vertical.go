package service

import (
	"image"

	utils "github.com/WelcomerTeam/Welcomer/welcomer-utils"
	"github.com/disintegration/imaging"
	"github.com/fogleman/gg"
	"golang.org/x/image/font"
)

func CreateVerticalImage(is *ImageService, args GenerateImageArguments) (resp GenerateThemeResponse, err error) {
	imageSize := image.Rect(0, 0, 750, 516)
	padding := image.Point{32, 32}
	overlaySize := image.Rect(0, 0, 686, 452)

	im := image.NewRGBA(overlaySize)
	context := gg.NewContextForRGBA(im)

	context.DrawImage(
		imaging.Resize(
			args.Avatar,
			236,
			236,
			imaging.MitchellNetravali,
		),
		225, 0,
	)

	err = drawMultiline(
		font.Drawer{Dst: im},
		is.CreateFontPackHook(args.ImageOptions.TextFont),
		MultilineArguments{
			DefaultFontSize: defaultFontSize,

			X: 0,
			Y: 252,

			Width:  686,
			Height: 200,

			Alignment: args.ImageOptions.TextAlign,

			StrokeWeight: args.ImageOptions.TextStroke,
			StrokeColor:  args.ImageOptions.TextStrokeColor,
			TextColor:    args.ImageOptions.TextColor,

			Text: args.ImageOptions.Text,
		},
	)

	return GenerateThemeResponse{
		Overlay: im,

		TargetImageSize:   imageSize,
		TargetImageWidth:  imageSize.Dx(),
		TargetImageHeight: imageSize.Dy(),

		TargetBackgroundSize: imageSize,
		TargetBackgroundW:    imageSize.Dx(),
		TargetBackgroundH:    imageSize.Dy(),

		BackgroundAnchor: image.Point{},
		OverlayAnchor:    padding,
	}, err
}

func init() {
	registerThemeFunc(utils.ImageThemeVertical, CreateVerticalImage)
}
