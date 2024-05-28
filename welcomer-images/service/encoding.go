package service

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	"image/png"
	"sync"

	utils "github.com/WelcomerTeam/Welcomer/welcomer-utils"
	gotils_strconv "github.com/savsgio/gotils/strconv"
	"github.com/ultimate-guitar/go-imagequant"
)

var (
	attr, _ = imagequant.NewAttributes()
)

func encodeFrames(frames []image.Image, background *FullImage) ([]byte, utils.ImageFileType, error) {
	if len(frames) == 0 {
		return nil, utils.ImageFileTypeUnknown, ErrMissingFrames
	}

	if len(frames) > 1 {
		return encodeFramesAsGif(frames, background)
	}

	return encodeFramesAsPng(frames, background)
}

func encodeFramesAsPng(frames []image.Image, background *FullImage) ([]byte, utils.ImageFileType, error) {
	b := bytes.NewBuffer(nil)

	err := png.Encode(b, frames[0])
	if err != nil {
		return nil, utils.ImageFileTypeUnknown, err
	}

	return b.Bytes(), utils.ImageFileTypeImagePng, nil
}

func encodeFramesAsGif(frames []image.Image, background *FullImage) ([]byte, utils.ImageFileType, error) {
	_frames := make([]*image.Paletted, len(frames))

	wg := sync.WaitGroup{}
	for frameNumber, frame := range frames {
		wg.Add(1)

		go func(_frameNumber int, frame image.Image) {
			p, _ := quantizeImage(frame)
			_frames[_frameNumber] = p

			wg.Done()
		}(frameNumber, frame)
	}

	wg.Wait()

	b := bytes.NewBuffer(nil)

	err := gif.EncodeAll(b, &gif.GIF{
		Image:           _frames,
		Delay:           background.Delay,
		LoopCount:       background.LoopCount,
		Disposal:        background.Disposal,
		Config:          background.Config,
		BackgroundIndex: background.BackgroundIndex,
	})
	if err != nil {
		return nil, utils.ImageFileTypeUnknown, err
	}

	return b.Bytes(), utils.ImageFileTypeImageGif, nil
}

// quantizeImage converts an image.Image to image.Paletted via imagequant.
func quantizeImage(src image.Image) (*image.Paletted, error) {
	b := src.Bounds()

	img, err := imagequant.NewImage(
		attr,
		gotils_strconv.B2S(imagequant.ImageToRgba32(src)),
		b.Dx(), b.Dy(), 1,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create imagequant image: %v", err)
	}

	defer img.Release()

	result, err := img.Quantize(attr)
	if err != nil {
		return nil, fmt.Errorf("failed to quantize image: %v", err)
	}

	defer result.Release()

	dst := image.NewPaletted(src.Bounds(), result.GetPalette())

	// WriteRemappedImage returns a list of bytes pointing to direct
	// palette indexes so we can just copy it over and it will be
	// using the optimal indexes.
	pixelMap, err := result.WriteRemappedImage()
	if err != nil {
		return dst, err
	}

	dst.Pix = pixelMap

	return dst, nil
}
