package welcomerimages

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/png"
	"math"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/WelcomerTeam/WelcomerImages/pkg/multiface"
	"github.com/disintegration/imaging"
	"github.com/fogleman/gg"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	gotils "github.com/savsgio/gotils/strconv"
	"github.com/ultimate-guitar/go-imagequant"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
	"golang.org/x/xerrors"
)

type Xalignment uint8
type Yalignment uint8

const (
	Left Xalignment = iota
	Middle
	Right
)

const (
	Top Yalignment = iota
	Center
	Bottom
)

type ImageOpts struct {
	// Newline split message
	Text string `json:"text"`

	GuildId int64 `json:"guild_id"`

	UserId int64  `json:"user_id"`
	Avatar string `json:"avatar"`

	AllowGIF bool `json:"allow_gif"`

	// Which layout to use when generating images
	Layout int `json:"layout"` // todo: type

	// Identifier for background
	Background string `json:"background"`

	// Identifier for font to use (along with Noto)
	Font string `json:"font"`

	// Border applied to entire image. If transparent, there is no border.
	BorderColour color.Color `json:"border_colour"`
	BorderWidth  int         `json:"border_width"`

	// Alignment of left or right (assuming not vertical layout)
	ProfileAlignment int `json:"profile_alignment"` // todo: type

	// Include a border arround profile pictures. This also fills
	// under the profile.
	ProfileBorderColour color.Color `json:"profile_border_colour"`
	// Padding applied to profile pictures inside profile border
	ProfileBorderWidth int `json:"profile_border_width"`
	// Type of curving on the profile border (square, circle, rounded)
	ProfileBorderCurve int `json:"profile_border_curve"` // todo: type

	// Text stroke. If 0, there is no stroke
	TextStroke       int         `json:"text_stroke"`
	TextStrokeColour color.Color `json:"text_stroke_colour"`

	TextColour color.Color `json:"text_colour"`
}

// quantizeImage converts an image.Image to image.Paletted via imagequant
func quantizeImage(l zerolog.Logger, src image.Image) (*image.Paletted, error) {
	b := src.Bounds()

	qimg, err := imagequant.NewImage(attr, gotils.B2S(imagequant.ImageToRgba32(src)), b.Dx(), b.Dy(), 1)
	if err != nil {
		l.Panic().Err(err).Msg("Failed to create new imagequant image")
	}

	pm, err := qimg.Quantize(attr)
	if err != nil {
		f, _ := os.Create(uuid.New().String() + ".png")
		png.Encode(f, src)
		f.Close()
		l.Panic().Err(err).Msg("Failed to quantize image")
	}

	dst := image.NewPaletted(src.Bounds(), pm.GetPalette())

	// WriteRemappedImage returns a list of bytes pointing to direct
	// palette indexes so we can just copy it over and it will be
	// using the optimimal indexes.
	rmap, err := pm.WriteRemappedImage()
	if err != nil {
		return dst, err
	}

	dst.Pix = rmap
	// copy(dst.Pix, rmap)

	pm.Release()
	qimg.Release()

	return dst, nil
}

// EncodeImages encodes a list of []image.Image and takes an input of AllowGifs.
// Outputs the file extention.
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

// CreateFontPack creates a pack of fonts with fallback and the one passed
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

// MultilineArguments is a list of arguments for the DrawMultiline function
type MultilineArguments struct {
	DefaultFontSize float64 // default font size to start with

	X int
	Y int

	Width  int
	Height int

	HorizontalAlignment Xalignment
	VerticalAlignment   Yalignment

	StrokeWeight int
	StrokeColour color.Color
	TextColour   color.Color

	Text string
}

// CreateFontPackHook returns a newFace function with an argument
func (wi *WelcomerImageService) CreateFontPackHook(f string) func(float64) font.Face {
	return func(i float64) font.Face {
		return wi.CreateFontPack(f, i)
	}
}

// DrawMultiline draws text using multilines and adds stroke
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
			s = float64(s) * 0.90
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

// roundImage cuts out a rounded segment from an image
func roundImage(im image.Image, r float64) image.Image {
	b := im.Bounds()
	r = math.Max(0, math.Min(r, float64(b.Dy())/2))
	context := gg.NewContext(b.Dx(), b.Dy())
	context.DrawRoundedRectangle(0, 0, float64(b.Dx()), float64(b.Dy()), r)
	context.Clip()
	context.DrawImage(im, 0, 0)

	return context.Image()
}

// GenerateImage generates an Image
func (wi *WelcomerImageService) GenerateImage(b *bytes.Buffer, imageOpts ImageOpts) (string, error) {
	// defer func() {
	// 	recover()
	// }()

	a := time.Now().UTC()
	// Create profile
	bg, err := wi.FetchBackground(imageOpts.Background, imageOpts.AllowGIF)
	if err != nil {
		wi.Logger.Error().Err(err).Msg("Failed to fetch avatar")
		return "", err
	}

	println("FETCH BG", time.Since(a).Round(time.Millisecond).Milliseconds())
	a = time.Now().UTC()

	avatar, err := wi.FetchAvatar(imageOpts.UserId, imageOpts.Avatar)
	if err != nil {
		wi.Logger.Error().Err(err).Msg("Failed to fetch avatar")
		return "", err
	}

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

	println("FETCH AVA", time.Since(a).Round(time.Millisecond).Milliseconds())
	canvas := image.NewRGBA(image.Rect(0, 0, canvasWidth, canvasHeight))
	drawer := font.Drawer{Dst: canvas}

	context := gg.NewContextForRGBA(canvas)

	if hasBorder {
		context.SetColor(borderColour)
		context.DrawRectangle(0, 0, float64(canvasWidth), float64(borderWidth))
		context.DrawRectangle(0, float64(canvasHeight-borderWidth), float64(canvasWidth), float64(canvasHeight))
		context.DrawRectangle(0, float64(borderWidth), float64(borderWidth), float64(canvasHeight-borderWidth))
		context.DrawRectangle(float64(canvasWidth-borderWidth), float64(borderWidth), float64(canvasWidth), float64(canvasHeight-borderWidth))
	}

	a = time.Now().UTC()
	DrawMultiline(drawer, wi.CreateFontPackHook(imageOpts.Font), MultilineArguments{
		DefaultFontSize: 96,

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

	bounds := avatar.Image.Bounds()

	// If appropriate, add extra padding to profile pictures that like will have corners cut off a little.
	if avatar.Image.At(32, 32) == image.Transparent && avatar.Image.At(bounds.Dx()-32, bounds.Dy()-32) == image.Transparent {
		avatar.Image = imaging.PasteCenter(
			imaging.New(256+32, 256+32, color.Transparent),
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

	backgroundPt := image.Pt(borderWidth, borderWidth)

	a = time.Now().UTC()

	fi, _ := os.Create("overlay.png")
	png.Encode(fi, canvas)
	fi.Close()

	frames := bg.GetFrames()
	wg := sync.WaitGroup{}
	for i, frame := range frames {
		wg.Add(1)
		go func(i int, frame image.Image) {
			rframe := image.NewNRGBA(image.Rect(0, 0, int(canvasWidth), int(canvasHeight)))

			a := time.Now()
			draw.Draw(
				rframe, rframe.Bounds().Add(backgroundPt),
				imaging.Fill(frame, targetWidth, targetHeight, imaging.Center, imaging.NearestNeighbor),
				image.Point{}, draw.Src)

			b := time.Now()

			draw.Draw(rframe, rframe.Bounds(), canvas, image.Point{}, draw.Over)

			println(time.Since(a).Milliseconds(), time.Since(b).Milliseconds())
			frames[i] = rframe
			wg.Done()
		}(i, frame)
	}
	wg.Wait()

	println("FIT", time.Since(a).Round(time.Millisecond).Milliseconds())

	bg.Config.Width = canvasWidth
	bg.Config.Height = canvasHeight
	bg.Disposal = nil
	bg.Config.ColorModel = nil
	return wi.EncodeImages(b, frames, bg, imageOpts.AllowGIF)
}
