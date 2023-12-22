package service

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/gif"

	core "github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/gofrs/uuid"
)

const (
	hexBase      = 16
	int64Base    = 10
	int64BitSize = 64

	SolidProfileLuminance = 0.7
)

func (is *ImageService) FetchBackground(background string, allowAnimated bool, avatar image.Image) (*FullImage, error) {
	backgroundType, _ := core.ParseBackground(background)

	switch backgroundType.Type {
	case core.BackgroundTypeWelcomer:
		return is.FetchBackgroundWelcomer(backgroundType.Value, allowAnimated)
	case core.BackgroundTypeSolid:
		return is.FetchBackgroundSolid(backgroundType.Value)
	case core.BackgroundTypeSolidProfile:
		return is.FetchBackgroundSolidProfile(avatar)
	case core.BackgroundTypeUnsplash:
		return is.FetchBackgroundUnsplash(backgroundType.Value)
	case core.BackgroundTypeUrl:
		return is.FetchBackgroundURL(backgroundType.Value, allowAnimated)
	default:
		return is.FetchBackgroundDefault(backgroundType.Value)
	}
}

// FetchBackgroundDefault returns an image from the static backgrounds.
func (is *ImageService) FetchBackgroundDefault(value string) (*FullImage, error) {
	background, ok := backgrounds[value]
	if !ok {
		background = backgrounds["default"]
	}

	return &FullImage{Frames: []image.Image{background}}, nil
}

// FetchBackgroundWelcomer returns an image from the database.
func (is *ImageService) FetchBackgroundWelcomer(value string, allowAnimated bool) (*FullImage, error) {
	// fetch from database
	var backgroundUuid uuid.UUID
	err := backgroundUuid.Parse(value)
	if err != nil {
		is.Logger.Error().Err(err).Str("value", value).Msg("Failed to convert value to valid UUID for background")

		return nil, err
	}

	background, err := is.Database.GetWelcomerImages(is.ctx, backgroundUuid)
	if err != nil {
		is.Logger.Error().Err(err).Str("value", value).Msg("Failed to fetch background from database")

		return nil, err
	}

	fullImage, err := openImage(background.Data, background.ImageType)
	if err != nil {
		is.Logger.Error().Err(err).Str("value", value).Msg("Failed to fetch background from database")

		return nil, err
	}

	return fullImage, nil
}

// getFullImageForColour creates a FullImage structure with a single pixel representing the given color.
// It takes a color.RGBA 'colour' and generates an image containing a single pixel of that color.
// The function returns a FullImage structure representing the single-pixel image.
func getFullImageForColour(colour color.RGBA) *FullImage {
	// Create a new RGBA image with a single pixel of the specified color
	im := image.NewRGBA(image.Rect(0, 0, 1, 1))
	im.Set(0, 0, colour)

	// Generate a FullImage structure representing the single-pixel image
	return &FullImage{
		Format: core.ImageFileTypeImagePng,
		Frames: []image.Image{im},
		Config: image.Config{
			Width:  1,
			Height: 1,
		},
	}
}

// FetchBackgroundSolid returns an image using the color provided as the value.
func (is *ImageService) FetchBackgroundSolid(value string) (*FullImage, error) {
	background, err := core.ParseColour(value, "")
	if err != nil {
		return nil, fmt.Errorf("failed to parse colour %s: %v", value, err)
	}

	return getFullImageForColour(*background), nil
}

// getCommonLuminance calculates the most common light color in the image.
// It takes an image.Image 'src' and a 'threshold' for defining lightness.
// It returns the most common light color as a color.RGBA value and a boolean flag ('ok') indicating success.
func getCommonLuminance(src image.Image, threshold float64) (colour color.RGBA, ok bool) {
	// Map to store the count of each color with its occurrence
	colorCount := make(map[color.RGBA]int)

	// Traverse through each pixel in the image
	bounds := src.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			// Extract RGB values of the pixel
			r, g, b, _ := src.At(x, y).RGBA()

			// Check if the color is considered light based on the provided threshold
			if isLightColorLuminance(r, g, b, threshold) {
				// Store the color in the map and count its occurrences
				colorCount[color.RGBA{uint8(r), uint8(g), uint8(b), 255}]++
			}
		}
	}

	// If no light color found, return default black color and set 'ok' flag to false
	if len(colorCount) == 0 {
		return colour, false
	}

	// Find the most common light color by counting occurrences
	maxOccurrences := 0
	for color, count := range colorCount {
		if count > maxOccurrences {
			colour = color
			maxOccurrences = count
		}
	}

	return colour, true
}

// isLightColorLuminance determines if a color is considered 'light' based on its luminance and a threshold value.
// It takes the red (r), green (g), blue (b) values and a threshold for luminance.
// It returns true if the color is light, false otherwise.
func isLightColorLuminance(r, g, b uint32, threshold float64) bool {
	// Calculate luminance of the color using RGB values
	lum := 0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)

	// Check if the luminance exceeds the threshold
	return lum > threshold
}

// FetchBackgroundSolidProfile uses the primary color of an avatar as the background.
// It attempts to identify the primary background color by analysing the provided image 'src'.
// It iterates through various thresholds for color luminance until a suitable color is found.
// It returns a FullImage structure representing the identified primary color as the background,
// or an error if the process encounters an issue.
func (is *ImageService) FetchBackgroundSolidProfile(src image.Image) (*FullImage, error) {
	// Initial threshold for solid profile luminance
	threshold := SolidProfileLuminance

	var colour color.RGBA
	var ok bool

	// Iterate through thresholds to find the primary color
	for threshold > 0 {
		// Attempt to identify the most common light color based on the current threshold
		colour, ok = getCommonLuminance(src, threshold)

		// If not found, decrease the threshold and try again
		if !ok {
			threshold -= 0.1
		} else {
			break // Exit the loop if a color is found
		}
	}

	// Generate a FullImage representation for the identified color
	return getFullImageForColour(colour), nil
}

// FetchBackgroundUnsplash returns an image from unsplash, identified by the value.
func (is *ImageService) FetchBackgroundUnsplash(value string) (*FullImage, error) {
	// fetch from unsplash

	return nil, ErrNotImplemented
}

// FetchBackgroundURL returns an image from a specific URL.
func (is *ImageService) FetchBackgroundURL(value string, allowAnimated bool) (*FullImage, error) {
	// fetch from url.

	return nil, ErrNotImplemented
}

// openImage decodes an image from byte data based on the specified format.
// It takes a byte slice 'src' containing the image data and a 'format' string indicating the image format.
// It returns a FullImage structure representing the decoded image and an error if the decoding process encounters any issues.
func openImage(src []byte, format string) (fullImage *FullImage, err error) {
	// Attempt to parse the image file format
	fileFormat, err := core.ParseImageFileType(format)
	if err != nil {
		// Set a default format to PNG if unable to parse
		fileFormat = core.ImageFileTypeImagePng
	}

	// Create a buffer with the image data
	b := bytes.NewBuffer(src)

	// Decode the image based on its format
	switch fileFormat {
	case core.ImageFileTypeImageGif:
		// Decode GIF images
		gif, err := gif.DecodeAll(b)
		if err != nil {
			return nil, err
		}

		// Populate FullImage structure for GIF images
		fullImage = &FullImage{
			Format:          core.ImageFileTypeImageGif,
			Frames:          make([]image.Image, len(gif.Image)),
			Config:          gif.Config,
			Delay:           gif.Delay,
			LoopCount:       gif.LoopCount,
			Disposal:        gif.Disposal,
			BackgroundIndex: gif.BackgroundIndex,
		}

		// Store individual frames of the GIF
		for frameIndex, frame := range gif.Image {
			fullImage.Frames[frameIndex] = frame
		}
	default:
		// Decode non-GIF images (e.g., PNG, JPEG, etc.)
		im, _, err := image.Decode(b)
		if err != nil {
			return nil, err
		}

		// Populate FullImage structure for non-GIF images
		fullImage = &FullImage{
			Format: fileFormat,
			Frames: []image.Image{im},
			Config: image.Config{
				Width:  im.Bounds().Dx(),
				Height: im.Bounds().Dy(),
			},
		}
	}

	return
}
