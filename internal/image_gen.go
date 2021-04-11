package welcomerimages

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/png"
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

// EncodeImages encodes a list of []image.Image and takes an input of AllowGifs.
// Outputs the file extension.
func (wi *WelcomerImageService) EncodeImages(b *bytes.Buffer, frames []image.Image, im *ImageCache, allowGIF bool) (string, error) {
	var err error

	start := time.Now().UTC()

	if frames == nil {
		frames = im.Frames
	}

	if len(frames) == 0 {
		return "", xerrors.New("empty frame list")
	}

	if len(frames) > 1 && allowGIF {
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
		case Left:
			Dx = 0
		case Middle:
			Dx = int((args.Width - adv.Ceil()) / 2)
		case Right:
			Dx = args.Width - adv.Ceil()
		}

		switch args.VerticalAlignment {
		case Top:
			Dy = lineNo * fh
		case Center:
			Dy = (lineNo * fh) + (args.Height / 2) - (th / 2)
		case Bottom:
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

// GenerateImage generates an Image.
func (wi *WelcomerImageService) GenerateImage(b *bytes.Buffer, imageOpts ImageOpts) (string, error) {
	bench := NewBench()

	bench.Add("start", time.Now().UTC())

	// Create profile
	bg, err := wi.FetchBackground(imageOpts.Background, imageOpts.AllowGIF)
	if err != nil {
		wi.Logger.Error().Err(err).Msg("Failed to fetch avatar")

		return "", err
	}

	bench.Add("fetch background", time.Now().UTC())

	avatar, err := wi.FetchAvatar(imageOpts.UserId, imageOpts.Avatar)
	if err != nil {
		wi.Logger.Error().Err(err).Msg("Failed to fetch avatar")

		return "", err
	}

	bench.Add("fetch avatar", time.Now().UTC())

	borderWidth := 16

	// Target canvas size for image and background.
	// Difference with canvas means there is a border
	targetWidth := 1000
	targetHeight := 300

	// Total width and height of image
	canvasWidth := 1000 + (borderWidth * 2) // 1064
	canvasHeight := 300 + (borderWidth * 2) // 364

	// Prepare border
	hasBorder := (targetWidth != canvasWidth) || (targetHeight != canvasHeight)
	borderColour := color.NRGBA{255, 255, 255, 255}
	canvas := image.NewRGBA(image.Rect(0, 0, canvasWidth, canvasHeight))
	drawer := font.Drawer{Dst: canvas}
	context := gg.NewContextForRGBA(canvas)

	bench.Add("init", time.Now().UTC())

	if hasBorder {
		context.SetColor(borderColour)
		context.DrawRectangle(
			0, 0,
			float64(canvasWidth), float64(borderWidth),
		)
		context.DrawRectangle(
			0, float64(canvasHeight-borderWidth),
			float64(canvasWidth), float64(canvasHeight),
		)
		context.DrawRectangle(
			0, float64(borderWidth),
			float64(borderWidth), float64(canvasHeight-borderWidth),
		)
		context.DrawRectangle(
			float64(canvasWidth-borderWidth), float64(borderWidth),
			float64(canvasWidth), float64(canvasHeight-borderWidth),
		)
	}

	bench.Add("border", time.Now().UTC())

	DrawMultiline(drawer, wi.CreateFontPackHook(imageOpts.Font), MultilineArguments{
		DefaultFontSize: defaultFontSize,

		X: 268 + 32 + borderWidth,
		Y: 0 + 32 + borderWidth,

		Width:  668,
		Height: 236,

		HorizontalAlignment: Left,
		VerticalAlignment:   Center,

		StrokeWeight: 3,
		StrokeColour: color.Black,
		TextColour:   color.White,

		Text: "Welcome ImRock\nto the Welcomer Support Guild\nyou are the 5588th member!",
	})

	bench.Add("draw text", time.Now().UTC())

	bounds := avatar.Image.Bounds()

	// If appropriate, add extra padding to profile pictures that like will have corners cut off a little.
	if avatar.Image.At(transAnchor, transAnchor) == image.Transparent &&
		avatar.Image.At(bounds.Dx()-transAnchor, bounds.Dy()-transAnchor) == image.Transparent {
		avatar.Image = imaging.PasteCenter(
			imaging.New(
				avatarSize+transAnchor,
				avatarSize+transAnchor,
				color.Transparent,
			),
			avatar.Image,
		)
	}

	// Resize image, round it, add background behind image then overlay
	avatarr := imaging.Resize(
		roundImage(
			avatar.Image,
			1000,
		),
		204, 204, imaging.Lanczos)

	context.SetColor(image.NewUniform(color.NRGBA{255, 255, 255, 255}))
	context.DrawCircle(float64(118+32+borderWidth), float64(118+32+borderWidth), 118)
	context.Fill()

	context.DrawImage(
		avatarr,
		(118-(avatarr.Rect.Dx()/2))+32+borderWidth,
		(118-(avatarr.Rect.Dy()/2))+32+borderWidth,
	)

	bench.Add("avatar", time.Now().UTC())

	backgroundPt := image.Pt(borderWidth, borderWidth)

	frames := bg.GetFrames()
	wg := sync.WaitGroup{}

	bench.Add("get frames", time.Now().UTC())

	for i, frame := range frames {
		wg.Add(1)

		go func(i int, frame image.Image) {
			rframe := image.NewNRGBA(image.Rect(0, 0, int(canvasWidth), int(canvasHeight)))

			draw.Draw(
				rframe, rframe.Bounds().Add(backgroundPt),
				imaging.Fill(frame, targetWidth, targetHeight, imaging.Center, imaging.NearestNeighbor),
				image.Point{}, draw.Src)

			draw.Draw(rframe, rframe.Bounds(), canvas, image.Point{}, draw.Over)

			frames[i] = rframe

			wg.Done()
		}(i, frame)
	}

	bench.Add("overlay", time.Now().UTC())

	wg.Wait()

	bg.Config.Width = canvasWidth
	bg.Config.Height = canvasHeight
	bg.Disposal = nil
	bg.Config.ColorModel = nil

	defer func() {
		bench.Add("encode", time.Now().UTC())
		bench.Print()
	}()

	return wi.EncodeImages(b, frames, bg, imageOpts.AllowGIF)
}
