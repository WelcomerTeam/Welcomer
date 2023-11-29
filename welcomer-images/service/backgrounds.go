package service

import (
	"bytes"
	"image"
	"image/gif"

	core "github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/gofrs/uuid"
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

// FetchBackgroundSolid returns an image using the color provided as the value.
func (is *ImageService) FetchBackgroundSolid(value string) (*FullImage, error) {
	// parse value and make image

	return nil, ErrNotImplemented
}

// FetchBackgroundSolidProfile uses the primary color of an avatar as the background.
func (is *ImageService) FetchBackgroundSolidProfile(src image.Image) (*FullImage, error) {
	// try to get primary image from avatar. Histogram?

	return nil, ErrNotImplemented
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

func openImage(src []byte, format string) (fullImage *FullImage, err error) {
	fileFormat, err := core.ParseImageFileType(format)
	if err != nil {
		fileFormat = core.ImageFileTypeImagePng
	}

	b := bytes.NewBuffer(src)

	switch fileFormat {
	case core.ImageFileTypeImageGif:
		gif, err := gif.DecodeAll(b)
		if err != nil {
			return nil, err
		}

		fullImage = &FullImage{
			Format:          core.ImageFileTypeImageGif,
			Frames:          make([]image.Image, len(gif.Image)),
			Config:          gif.Config,
			Delay:           gif.Delay,
			LoopCount:       gif.LoopCount,
			Disposal:        gif.Disposal,
			BackgroundIndex: gif.BackgroundIndex,
		}

		for frameIndex, frame := range gif.Image {
			fullImage.Frames[frameIndex] = frame
		}
	default:
		im, _, err := image.Decode(b)
		if err != nil {
			return nil, err
		}

		fullImage = &FullImage{
			Format: core.ImageFileTypeImagePng,
			Frames: []image.Image{im},
			Config: image.Config{
				Width:  im.Bounds().Dx(),
				Height: im.Bounds().Dy(),
			},
		}
	}

	return
}
