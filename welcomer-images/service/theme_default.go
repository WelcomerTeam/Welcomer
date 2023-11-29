package service

import (
	"image"

	core "github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/disintegration/imaging"
	"github.com/fogleman/gg"
	"golang.org/x/image/font"
)

func CreateRegularImage(
	is *ImageService, args GenerateImageArguments) (resp GenerateThemeResponse, err error) {
	imageSize := image.Rect(0, 0, 1000, 300)
	padding := image.Point{32, 32}
	overlaySize := image.Rect(0, 0, 936, 236)

	im := image.NewRGBA(overlaySize)
	context := gg.NewContextForRGBA(im)

	imagePoint := image.Point{}
	textPoint := image.Point{}

	switch args.ImageOptions.ProfileFloat {
	case core.ImageAlignmentLeft: // left
		imagePoint = image.Point{0, 0}
		textPoint = image.Point{268, 0}
	case core.ImageAlignmentRight: // right
		imagePoint = image.Point{700, 0}
		textPoint = image.Point{0, 0}
	default:
		err = ErrUnknownProfileFloat

		return
	}

	context.DrawImage(
		imaging.Resize(
			args.Avatar,
			236,
			236,
			imaging.Lanczos,
		),
		imagePoint.X, imagePoint.Y,
	)

	err = drawMultiline(
		font.Drawer{Dst: im},
		is.CreateFontPackHook(args.ImageOptions.TextFont),
		MultilineArguments{
			DefaultFontSize: defaultFontSize,

			X: textPoint.X,
			Y: textPoint.Y,

			Width:  668,
			Height: 236,

			HorizontalAlignment: args.ImageOptions.TextHorizontalAlign,
			VerticalAlignment:   args.ImageOptions.TextVerticalAlign,

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
	registerThemeFunc(core.ImageThemeDefault, CreateRegularImage)
}
