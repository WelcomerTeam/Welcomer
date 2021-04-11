package welcomerimages

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/png"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/WelcomerTeam/WelcomerImages/pkg/multiface"
	"github.com/disintegration/imaging"
	"github.com/fogleman/gg"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
	"golang.org/x/xerrors"
)

const (
	fontResize      = 0.9
	defaultFontSize = 96
	transAnchor     = 32
	avatarSize      = 256
)

var (
	themesMu = sync.RWMutex{}
	themes   = make(map[Theme]func(*WelcomerImageService, *bytes.Buffer, GenerateImageArgs) (GenerateThemeResp, error))

	transparent = color.NRGBA{0, 0, 0, 0}
)

func RegisterFormat(theme Theme, f func(*WelcomerImageService, *bytes.Buffer, GenerateImageArgs) (GenerateThemeResp, error)) {
	themesMu.Lock()
	themes[theme] = f
	themesMu.Unlock()

	println("Registered", theme)
}

// EncodeImages encodes a list of []image.Image and takes an input of AllowGifs.
// Outputs the file extension.
func (wi *WelcomerImageService) EncodeImages(b *bytes.Buffer, frames []image.Image, im *ImageCache) (string, error) {
	var err error

	start := time.Now().UTC()

	if frames == nil {
		frames = im.Frames
	}

	if len(frames) == 0 {
		return "", xerrors.New("empty frame list")
	}

	if len(frames) > 1 {
		_frames := make([]*image.Paletted, len(frames))

		wg := sync.WaitGroup{}
		for framenum, frame := range frames {
			wg.Add(1)

			go func(framenum int, frame image.Image) {
				t := quantizationLimiter.Wait()

				p, err := quantizeImage(wi.Logger, frame)
				if err != nil {
					wi.Logger.Error().Err(err).Msg("Failed to quantize frame")
				}

				_frames[framenum] = p

				quantizationLimiter.FreeTicket(t)

				wg.Done()
			}(framenum, frame)
		}

		wg.Wait()

		quant_end := time.Since(start)

		err = gif.EncodeAll(b, &gif.GIF{
			Image:           _frames,
			Delay:           im.Delay,
			LoopCount:       im.LoopCount,
			Disposal:        im.Disposal,
			Config:          im.Config,
			BackgroundIndex: im.BackgroundIndex,
		})

		if err != nil {
			wi.Logger.Error().
				Err(err).
				Msg("Failed to generate GIF")

			return "", err
		}

		wi.Logger.Debug().
			Int("frames", len(frames)).
			Int64("qms", quant_end.Round(time.Millisecond).Milliseconds()).
			Int64("ms", time.Since(start).Round(time.Millisecond).Milliseconds()).
			Msg("Generated GIF")

		return "gif", nil
	}

	err = png.Encode(b, frames[0])
	if err != nil {
		wi.Logger.Error().
			Err(err).
			Msg("Failed to generate PNG")

		return "", err
	}

	wi.Logger.Debug().
		Int64("ms", time.Since(start).Round(time.Millisecond).Milliseconds()).
		Msg("Generated PNG")

	return "png", err
}

// CreateFontPack creates a pack of fonts with fallback and the one passed.
func (wi *WelcomerImageService) CreateFontPack(font string, size float64) *multiface.Face {
	face := new(multiface.Face)

	f, fo, err := wi.FetchFont(font, size)
	if err != nil {
		wi.Logger.Warn().Err(err).Str("font", font).Msg("Failed to fetch font in font pack")
	} else {
		face.AddTruetypeFace(*f.Face, fo.Font)
	}

	for _, fallback := range wi.FallbackFonts {
		f, fo, err := wi.FetchFont(fallback, size)
		if err != nil {
			wi.Logger.Warn().Err(err).Str("font", font).Msg("Failed to fetch fallback font in font pack")
		} else {
			face.AddTruetypeFace(*f.Face, fo.Font)
		}
	}

	return face
}

// CreateFontPackHook returns a newFace function with an argument.
func (wi *WelcomerImageService) CreateFontPackHook(f string) func(float64) font.Face {
	return func(i float64) font.Face {
		return wi.CreateFontPack(f, i)
	}
}

// DrawMultiline draws text using multilines and adds stroke.
func DrawMultiline(d font.Drawer, newFace func(float64) font.Face, args MultilineArguments) error {
	lines := strings.Split(args.Text, "\n")
	if len(lines) == 0 {
		return nil
	}

	args.X += args.StrokeWeight
	args.Y += args.StrokeWeight

	args.Width -= args.StrokeWeight * 2
	args.Height -= args.StrokeWeight * 2

	var ls string

	var ll int

	d.Face = newFace(args.DefaultFontSize)

	// Calculate the widest line so we can use it as a baseline for other
	// font size calculations.
	for _, l := range lines {
		_, adv := d.BoundString(l)
		if adv.Ceil() > ll {
			ls = l
			ll = adv.Ceil()
		}
	}

	s := args.DefaultFontSize

	// If the widest line does not fit or the total line height is larger
	// than the height, we need to decrease the font size.
	if (ll > args.Width) || ((d.Face.Metrics().Height.Ceil() * len(lines)) > args.Height) {
		// Keep decreasing the font size until we can fit the width and height
		for {
			face := newFace(s)

			_, adv := font.BoundString(face, ls)

			// We will keep decreasing size until it fits, unless the font is already 1.
			// It may not fit but its likely the text will never fit after that point.
			if ((adv.Ceil() <= args.Width) && ((face.Metrics().Height.Ceil() * len(lines)) <= args.Height)) || s <= 1 {
				d.Face = face

				break
			}

			// TODO: Scale aware resizing so we do less operations
			s = float64(s) * fontResize
		}
	}

	fa := d.Face.Metrics().Ascent.Ceil()
	fh := d.Face.Metrics().Height.Ceil()
	th := len(lines) * fh

	sc := image.NewUniform(args.StrokeColour)
	tc := image.NewUniform(args.TextColour)

	for lineNo, l := range lines {
		_, adv := d.BoundString(l)

		var Dx int

		var Dy int

		switch args.HorizontalAlignment {
		case AlignLeft:
			Dx = 0
		case AlignMiddle:
			Dx = int((args.Width - adv.Ceil()) / 2)
		case AlignRight:
			Dx = args.Width - adv.Ceil()
		}

		switch args.VerticalAlignment {
		case AlignTop:
			Dy = lineNo * fh
		case AlignCenter:
			Dy = (lineNo * fh) + (args.Height / 2) - (th / 2)
		case AlignBottom:
			Dy = args.Height - th + (lineNo * fh)
		}

		Dx = args.X + Dx
		Dy = args.Y + Dy

		if args.StrokeWeight > 0 {
			p := args.StrokeWeight * args.StrokeWeight

			for dy := -args.StrokeWeight; dy <= args.StrokeWeight; dy++ {
				for dx := -args.StrokeWeight; dx <= args.StrokeWeight; dx++ {
					// Round out stroke
					if dx*dx+dy*dy >= p {
						continue
					}

					d.Dot = fixed.P(Dx+dx, Dy+dy+fa)
					d.Src = sc
					d.DrawString(l)
				}
			}
		}

		d.Dot = fixed.P(Dx, Dy+fa)
		d.Src = tc

		d.DrawString(l)
	}

	return nil
}

// GenerateAvatar applies masking and resizing to the avatar.
// Outputs an image.Image with same dimension as src image.
func (wi *WelcomerImageService) GenerateAvatar(avatar *StaticImageCache, imageOpts ImageOpts) (image.Image, error) {
	cropPix := int(math.Floor(float64(avatar.Image.Bounds().Dx()) / 8))

	atlas := image.NewRGBA(image.Rect(
		0, 0,
		avatar.Image.Bounds().Dx()+(imageOpts.ProfileBorderWidth*2),
		avatar.Image.Bounds().Dy()+(imageOpts.ProfileBorderWidth*2),
	))

	context := gg.NewContextForRGBA(atlas)

	context.SetColor(imageOpts.ProfileBorderColour)
	context.Clear()

	rounding := float64(0)

	var avatarImage image.Image

	switch imageOpts.ProfileBorderCurve {
	case CurveCircle:
		rounding = 1000

		canCrop := (avatar.Image.At(
			cropPix,
			cropPix,
		) == transparent &&
			avatar.Image.At(
				avatar.Image.Bounds().Dx()-cropPix,
				avatar.Image.Bounds().Dy()-cropPix,
			) == transparent)

		if canCrop {
			avatarMinimimze := image.NewRGBA(avatar.Image.Bounds())
			avatarContext := gg.NewContextForRGBA(avatarMinimimze)

			avatarContext.DrawImage(
				imaging.Resize(
					avatar.Image,
					(avatar.Image.Bounds().Dx()-(cropPix*2)),
					(avatar.Image.Bounds().Dx()-(cropPix*2)),
					imaging.Lanczos,
				),
				cropPix,
				cropPix,
			)

			avatarImage = roundImage(avatarMinimimze, 1000)
		} else {
			avatarImage = roundImage(avatar.Image, 1000)
		}
	case CurveSoft:
		rounding = 16
		avatarImage = roundImage(avatar.Image, 8)
	case CurveSquare:
		avatarImage = avatar.Image
	}

	context.DrawImageAnchored(
		avatarImage,
		context.Width()/2,
		context.Height()/2,
		0.5,
		0.5,
	)

	return roundImage(atlas, rounding), nil
}

// GenerateImage generates an Image.
func (wi *WelcomerImageService) GenerateImage(b *bytes.Buffer, imageOpts ImageOpts) (string, error) {
	bench := NewBench()

	bench.Add("init", time.Now().UTC())

	// Fetch Theme

	themesMu.RLock()
	theme, ok := themes[imageOpts.Theme]
	themesMu.RUnlock()

	if !ok {
		theme = themes[ThemeRegular]
	}

	bench.Add("fetch theme", time.Now().UTC())

	// Fetch Background
	background, err := wi.FetchBackground(
		imageOpts.Background,
		imageOpts.AllowGIF,
	)
	if err != nil {
		wi.Logger.Error().Err(err).
			Str("background", imageOpts.Background).
			Bool("allow_gif", imageOpts.AllowGIF).
			Msg("Failed to fetch background")

		return "", err
	}

	bench.Add("fetch background", time.Now().UTC())

	// Fetch Avatar
	avatar, err := wi.FetchAvatar(
		imageOpts.UserId,
		imageOpts.Avatar,
	)
	if err != nil {
		wi.Logger.Warn().Err(err).
			Int64("user", imageOpts.UserId).
			Str("avatar", imageOpts.Avatar).
			Msg("Failed to fetch avatar")

		return "", err
	}

	bench.Add("fetch avatar", time.Now().UTC())

	// Convert Avatar to Proper Shape
	avatarOverlay, err := wi.GenerateAvatar(
		avatar,
		imageOpts,
	)
	if err != nil {
		wi.Logger.Error().Err(err).
			Msg("Failed to generate avatar")

		return "", err
	}

	bench.Add("avatar convert", time.Now().UTC())

	// Create overlay image
	themeArgs := GenerateImageArgs{
		ImageOpts: imageOpts,
		Avatar:    avatarOverlay,
	}

	themeResp, err := theme(wi, b, themeArgs)
	if err != nil {
		wi.Logger.Error().Err(err).
			Msg("Failed to generate overlay")
	}

	bench.Add("generate overlay", time.Now().UTC())

	// Create border if required
	if imageOpts.BorderWidth > 0 {
		border := image.Point{imageOpts.BorderWidth, imageOpts.BorderWidth}
		d := border.Add(border)

		// Increases size and adds offset to TargetImageSize
		themeResp.TargetImageW += d.X
		themeResp.TargetImageH += d.Y
		themeResp.TargetImageSize.Max = themeResp.TargetImageSize.Max.Add(d)

		borderOverlay := image.NewRGBA(themeResp.TargetImageSize)

		context := gg.NewContextForRGBA(borderOverlay)

		context.SetColor(imageOpts.BorderColour)

		// top
		context.DrawRectangle(
			0,
			0,
			float64(themeResp.TargetImageW),
			float64(border.X),
		)

		// right
		context.DrawRectangle(
			float64(themeResp.TargetImageW-border.X),
			float64(border.Y),
			float64(border.X),
			float64(themeResp.TargetBackgroundW-(border.Y*2)),
		)

		// bottom
		context.DrawRectangle(
			0,
			float64(themeResp.TargetImageH-border.Y),
			float64(themeResp.TargetImageW),
			float64(border.Y),
		)

		// left
		context.DrawRectangle(
			0,
			float64(border.Y),
			float64(border.X),
			float64(themeResp.TargetImageH-(border.Y*2)),
		)

		context.Fill()

		context.DrawImage(
			themeResp.Overlay,
			border.X+themeResp.OverlayAnchor.X,
			border.Y+themeResp.OverlayAnchor.Y,
		)

		themeResp.Overlay = borderOverlay

		themeResp.OverlayAnchor = image.Point{}
		themeResp.BackgroundAnchor = themeResp.BackgroundAnchor.Add(border)

		bench.Add("generate border", time.Now().UTC())
	}

	// Resize frames and overlay with overlay
	frames := background.GetFrames()
	wg := sync.WaitGroup{}

	for frameNumber, frame := range frames {
		wg.Add(1)

		go func(frameNumber int, frame image.Image) {
			resizedFrame := image.NewRGBA(themeResp.TargetImageSize)

			// Draw resized background frame
			draw.Draw(
				resizedFrame, resizedFrame.Rect.Add(themeResp.BackgroundAnchor),
				imaging.Fill(
					frame,
					themeResp.TargetBackgroundW, themeResp.TargetBackgroundH,
					imaging.Center, imaging.Lanczos,
				),
				image.Point{}, draw.Src,
			)

			// Draw overlay ontop
			draw.Draw(
				resizedFrame, resizedFrame.Rect.Add(themeResp.OverlayAnchor),
				themeResp.Overlay,
				image.Point{}, draw.Over,
			)

			frames[frameNumber] = resizedFrame

			wg.Done()
		}(frameNumber, frame)
	}

	wg.Wait()

	bench.Add("overlay frames", time.Now().UTC())

	background.Config = image.Config{
		Width:      themeResp.TargetImageW,
		Height:     themeResp.TargetImageH,
		ColorModel: nil,
	}

	background.Disposal = nil

	// Encode final image
	format, err := wi.EncodeImages(b, frames, background)
	if err != nil {
		wi.Logger.Error().Err(err).
			Msg("Failed to encode image")
	}

	bench.Add("generate images", time.Now().UTC())

	bench.Print()

	return format, err
}
