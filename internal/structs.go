package welcomerimages

import (
	"image"
	"image/color"
	"sync"
	"time"

	"golang.org/x/image/font"
	"golang.org/x/image/font/sfnt"
)

const bitmask = 255

type (
	Xalignment         uint8
	Yalignment         uint8
	ProfileAlignment   uint8
	ProfileBorderCurve uint8
	Theme              uint8
)

const (
	AlignLeft Xalignment = iota
	AlignMiddle
	AlignRight
)

const (
	AlignTop Yalignment = iota
	AlignCenter
	AlignBottom
)

const (
	FloatLeft ProfileAlignment = iota
	FloatRight
)

const (
	CurveCircle ProfileBorderCurve = iota
	CurveSoft
	CurveSquare
)

const (
	ThemeRegular Theme = iota
	ThemeBadge
	ThemeVertical
)

type ImageOpts struct {
	// Newline split message
	Text string `json:"text"`

	GuildId int64 `json:"guild_id"`

	UserId int64  `json:"user_id"`
	Avatar string `json:"avatar"`

	AllowGIF bool `json:"allow_gif"`

	// Which theme to use when generating images
	Theme Theme `json:"layout"`

	// Identifier for background
	Background string `json:"background"`

	// Identifier for font to use (along with Noto)
	Font string `json:"font"`

	// Border applied to entire image. If transparent, there is no border.
	BorderColour    color.RGBA `json:"-"`
	BorderColourHex string     `json:"border_colour"`
	BorderWidth     int        `json:"border_width"`

	// Alignment of left or right (assuming not vertical layout)
	ProfileAlignment ProfileAlignment `json:"profile_alignment"`

	// Text alignment (left, center, right) (top, middle, bottom)
	TextAlignmentX Xalignment `json:"text_alignment_x"`
	TextAlignmentY Yalignment `json:"text_alignment_y"`

	// Include a border around profile pictures. This also fills
	// under the profile.
	ProfileBorderColour    color.RGBA `json:"-"`
	ProfileBorderColourHex string     `json:"profile_border_colour"`
	// Padding applied to profile pictures inside profile border
	ProfileBorderWidth int `json:"profile_border_width"`
	// Type of curving on the profile border (circle, rounded, square)
	ProfileBorderCurve ProfileBorderCurve `json:"profile_border_curve"`

	// Text stroke. If 0, there is no stroke
	TextStroke          int        `json:"text_stroke"`
	TextStrokeColour    color.RGBA `json:"-"`
	TextStrokeColourHex string     `json:"text_stroke_colour"`

	TextColour    color.RGBA `json:"-"`
	TextColourHex string     `json:"text_colour"`
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

type GenerateImageArgs struct {
	ImageOpts ImageOpts

	// Avatar with mask and background pre-applied
	Avatar image.Image
}

type GenerateThemeResp struct {
	// Overlay
	Overlay image.Image

	// The target size of entire image
	TargetImageSize            image.Rectangle
	TargetImageW, TargetImageH int

	// The target size of backgrounds. This is
	// equal to TargetImage however changes if
	// there is a border.
	TargetBackgroundSize                 image.Rectangle
	TargetBackgroundW, TargetBackgroundH int

	// Point to move from (0,0) when
	// rendering the backgrounds
	BackgroundAnchor image.Point

	// Point to move from (0,0) when
	// rendering the overlay
	OverlayAnchor image.Point
}
