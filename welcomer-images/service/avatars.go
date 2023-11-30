package service

import (
	"image"
	"image/color"
	"math"
	"net/http"

	"github.com/WelcomerTeam/Discord/discord"
	core "github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/disintegration/imaging"
	"github.com/fogleman/gg"
)

var (
	transparent = color.NRGBA{0, 0, 0, 0}
)

const (
	UserAgent = "WelcomerImageService (https://github.com/WelcomerTeam/Welcomer/welcomer-images, " + VERSION + ")"
)

func (is *ImageService) FetchAvatar(userID discord.Snowflake, avatarURL string) (image.Image, error) {
	req, err := http.NewRequest(http.MethodGet, avatarURL, nil)
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
	}

	return im, nil
}

// applyAvatarEffects applies masking and resizing to the avatar.
// Outputs an image.Image with same dimension as src image.
func applyAvatarEffects(avatar image.Image, generateImageOptions GenerateImageOptions) (im image.Image, err error) {
	cropPix := int(math.Floor(float64(avatar.Bounds().Dx()) / 8))

	atlas := image.NewRGBA(image.Rect(
		0, 0,
		avatar.Bounds().Dx()+(generateImageOptions.ProfileBorderWidth*2),
		avatar.Bounds().Dy()+(generateImageOptions.ProfileBorderWidth*2),
	))

	context := gg.NewContextForRGBA(atlas)

	context.SetColor(generateImageOptions.ProfileBorderColor)
	context.Clear()

	rounding := float64(0)

	var avatarImage image.Image

	switch generateImageOptions.ProfileBorderCurve {
	case core.ImageProfileBorderTypeCircular:
		rounding = 1000

		canCrop := (avatar.At(cropPix, cropPix) == transparent &&
			avatar.At(
				avatar.Bounds().Dx()-cropPix,
				avatar.Bounds().Dy()-cropPix,
			) == transparent)

		if canCrop {
			avatarMinimize := image.NewRGBA(avatar.Bounds())
			avatarContext := gg.NewContextForRGBA(avatarMinimize)

			avatarContext.DrawImage(
				imaging.Resize(
					avatar,
					(avatar.Bounds().Dx()-(cropPix*2)),
					(avatar.Bounds().Dx()-(cropPix*2)),
					imaging.Lanczos,
				),
				cropPix,
				cropPix,
			)

			avatarImage = roundImage(avatarMinimize, 1000)
		} else {
			avatarImage = roundImage(avatar, 1000)
		}
	case core.ImageProfileBorderTypeRounded:
		rounding = 16 + float64(generateImageOptions.ProfileBorderWidth)
		avatarImage = roundImage(avatar, 16)
	case core.ImageProfileBorderTypeSquared:
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
func roundImage(im image.Image, r float64) image.Image {
	b := im.Bounds()
	r = math.Max(0, math.Min(r, float64(b.Dy())/2))
	context := gg.NewContext(b.Dx(), b.Dy())
	context.DrawRoundedRectangle(0, 0, float64(b.Dx()), float64(b.Dy()), r)
	context.Clip()
	context.DrawImage(im, 0, 0)

	return context.Image()
}
