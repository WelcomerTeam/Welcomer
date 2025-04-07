package service

import (
	"context"
	"image"
	"image/color"
	"math"
	"net/http"
	"net/url"

	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/disintegration/imaging"
	"github.com/fogleman/gg"
)

var transparent = color.RGBA{0, 0, 0, 0}

const (
	UserAgent = "WelcomerImageService (https://github.com/WelcomerTeam/Welcomer, " + VERSION + ")"
)

func (is *ImageService) FetchAvatar(ctx context.Context, avatarURL string) (image.Image, error) {
	parsedURL, isValidURL := welcomer.IsValidURL(avatarURL)
	if parsedURL == nil || !isValidURL {
		return nil, ErrInvalidURL
	}

	queryValues, err := url.ParseQuery(parsedURL.RawQuery)
	if err != nil {
		return nil, err
	}

	if queryValues.Has("size") {
		// Set size to 256, if present.
		queryValues.Set("size", "256")
		parsedURL.RawQuery = queryValues.Encode()
		avatarURL = parsedURL.String()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, avatarURL, nil)
	if err != nil {
		welcomer.Logger.Error().Err(err).Str("url", avatarURL).Msg("Failed to create new request for avatar")

		return nil, err
	}

	req.Header.Set("User-Agent", UserAgent)

	resp, err := is.Client.Do(req)
	if err != nil || resp == nil {
		welcomer.Logger.Error().Err(err).Str("url", avatarURL).Msg("Failed to fetch profile picture for avatar")

		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		welcomer.Logger.Error().Err(err).Int("status", resp.StatusCode).Str("url", avatarURL).Msg("Failed to fetch profile picture for avatar")

		return nil, ErrAvatarFetchFailed
	}

	im, _, err := image.Decode(resp.Body)
	if err != nil || im == nil {
		welcomer.Logger.Error().Err(err).Msg("Failed to decode profile picture of user")

		return nil, err
	}

	return im, nil
}

func hasImageTransparentBorder(image image.Image, w, h int) (bool, int) {
	padding := w / 8

	// Check the sides of the image to see if they are all transparent.
	// This makes certain avatars look a bit nicer.

	hasTransparentBorder := image.At(padding, padding) == transparent &&
		image.At(w/2, padding) == transparent &&
		image.At(w-padding, padding) == transparent &&
		image.At(padding, h/2) == transparent &&
		image.At(w-padding, h/2) == transparent &&
		image.At(padding, h-padding) == transparent &&
		image.At(w/2, h-padding) == transparent &&
		image.At(w-padding, h-padding) == transparent

	return hasTransparentBorder, padding
}

// applyAvatarEffects applies masking and resizing to the avatar.
// Outputs an image.Image with same dimension as src image.
func applyAvatarEffects(avatar image.Image, generateImageOptions GenerateImageOptions) (im image.Image, err error) {
	bounds := avatar.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	atlas := image.NewRGBA(image.Rect(
		0, 0,
		width+(generateImageOptions.ProfileBorderWidth*2),
		height+(generateImageOptions.ProfileBorderWidth*2),
	))

	context := gg.NewContextForRGBA(atlas)

	if generateImageOptions.ProfileBorderWidth == 0 {
		context.SetColor(transparent)
	} else {
		context.SetColor(generateImageOptions.ProfileBorderColor)
	}

	context.Clear()

	rounding := float64(0)

	var avatarImage image.Image

	switch generateImageOptions.ProfileBorderCurve {
	case welcomer.ImageProfileBorderTypeCircular:
		rounding = 1000

		if canCrop, cropPix := hasImageTransparentBorder(avatar, width, height); canCrop {
			avatarMinimize := image.NewRGBA(bounds)
			avatarContext := gg.NewContextForRGBA(avatarMinimize)

			avatarContext.DrawImage(
				imaging.Resize(
					avatar,
					(width-(cropPix*2)),
					(height-(cropPix*2)),
					imaging.MitchellNetravali,
				),
				cropPix,
				cropPix,
			)

			avatarImage = roundImage(avatarMinimize, 1000)
		} else {
			avatarImage = roundImage(avatar, 1000)
		}
	case welcomer.ImageProfileBorderTypeRounded:
		rounding = 16 + float64(generateImageOptions.ProfileBorderWidth)
		avatarImage = roundImage(avatar, 16)
	case welcomer.ImageProfileBorderTypeSquared:
		avatarImage = avatar
	default:
		return avatar, ErrUnknownProfileBorderType
	}

	context.DrawImageAnchored(
		avatarImage,
		context.Width()/2,
		context.Height()/2,
		0.5,
		0.5,
	)

	if rounding > 0 {
		im = roundImage(atlas, rounding)
	} else {
		im = atlas
	}

	return im, nil
}

// roundImage cuts out a rounded segment from an image.
func roundImage(im image.Image, radius float64) image.Image {
	bounds := im.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Add minor offset to the rounding of images to prevent cutting off the edges.
	offset := float64(4)

	radius = math.Max(0, math.Min(radius, float64(height)/2))
	context := gg.NewContext(width, height)
	context.DrawRoundedRectangle(offset, offset, float64(width)-offset, float64(height)-offset, radius)
	context.Clip()
	context.DrawImage(im, 0, 0)

	return context.Image()
}
