package welcomerimages

import (
	"image"
	"image/color"
	"sync"
	"time"

	"golang.org/x/image/font"
	"golang.org/x/image/font/sfnt"
)

type (
	Xalignment uint8
	Yalignment uint8
)

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

	// Include a border around profile pictures. This also fills
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

// MultilineArguments is a list of arguments for the DrawMultiline function.
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

// LastAccess stores the last access of the structure.
type LastAccess struct {
	sync.RWMutex   // Used to stop deletion whilst being used
	LastAccessed   time.Time
	LastAccessedMu sync.RWMutex
}

type ImageData struct {
	ID        string    `json:"id" msgpack:"i"`
	GuildID   int64     `json:"guild_id" msgpack:"g"`
	Size      int       `json:"size" msgpack:"s"`
	Path      string    `json:"path" msgpack:"p"`
	ExpiresAt time.Time `json:"expires_at" msgpack:"e"`
	CreatedAt time.Time `json:"created_at" msgpack:"c"`

	isDefault bool
}

// FontCache stores the Font, last accessed and Faces for different sizes.
type FontCache struct {
	LastAccessedMu sync.RWMutex
	LastAccessed   time.Time

	Font        *sfnt.Font
	FaceCacheMu sync.RWMutex
	FaceCache   map[float64]*FaceCache
}

// FaceCache stores the Face and when it was last accessed.
type FaceCache struct {
	LastAccess

	Face *font.Face
}

// FileCache stores the file body and when it was last accessed.
type FileCache struct {
	LastAccess

	Filename string
	Ext      string
	Path     string
	Body     []byte
}

// StaticImageCache stores just an image.
type StaticImageCache struct {
	LastAccess

	Format string
	Image  image.Image
}

// RequestCache stores the request body and when it was last accessed.
type RequestCache struct {
	LastAccess

	URL  string
	Body []byte
}

// ImageCache stores the image and the extension for it.
type ImageCache struct {
	LastAccess

	// The image format that is represented
	Format string

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

// GetFrames returns a copy of the ImageCache.Frames.
func (ic *ImageCache) GetFrames() []image.Image {
	im := make([]image.Image, len(ic.Frames))
	copy(im, ic.Frames)

	return im
}
