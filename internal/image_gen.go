package welcomerimages

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/png"
	"sync"
	"time"

	"github.com/savsgio/gotils"
	"github.com/ultimate-guitar/go-imagequant"
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

var attr, _ = imagequant.NewAttributes()

func init() {
	attr.SetSpeed(5)
}

// quantizeImage converts an image.Image to image.Paletted via imagequant
func quantizeImage(im image.Image) *image.Paletted {
	b := im.Bounds()

	qimg, err := imagequant.NewImage(attr, gotils.B2S(imagequant.ImageToRgba32(im)), b.Dx(), b.Dy(), 1)
	if err != nil {
		panic(err)
	}

	pm, err := qimg.Quantize(attr)
	if err != nil {
		panic(err)
	}

	np := image.NewPaletted(im.Bounds(), pm.GetPalette())
	draw.Draw(np, np.Rect, im, im.Bounds().Min, draw.Over)
	// TODO: Figure out how to speed this Draw up

	pm.Release()
	qimg.Release()

	return np
}

// GenerateImage generates an Image
func (wi *WelcomerImageService) GenerateImage(b *bytes.Buffer, imageOpts ImageOpts) (string, error) {

	// a, err := wi.FetchAvatar(imageOpts.UserId, imageOpts.Avatar)
	// if err != nil {
	// 	wi.Logger.Error().Err(err).Msg("Failed to fetch avatar")
	// 	return "", err
	// }

	// _, err = b.Write(a)
	// if err != nil {
	// 	wi.Logger.Error().Err(err).Msg("Failed to write to buffer")
	// }

	a, err := wi.FetchBackground(imageOpts.Background, imageOpts.AllowGIF)
	if err != nil {
		wi.Logger.Error().Err(err).Msg("Failed to fetch avatar")
		return "", err
	}

	start := time.Now()
	if len(a.Frames) > 1 && imageOpts.AllowGIF {
		wg := sync.WaitGroup{}
		_frames := make([]*image.Paletted, len(a.Frames), len(a.Frames))
		for c, frame := range a.Frames {
			wg.Add(1)
			go func(im image.Image, jobnum int) {
				_frames[jobnum] = quantizeImage(im)
				wg.Done()
			}(frame, c)
		}
		wg.Wait()

		gif.EncodeAll(b, &gif.GIF{
			Image:           _frames,
			Delay:           a.Delay,
			LoopCount:       a.LoopCount,
			Disposal:        a.Disposal,
			Config:          a.Config,
			BackgroundIndex: a.BackgroundIndex,
		})

		wi.Logger.Debug().
			Int("frames", len(a.Frames)).
			Int64("ms", time.Now().Sub(start).Round(time.Millisecond).Milliseconds()).
			Msg("Generated GIF")
	} else {
		png.Encode(b, a.Frames[0])

		wi.Logger.Debug().
			Int64("ms", time.Now().Sub(start).Round(time.Millisecond).Milliseconds()).
			Msg("Generated PNG")
	}

	return ".png", nil
}
