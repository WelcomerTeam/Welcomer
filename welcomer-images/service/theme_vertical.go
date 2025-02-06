package service

import (
	utils "github.com/WelcomerTeam/Welcomer/welcomer-utils"
	"github.com/disintegration/imaging"
	"github.com/fogleman/gg"
	"golang.org/x/image/font"
	"image"
)

func CreateVerticalImage(is *ImageService, args GenerateImageArguments) (resp *GenerateThemeResponse, err error) {
	imageSize := image.Rect(0, 0, 750, 516)
	padding := image.Point{32, 32}
	overlaySize := image.Rect(0, 0, 686, 452)

	im := image.NewRGBA(overlaySize)
	context := gg.NewContextForRGBA(im)

	imagePoint := image.Point{}
	textPoint := image.Point{}

	if args.Avatar == nil {
		err = drawMultiline(
			font.Drawer{Dst: im},
			is.CreateFontPackHook(args.ImageOptions.TextFont),
			MultilineArguments{
				DefaultFontSize: defaultFontSize,

				X: textPoint.X,
				Y: textPoint.Y,

				Width:  686,
				Height: 452,

				Alignment: args.ImageOptions.TextAlign,

				StrokeWeight: args.ImageOptions.TextStroke,
				StrokeColor:  args.ImageOptions.TextStrokeColor,
				TextColor:    args.ImageOptions.TextColor,

				Text: args.ImageOptions.Text,
			},
		)
	} else {
		switch args.ImageOptions.ProfileFloat {
		case utils.ImageAlignmentLeft: // left
			imagePoint = image.Point{0, 0}
			textPoint = image.Point{0, 236}
		case utils.ImageAlignmentCenter: // center
			imagePoint = image.Point{225, 0}
			textPoint = image.Point{0, 236}
		case utils.ImageAlignmentRight: // right
			imagePoint = image.Point{450, 0}
			textPoint = image.Point{0, 236}
		default:
			err = ErrUnknownProfileFloat

			return
		}

		context.DrawImage(
			imaging.Resize(
				args.Avatar,
				236,
				236,
				imaging.MitchellNetravali,
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

				Width:  686,
				Height: 200,

				Alignment: args.ImageOptions.TextAlign,

				StrokeWeight: args.ImageOptions.TextStroke,
				StrokeColor:  args.ImageOptions.TextStrokeColor,
				TextColor:    args.ImageOptions.TextColor,

				Text: args.ImageOptions.Text,
			},
		)
	}

	return &GenerateThemeResponse{
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
