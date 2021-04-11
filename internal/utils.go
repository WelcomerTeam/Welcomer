package welcomerimages

import (
	"encoding/hex"
	"image"
	"image/color"
	"image/gif"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fogleman/gg"
	"github.com/rs/zerolog"
	gotils "github.com/savsgio/gotils/strconv"
	"github.com/ultimate-guitar/go-imagequant"
)

const (
	rgbLength  = 6
	argbLength = 8
)

var (
	colorWhite = color.RGBA{255, 255, 255, 255}
	colorBlack = color.RGBA{0, 0, 0, 255}
)

// fsExists checks if a file or folder exists and returns if it does.
func fsExists(path string) bool {
	_, err := os.Stat(path)

	return !os.IsNotExist(err)
}

// openImage returns an image, format and error.
func openImage(path string) (image.Image, *gif.GIF, image.Config, string, error) {
	fi, err := os.Open(path)
	if err != nil {
		return nil, nil, image.Config{}, "", err
	}

	if strings.HasSuffix(path, ".gif") {
		g, err := gif.DecodeAll(fi)
		if err != nil {
			return nil, nil, image.Config{}, "", err
		}

		return nil, g, g.Config, "gif", nil
	} else {
		i, f, err := image.Decode(fi)
		if err != nil {
			return nil, nil, image.Config{}, f, err
		}

		return i, nil, image.Config{
			Width:  i.Bounds().Dx(),
			Height: i.Bounds().Dy(),
		}, f, nil
	}
}

// quantizeImage converts an image.Image to image.Paletted via imagequant.
func quantizeImage(l zerolog.Logger, src image.Image) (*image.Paletted, error) {
	b := src.Bounds()

	qimg, err := imagequant.NewImage(attr, gotils.B2S(imagequant.ImageToRgba32(src)), b.Dx(), b.Dy(), 1)
	if err != nil {
		l.Panic().Err(err).Msg("Failed to create new imagequant image")
	}

	pm, err := qimg.Quantize(attr)
	if err != nil {
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

// roundImage cuts out a rounded segment from an image.
func roundImage(im image.Image, r float64) image.Image {
	b := im.Bounds()
	r = math.Max(0, math.Min(r, float64(b.Dy())/2))
	context := gg.NewContext(b.Dx(), b.Dy())
	context.DrawRoundedRectangle(0, 0, float64(b.Dx()), float64(b.Dy()), r)
	context.Clip()
	context.DrawImage(im, 0, 0)

	return context.Image()
}

func fitTo(s string, l int) string {
	return s + strings.Repeat(" ", l-len(s))
}

// Debugging tool showing timings between steps.
type Bench struct {
	sync.RWMutex
	benches map[int]time.Time
	labels  map[int]string
}

func NewBench() *Bench {
	return &Bench{
		benches: make(map[int]time.Time),
		labels:  make(map[int]string),
	}
}

func (b *Bench) Add(l string, t time.Time) {
	b.Lock()
	defer b.Unlock()

	bn := len(b.benches)
	b.benches[bn] = t
	b.labels[bn] = l
}

func (b *Bench) Print() {
	b.RLock()
	b.RUnlock()

	lw := 6
	for _, l := range b.labels {
		if len(l) > lw {
			lw = len(l)
		}
	}

	var t int64

	var tt int64

	println(fitTo("Label", lw) + " | Dur     | Elapsed")
	println(strings.Repeat("-", lw) + " | ------- | -------")

	for bn := 0; bn < len(b.labels); bn++ {
		l := b.labels[bn]
		d := b.benches[bn]

		if bn > 0 {
			t = d.Sub(b.benches[bn-1]).Round(time.Millisecond).Milliseconds()
		} else {
			t = 0
		}

		tt += t

		println(
			fitTo(l, lw) +
				" | " +
				fitTo(strconv.FormatInt(t, 10), 5) +
				"ms | " +
				fitTo(strconv.FormatInt(tt, 10), 5) +
				"ms",
		)
	}

	println(strings.Repeat("-", lw) + " | ------- | -------")
	println("Total time taken: " + strconv.FormatInt(tt, 10) + "ms")
}

// converts #AARRGGBB to color.RGBA format.
func convertARGB(input string, d color.RGBA) color.RGBA {
	input = strings.ReplaceAll(input, "#", "")

	switch len(input) {
	case rgbLength:
		h, err := hex.DecodeString(input)
		if err != nil || len(h) != 3 {
			return d
		}

		return color.RGBA{h[0], h[1], h[2], 255}
	case argbLength:
		h, err := hex.DecodeString(input)
		if err != nil || len(h) != 4 {
			return d
		}

		return color.RGBA{h[1], h[2], h[3], h[0]}
	default:
		return d
	}
}
