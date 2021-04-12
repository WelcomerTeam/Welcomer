package welcomerimages

import (
	"bytes"
	"image"

	"github.com/disintegration/imaging"
	"github.com/fogleman/gg"
	"golang.org/x/image/font"
)

func CreateBadgeImage(
	wi *WelcomerImageService,
	b *bytes.Buffer, args GenerateImageArgs) (GenerateThemeResp, error) {
	imageSize := image.Rect(0, 0, 964, 320)
	padding := image.Point{32, 32}
	overlaySize := image.Rect(0, 0, 900, 256)

	im := image.NewRGBA(overlaySize)
	context := gg.NewContextForRGBA(im)

	context.SetColor(args.ImageOpts.ProfileBorderColour)
	context.DrawRoundedRectangle(50, 33, 850, 190, 32)
	context.Fill()

	context.DrawImage(
		imaging.Resize(
			args.Avatar,
			256,
			256,
			imaging.Lanczos,
		),
		0, 0,
	)

	DrawMultiline(
		font.Drawer{Dst: im},
		wi.CreateFontPackHook(args.ImageOpts.Font),
		MultilineArguments{
			DefaultFontSize: defaultFontSize,

			X: 288,
			Y: 50,

			Width:  579,
			Height: 159,

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

	// profile (0,0) 236 236
	// txt 579 159 288,66
	// square (50, 49, 950, 242) same bg as image rad 16
}

func init() {
	RegisterFormat(ThemeBadge, CreateBadgeImage)
}
