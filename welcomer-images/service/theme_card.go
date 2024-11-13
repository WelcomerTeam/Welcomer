package service

import (
	"image"

	utils "github.com/WelcomerTeam/Welcomer/welcomer-utils"
	"github.com/disintegration/imaging"
	"github.com/fogleman/gg"
	"golang.org/x/image/font"
)

func CreateBadgeImage(is *ImageService, args GenerateImageArguments) (resp *GenerateThemeResponse, err error) {
	imageSize := image.Rect(0, 0, 964, 320)
	padding := image.Point{32, 32}
	overlaySize := image.Rect(0, 0, 900, 256)

	im := image.NewRGBA(overlaySize)
	context := gg.NewContextForRGBA(im)

	context.SetColor(args.ImageOptions.ProfileBorderColor)
	context.DrawRoundedRectangle(50, 33, 850, 190, 32)
	context.Fill()

	context.DrawImage(
		imaging.Resize(
			args.Avatar,
			256,
			256,
			imaging.MitchellNetravali,
		),
		0, 0,
	)

	err = drawMultiline(
		font.Drawer{Dst: im},
		is.CreateFontPackHook(args.ImageOptions.TextFont),
		MultilineArguments{
			DefaultFontSize: defaultFontSize,

			X: 288,
			Y: 50,

			Width:  579,
			Height: 159,

			Alignment: args.ImageOptions.TextAlign,

			StrokeWeight: args.ImageOptions.TextStroke,
			StrokeColor:  args.ImageOptions.TextStrokeColor,
			TextColor:    args.ImageOptions.TextColor,

			Text: args.ImageOptions.Text,
		},
	)

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
	registerThemeFunc(utils.ImageThemeCard, CreateBadgeImage)
}
