package welcomerimages

import (
	"bytes"
	"image"

	"github.com/disintegration/imaging"
	"github.com/fogleman/gg"
	"golang.org/x/image/font"
)

func CreateDefaultImage(
	wi *WelcomerImageService,
	b *bytes.Buffer, args GenerateImageArgs) (GenerateThemeResp, error) {
	imageSize := image.Rect(0, 0, 1000, 300)
	padding := image.Point{32, 32}
	overlaySize := image.Rect(
		0, 0,
		imageSize.Dx()-(padding.X*2),
		imageSize.Dy()-(padding.Y*2),
	)

	im := image.NewRGBA(overlaySize)
	context := gg.NewContextForRGBA(im)

	imagePoint := image.Point{}
	textPoint := image.Point{}

	switch args.ImageOpts.ProfileAlignment {
	case FloatLeft: // left
		imagePoint = image.Point{0, 0}
		textPoint = image.Point{overlaySize.Dy() + 32, 0}
	case FloatRight: // right
		imagePoint = image.Point{overlaySize.Dx() - overlaySize.Dy(), 0}
		textPoint = image.Point{0, 0}
	}

	context.DrawImage(
		imaging.Resize(
			args.Avatar,
			overlaySize.Dy(),
			overlaySize.Dy(),
			imaging.Lanczos,
		),
		imagePoint.X, imagePoint.Y,
	)

	DrawMultiline(
		font.Drawer{Dst: im},
		wi.CreateFontPackHook(args.ImageOpts.Font),
		MultilineArguments{
			DefaultFontSize: defaultFontSize,

			X: textPoint.X,
			Y: textPoint.Y,

			Width:  overlaySize.Dx() - overlaySize.Dy() - 32,
			Height: overlaySize.Dy(),

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
