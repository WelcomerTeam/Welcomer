package service

import (
	"context"
	"image"
	"image/color"
	"math"
	"net/http"

	"github.com/WelcomerTeam/Discord/discord"
	utils "github.com/WelcomerTeam/Welcomer/welcomer-utils"
	"github.com/disintegration/imaging"
	"github.com/fogleman/gg"
)

var (
	transparent = color.RGBA{0, 0, 0, 0}
)

const (
	UserAgent = "WelcomerImageService (https://github.com/WelcomerTeam/Welcomer, " + VERSION + ")"
)

func (is *ImageService) FetchAvatar(ctx context.Context, userID discord.Snowflake, avatarURL string) (image.Image, error) {
	url, isValidURL := utils.IsValidURL(avatarURL)
	if !isValidURL {
		return nil, ErrInvalidURL
	}

	query := url.Query()
	if query.Has("size") {
		// Set size to 256, if present.
		query.Set("size", "256")
		url.RawQuery = query.Encode()
		avatarURL = url.String()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, avatarURL, nil)
	if err != nil {
		is.Logger.Error().Err(err).Str("url", avatarURL).Msg("Failed to create new request for avatar")

		return nil, err
	}

	req.Header.Set("User-Agent", UserAgent)

	resp, err := is.Client.Do(req)
	if err != nil {
		is.Logger.Error().Err(err).Str("url", avatarURL).Msg("Failed to fetch profile picture for avatar")

		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		is.Logger.Error().Err(err).Int("status", resp.StatusCode).Str("url", avatarURL).Msg("Failed to fetch profile picture for avatar")

		return nil, ErrAvatarFetchFailed
	}

	im, _, err := image.Decode(resp.Body)
	if err != nil {
		is.Logger.Error().Err(err).Msg("Failed to decode profile picture of user")

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

	context.SetColor(generateImageOptions.ProfileBorderColor)
	context.Clear()

	rounding := float64(0)

	var avatarImage image.Image

	switch generateImageOptions.ProfileBorderCurve {
	case utils.ImageProfileBorderTypeCircular:
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
	case utils.ImageProfileBorderTypeRounded:
		rounding = 16 + float64(generateImageOptions.ProfileBorderWidth)
		avatarImage = roundImage(avatar, 16)
	case utils.ImageProfileBorderTypeSquared:
		avatarImage = avatar
	default:
		err = ErrUnknownProfileBorderType

		return
	}

	context.DrawImageAnchored(
		avatarImage,
		context.Width()/2,
		context.Height()/2,
		0.5,
		0.5,
	)

	return roundImage(atlas, rounding), nil
}

// roundImage cuts out a rounded segment from an image.
func roundImage(im image.Image, radius float64) image.Image {
	bounds := im.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	radius = math.Max(0, math.Min(radius, float64(height)/2))
	context := gg.NewContext(width, height)
	context.DrawRoundedRectangle(0, 0, float64(width), float64(height), radius)
	context.Clip()
	context.DrawImage(im, 0, 0)

	return context.Image()
}
