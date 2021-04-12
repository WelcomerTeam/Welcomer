package welcomerimages

import (
	"bytes"
	"image"

	"github.com/disintegration/imaging"
	"github.com/fogleman/gg"
	"golang.org/x/image/font"
)

func CreateVerticalImage(
	wi *WelcomerImageService,
	b *bytes.Buffer, args GenerateImageArgs) (GenerateThemeResp, error) {
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
			imaging.Lanczos,
		),
		225, 0,
	)

	DrawMultiline(
		font.Drawer{Dst: im},
		wi.CreateFontPackHook(args.ImageOpts.Font),
		MultilineArguments{
			DefaultFontSize: defaultFontSize,

			X: 0,
			Y: 252,

			Width:  686,
			Height: 200,

			HorizontalAlignment: args.ImageOpts.TextAlignmentX,
			VerticalAlignment:   args.ImageOpts.TextAlignmentY,

			StrokeWeight: args.ImageOpts.TextStroke,
			StrokeColour: args.ImageOpts.TextStrokeColour,
			TextColour:   args.ImageOpts.TextColour,

			Text: args.ImageOpts.Text,
		},
	)

	return GenerateThemeResp{
		Overlay: im,

		TargetImageSize: imageSize,
		TargetImageW:    imageSize.Dx(),
		TargetImageH:    imageSize.Dy(),

		TargetBackgroundSize: imageSize,
		TargetBackgroundW:    imageSize.Dx(),
		TargetBackgroundH:    imageSize.Dy(),

		BackgroundAnchor: image.Point{},
		OverlayAnchor:    padding,
	}, nil
}

func init() {
	RegisterFormat(ThemeVertical, CreateVerticalImage)
}
