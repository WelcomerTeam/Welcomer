package service

import (
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/font/sfnt"
)

const DefaultFont = "" // TODO

const (
	fontDPI = 72
)

type Font struct {
	Font *sfnt.Font
}

type FontFace struct {
	Face *font.Face
}

// CreateFontPackHook returns a newFace function with an argument.
func (is *ImageService) CreateFontPackHook(f string) func(float64) font.Face {
	return func(i float64) font.Face {
		return is.CreateFontPack(f, i)
	}
}

// CreateFontPack creates a pack of fonts with fallback and the one passed.
func (is *ImageService) CreateFontPack(font string, size float64) *MultiFace {
	face := new(MultiFace)

	f, fo, err := is.FetchFont(font, size)
	if err != nil {
		is.Logger.Warn().Err(err).Str("font", font).Msg("Failed to fetch font in font pack")
	} else {
		face.AddTrueTypeFace(*f.Face, fo.Font)
	}

	for fontName := range fallback {
		f, fo, err = is.FetchFont(fontName, size)
		if err != nil {
			is.Logger.Warn().Err(err).Str("font", fontName).Msg("Failed to fetch fallback font in font pack")
		} else {
			face.AddTrueTypeFace(*f.Face, fo.Font)
		}
	}

	return face
}

// FetchFont fetches a font face with the specified size.
func (is *ImageService) FetchFont(f string, size float64) (face *FontFace, font *Font, err error) {
	font, ok := fonts[f]
	if !ok {
		font, ok = fallback[f]
	}

	if !ok {
		return nil, nil, ErrNoFontFound
	}

	fc, err := opentype.NewFace(font.Font, &opentype.FaceOptions{
		Size: float64(size),
		DPI:  fontDPI,
	})
	if err != nil {
		is.Logger.Error().Err(err).Msg("Failed to create font face")

		return nil, nil, err
	}

	face = &FontFace{&fc}

	return face, font, nil
}
