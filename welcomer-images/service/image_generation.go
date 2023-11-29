package service

import (
	"image"
	"image/color"
	"image/draw"
	"sync"

	"github.com/WelcomerTeam/Discord/discord"
	core "github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/disintegration/imaging"
	"github.com/fogleman/gg"
)

// FullImage stores the image and any extra information.
type FullImage struct {
	// The image format that is represented
	Format core.ImageFileType

	Frames []image.Image

	// Config is the global color table (palette), width and height. A nil or
	// empty-color.Palette Config.ColorModel means that each frame has its own
	// color table and there is no global color table.
	Config image.Config

	// The successive delay times, one per frame, in 100ths of a second.
	Delay []int

	// LoopCount controls the number of times an animation will be
	// restarted during display.
	LoopCount int

	// Disposal is the successive disposal methods, one per frame.
	Disposal []byte

	// BackgroundIndex is the background index in the global color table, for
	// use with the DisposalBackground disposal method.
	BackgroundIndex byte
}

type GenerateImageOptions struct {
	GuildID             discord.Snowflake
	UserID              discord.Snowflake
	AllowAnimated       bool
	AvatarURL           string
	Theme               core.ImageTheme
	Background          string
	Text                string
	TextFont            string
	TextStroke          int
	TextHorizontalAlign core.ImageAlignment
	TextVerticalAlign   core.ImageAlignment
	TextColor           color.RGBA
	TextStrokeColor     color.RGBA
	ImageBorderColor    color.RGBA
	ImageBorderWidth    int
	ProfileFloat        core.ImageAlignment
	ProfileBorderColor  color.RGBA
	ProfileBorderWidth  int
	ProfileBorderCurve  core.ImageProfileBorderType
}

func (is *ImageService) GenerateImage(imageOptions GenerateImageOptions) ([]byte, core.ImageFileType, error) {
	theme, ok := themes[imageOptions.Theme]
	if !ok {
		theme = themes[core.ImageThemeDefault]
	}

	avatar, err := is.FetchAvatar(imageOptions.UserID, imageOptions.AvatarURL)
	if err != nil {
		is.Logger.Error().Err(err).Msg("Failed to fetch avatar")

		avatar = assetsDefaultAvatarImage
	}

	background, err := is.FetchBackground(imageOptions.Background, imageOptions.AllowAnimated, avatar)
	if err != nil {
		is.Logger.Error().Err(err).Str("background", imageOptions.Background).Msg("Failed to fetch background")

		background = &FullImage{Frames: []image.Image{backgroundsDefaultImage}}
	}

	avatarOverlay, err := applyAvatarEffects(avatar, imageOptions)
	if err != nil {
		is.Logger.Error().Err(err).Msg("Failed to generate avatar")
	}

	themeResponse, err := theme(is, GenerateImageArguments{
		ImageOptions: imageOptions,
		Avatar:       avatarOverlay,
	})
	if err != nil {
		is.Logger.Error().Err(err).Msg("Failed to generate theme overlay")

		return nil, core.ImageFileTypeUnknown, err
	}

	if imageOptions.ImageBorderWidth > 0 {
		applyImageBorder(&themeResponse, imageOptions)
	}

	frames := overlayFrames(&themeResponse, background)

	background.Config = image.Config{
		ColorModel: nil,
		Width:      themeResponse.TargetImageWidth,
		Height:     themeResponse.TargetImageHeight,
	}

	background.Disposal = nil

	file, format, err := encodeFrames(frames, background)
	if err != nil {
		is.Logger.Error().Err(err).Msg("Failed to encode frames")
	}

	return file, format, err
}

// overlay frames
func overlayFrames(themeResponse *GenerateThemeResponse, background *FullImage) []image.Image {
	wg := sync.WaitGroup{}

	frames := make([]image.Image, len(background.Frames))

	for frameNumber, frame := range background.Frames {
		wg.Add(1)

		go func(frameNumber int, frame image.Image) {
			resizedFrame := image.NewRGBA(themeResponse.TargetImageSize)

			// Draw resized background frame
			draw.Draw(
				resizedFrame, resizedFrame.Rect.Add(themeResponse.BackgroundAnchor),
				imaging.Fill(
					frame,
					themeResponse.TargetBackgroundW, themeResponse.TargetBackgroundH,
					imaging.Center, imaging.Lanczos,
				),
				image.Point{}, draw.Src,
			)

			// Draw overlay on top
			draw.Draw(
				resizedFrame, resizedFrame.Rect.Add(themeResponse.OverlayAnchor),
				themeResponse.Overlay,
				image.Point{}, draw.Over,
			)

			frames[frameNumber] = resizedFrame

			wg.Done()
		}(frameNumber, frame)
	}

	wg.Wait()

	return frames
}

// apply image border
func applyImageBorder(themeResponse *GenerateThemeResponse, imageOptions GenerateImageOptions) {
	border := image.Point{imageOptions.ImageBorderWidth, imageOptions.ImageBorderWidth}
	d := border.Add(border)

	// Increases size and adds offset to TargetImageSize
	themeResponse.TargetImageWidth += d.X
	themeResponse.TargetImageHeight += d.Y
	themeResponse.TargetImageSize.Max = themeResponse.TargetImageSize.Max.Add(d)

	borderOverlay := image.NewRGBA(themeResponse.TargetImageSize)

	context := gg.NewContextForRGBA(borderOverlay)

	context.SetColor(imageOptions.ImageBorderColor)

	// top
	context.DrawRectangle(
		0,
		0,
		float64(themeResponse.TargetImageWidth),
		float64(border.X),
	)

	// right
	context.DrawRectangle(
		float64(themeResponse.TargetImageWidth-border.X),
		float64(border.Y),
		float64(border.X),
		float64(themeResponse.TargetBackgroundW-(border.Y*2)),
	)

	// bottom
	context.DrawRectangle(
		0,
		float64(themeResponse.TargetImageHeight-border.Y),
		float64(themeResponse.TargetImageWidth),
		float64(border.Y),
	)

	// left
	context.DrawRectangle(
		0,
		float64(border.Y),
		float64(border.X),
		float64(themeResponse.TargetImageHeight-(border.Y*2)),
	)

	context.Fill()

	context.DrawImage(
		themeResponse.Overlay,
		border.X+themeResponse.OverlayAnchor.X,
		border.Y+themeResponse.OverlayAnchor.Y,
	)

	themeResponse.Overlay = borderOverlay

	themeResponse.OverlayAnchor = image.Point{}
	themeResponse.BackgroundAnchor = themeResponse.BackgroundAnchor.Add(border)
}
