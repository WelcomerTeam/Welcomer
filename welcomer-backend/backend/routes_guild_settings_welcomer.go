package backend

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	discord "github.com/WelcomerTeam/Discord/discord"
	recoder "github.com/WelcomerTeam/Recoder"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	utils "github.com/WelcomerTeam/Welcomer/welcomer-utils"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/savsgio/gotils/strconv"
	gotils_strconv "github.com/savsgio/gotils/strconv"
)

//go:embed imageFailure.png
var imageFailure []byte

var gen = uuid.NewGen()

const (
	MaxBackgroundSize = 20_000_000 // 20MB file size.
	MaxFileResolution = 16_777_216 // Maximum pixels. This is ~4096x4096 for a 1:1 image.

	MIMEPNG  = "image/png"
	MIMEJPEG = "image/jpeg"
	MIMEGIF  = "image/gif"
	MIMEWEBP = "image/webp"
)

var RecoderQuantizationAttributes = recoder.NewQuantizationAttributes()

// Route GET /api/guild/:guildID/welcomer
func getGuildSettingsWelcomer(ctx *gin.Context) {
	requireOAuthAuthorization(ctx, func(ctx *gin.Context) {
		requireGuildElevation(ctx, func(ctx *gin.Context) {
			guildID := tryGetGuildID(ctx)

			welcomerText, err := backend.Database.GetWelcomerTextGuildSettings(ctx, int64(guildID))
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					welcomerText = &database.GuildSettingsWelcomerText{
						GuildID:       int64(guildID),
						ToggleEnabled: database.DefaultWelcomerText.ToggleEnabled,
						Channel:       database.DefaultWelcomerText.Channel,
						MessageFormat: database.DefaultWelcomerText.MessageFormat,
					}
				}

				backend.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to get guild welcomer text settings")
			}

			welcomerImages, err := backend.Database.GetWelcomerImagesGuildSettings(ctx, int64(guildID))
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					welcomerImages = &database.GuildSettingsWelcomerImages{
						GuildID:                int64(guildID),
						ToggleEnabled:          database.DefaultWelcomerImages.ToggleEnabled,
						ToggleImageBorder:      database.DefaultWelcomerImages.ToggleImageBorder,
						ToggleShowAvatar:       database.DefaultWelcomerImages.ToggleShowAvatar,
						BackgroundName:         database.DefaultWelcomerImages.BackgroundName,
						ColourText:             database.DefaultWelcomerImages.ColourText,
						ColourTextBorder:       database.DefaultWelcomerImages.ColourTextBorder,
						ColourImageBorder:      database.DefaultWelcomerImages.ColourImageBorder,
						ColourProfileBorder:    database.DefaultWelcomerImages.ColourProfileBorder,
						ImageAlignment:         database.DefaultWelcomerImages.ImageAlignment,
						ImageTheme:             database.DefaultWelcomerImages.ImageTheme,
						ImageMessage:           database.DefaultWelcomerImages.ImageMessage,
						ImageProfileBorderType: database.DefaultWelcomerImages.ImageProfileBorderType,
					}
				}

				backend.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to get guild welcomer images settings")
			}

			welcomerDMs, err := backend.Database.GetWelcomerDMsGuildSettings(ctx, int64(guildID))
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					welcomerDMs = &database.GuildSettingsWelcomerDms{
						GuildID:             int64(guildID),
						ToggleEnabled:       database.DefaultWelcomerDms.ToggleEnabled,
						ToggleUseTextFormat: database.DefaultWelcomerDms.ToggleUseTextFormat,
						ToggleIncludeImage:  database.DefaultWelcomerDms.ToggleIncludeImage,
						MessageFormat:       database.DefaultWelcomerDms.MessageFormat,
					}
				}

				backend.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to get guild welcomer dms settings")
			}

			guildBackgrounds, err := backend.Database.GetWelcomerImagesByGuildId(ctx, int64(guildID))
			if err != nil {
				backend.Logger.Warn().Err(err).
					Int64("guild_id", int64(guildID)).
					Msg("Failed to get guild welcomer images backgrounds")
			}

			customIDs := make([]string, len(guildBackgrounds))
			for i, b := range guildBackgrounds {
				customIDs[i] = b.ImageUuid.String()
			}

			partial := GuildSettingsWelcomerSettingsToPartial(*welcomerText, *welcomerImages, *welcomerDMs, &GuildSettingsWelcomerCustom{
				CustomBackgroundIDs: customIDs,
			})

			ctx.JSON(http.StatusOK, BaseResponse{
				Ok:   true,
				Data: partial,
			})
		})
	})
}

// Route POST /api/guild/:guildID/welcomer
func setGuildSettingsWelcomer(ctx *gin.Context) {
	requireOAuthAuthorization(ctx, func(ctx *gin.Context) {
		requireGuildElevation(ctx, func(ctx *gin.Context) {
			partial := &GuildSettingsWelcomer{}

			var fileValue *multipart.FileHeader
			var err error

			switch ctx.ContentType() {
			case gin.MIMEMultipartPOSTForm:
				multipart, err := ctx.MultipartForm()
				if err == nil {
					fileValue = multipart.File["file"][0]
					jsonValue := multipart.Value["json"][0]

					err = json.Unmarshal(strconv.S2B(jsonValue), &partial)
					if err != nil {
						ctx.JSON(http.StatusBadRequest, BaseResponse{
							Ok:    false,
							Error: err.Error(),
						})

						return
					}
				}
			case gin.MIMEJSON:
				err = ctx.BindJSON(partial)
			default:
				ctx.JSON(http.StatusNotAcceptable, BaseResponse{
					Ok:    false,
					Error: ErrInvalidContentType.Error(),
				})

				return
			}

			if err != nil {
				ctx.JSON(http.StatusBadRequest, BaseResponse{
					Ok:    false,
					Error: err.Error(),
				})

				return
			}

			err = doValidateWelcomer(partial)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, BaseResponse{
					Ok:    false,
					Error: err.Error(),
				})

				return
			}

			guildID := tryGetGuildID(ctx)

			welcomerText, welcomerImages, welcomerDMs := PartialToGuildSettingsWelcomerSettings(int64(guildID), partial)

			if welcomerImages.BackgroundName == utils.CustomBackgroundPrefix+"upload" {
				if fileValue != nil {
					hasWelcomerPro, hasCustomBackgrounds, err := getGuildMembership(guildID)
					if err != nil {
						backend.Logger.Warn().Err(err).Int("guildID", int(guildID)).Msg("Exception getting welcomer membership")
					}

					if !hasWelcomerPro && !hasCustomBackgrounds {
						ctx.JSON(http.StatusBadRequest, BaseResponse{
							Ok:    false,
							Error: ErrCannotUseCustomBackgrounds.Error(),
						})

						return
					}

					if fileValue.Size > MaxBackgroundSize {
						ctx.JSON(http.StatusRequestEntityTooLarge, BaseResponse{
							Ok:    false,
							Error: ErrBackgroundTooLarge.Error(),
						})

						return
					}

					fileOpen, err := fileValue.Open()
					if err != nil {
						ctx.JSON(http.StatusInternalServerError, BaseResponse{
							Ok: false,
						})

						return
					}

					defer fileOpen.Close()

					var fileBytes bytes.Buffer

					_, err = fileBytes.ReadFrom(fileOpen)
					if err != nil {
						ctx.JSON(http.StatusInternalServerError, BaseResponse{
							Ok: false,
						})

						return
					}

					_, err = fileOpen.Seek(0, 0)
					if err != nil {
						ctx.JSON(http.StatusInternalServerError, BaseResponse{
							Ok: false,
						})

						return
					}

					mimeType := http.DetectContentType(fileBytes.Bytes())

					var res *database.WelcomerImages

					switch mimeType {
					case MIMEGIF, MIMEPNG, MIMEJPEG:
						switch {
						case mimeType == MIMEGIF && hasWelcomerPro:
							res, err = welcomerCustomBackgroundsUploadGIF(ctx, guildID, fileValue, fileOpen)
						case mimeType == MIMEPNG, mimeType == MIMEGIF && !hasWelcomerPro:
							// We will still accept GIFs if they do not have Welcomer Pro, however
							// they will be converted to PNG. This saves having to extract the first
							// frame every time we try to generate the resulting welcome image.
							// If you do not like this, get Welcomer Pro :)
							// It helps me out.
							res, err = welcomerCustomBackgroundsUploadPNG(ctx, guildID, fileOpen)
						case mimeType == MIMEJPEG:
							res, err = welcomerCustomBackgroundsUploadJPG(ctx, guildID, fileOpen)
						default:
							ctx.JSON(http.StatusBadRequest, BaseResponse{
								Ok:    false,
								Error: ErrFileNotSupported.Error(),
							})

							return
						}

						if err != nil {
							backend.Logger.Error().Err(err).
								Int64("guild_id", int64(guildID)).
								Int64("filesize", fileValue.Size).
								Str("mimetype", mimeType).
								Msg("Failed to upload custom utils.background")

							switch {
							case errors.Is(err, ErrBackgroundTooLarge),
								errors.Is(err, ErrFileSizeTooLarge),
								errors.Is(err, ErrFileNotSupported),
								errors.Is(err, ErrConversionFailed):

								ctx.JSON(http.StatusBadRequest, BaseResponse{
									Ok:    false,
									Error: err.Error(),
								})
							default:
								ctx.JSON(http.StatusInternalServerError, BaseResponse{
									Ok: false,
								})
							}

							return
						}

						// Set background name from custom:upload to custom:00000000-0000-0000-0000-000000000000
						// depending on uploaded file.
						welcomerImages.BackgroundName = utils.CustomBackgroundPrefix + res.ImageUuid.String()

						// Remove previous welcome images
						backgrounds, err := backend.Database.GetWelcomerImagesByGuildId(ctx, int64(guildID))
						if err == nil {
							for _, background := range backgrounds {
								if background.ImageUuid == res.ImageUuid {
									continue
								}

								_, err = backend.Database.DeleteWelcomerImage(ctx, background.ImageUuid)
								if err != nil {
									backend.Logger.Warn().
										Err(err).
										Int64("guild_id", int64(guildID)).
										Str("uuid", background.ImageUuid.String()).
										Msg("Failed to remove background database entry")

									continue
								}
							}
						}
					default:
						backend.Logger.Info().
							Int64("guild_id", int64(guildID)).
							Int64("filesize", fileValue.Size).
							Str("mimetype", mimeType).
							Msg("Rejected custom utils.background")

						ctx.JSON(http.StatusBadRequest, BaseResponse{
							Ok:    false,
							Error: ErrFileNotSupported.Error(),
						})

						return
					}
				}
			}

			databaseWelcomerTextGuildSettings := database.CreateOrUpdateWelcomerTextGuildSettingsParams(*welcomerText)

			user := tryGetUser(ctx)
			backend.Logger.Info().Int64("guild_id", int64(guildID)).Interface("obj", *welcomerText).Int64("user_id", int64(user.ID)).Msg("Creating or updating guild welcomerText settings")

			err = utils.RetryWithFallback(
				func() error {
					_, err = backend.Database.CreateOrUpdateWelcomerTextGuildSettings(ctx, databaseWelcomerTextGuildSettings)
					return err
				},
				func() error {
					return welcomer.EnsureGuild(ctx, backend.Database, discord.Snowflake(guildID))
				},
				nil,
			)
			if err != nil {
				backend.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to create or update guild welcomer text settings")

				ctx.JSON(http.StatusInternalServerError, BaseResponse{
					Ok: false,
				})

				return
			}

			databaseWelcomerImagesGuildSettings := database.CreateOrUpdateWelcomerImagesGuildSettingsParams(*welcomerImages)

			backend.Logger.Info().Int64("guild_id", int64(guildID)).Interface("obj", *welcomerImages).Int64("user_id", int64(user.ID)).Msg("Creating or updating guild welcomerImages settings")

			_, err = backend.Database.CreateOrUpdateWelcomerImagesGuildSettings(ctx, databaseWelcomerImagesGuildSettings)
			if err != nil {
				backend.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to create or update guild welcomer images settings")

				ctx.JSON(http.StatusInternalServerError, BaseResponse{
					Ok: false,
				})

				return
			}

			databaseWelcomerDMsGuildSettings := database.CreateOrUpdateWelcomerDMsGuildSettingsParams(*welcomerDMs)

			backend.Logger.Info().Int64("guild_id", int64(guildID)).Interface("obj", *welcomerDMs).Int64("user_id", int64(user.ID)).Msg("Creating or updating guild welcomerDMs settings")

			_, err = backend.Database.CreateOrUpdateWelcomerDMsGuildSettings(ctx, databaseWelcomerDMsGuildSettings)
			if err != nil {
				backend.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to create or update guild welcomer dms settings")

				ctx.JSON(http.StatusInternalServerError, BaseResponse{
					Ok: false,
				})

				return
			}

			getGuildSettingsWelcomer(ctx)
		})
	})
}

// Route GET /api/welcomer/preview/:key
func getGuildWelcomerPreview(ctx *gin.Context) {
	key := ctx.Param(KeyKey)

	uuid := uuid.UUID{}

	err := uuid.UnmarshalText(gotils_strconv.S2B(key))
	if err != nil {
		backend.Logger.Info().
			Str("key", key).Msg("Failed to unmarshal key to uuid")

		ctx.Data(http.StatusNotFound, "image/png", imageFailure)

		return
	}

	background, err := backend.Database.GetWelcomerImages(ctx, uuid)
	if err != nil || background == nil {
		backend.Logger.Info().Str("key", key).Msg("Failed to find utils.background with key")

		ctx.Data(http.StatusNotFound, "image/png", imageFailure)

		return
	}

	ctx.Data(http.StatusOK, background.ImageType, background.Data)
}

func welcomerCustomBackgroundsUploadGIF(ctx context.Context, guildID discord.Snowflake, file *multipart.FileHeader, fileBytes io.ReadSeeker) (*database.WelcomerImages, error) {
	start := time.Now()

	backend.Logger.Info().Int64("size", file.Size).Msg("Recoding image")

	recoderResult, err := recoder.RecodeImage(fileBytes, RecoderQuantizationAttributes)
	if err != nil {
		return nil, err
	}

	backend.Logger.Info().Dur("time", time.Since(start)).Msg("Recoded image successfully")

	buf := bytes.NewBuffer(nil)

	_, err = buf.ReadFrom(recoderResult)
	if err != nil {
		return nil, err
	}

	var welcomerImageUuid uuid.UUID
	welcomerImageUuid, _ = gen.NewV7()

	return backend.Database.CreateWelcomerImages(ctx, database.CreateWelcomerImagesParams{
		ImageUuid: welcomerImageUuid,
		GuildID:   int64(guildID),
		CreatedAt: time.Now(),
		ImageType: utils.ImageFileTypeImageGif.String(),
		Data:      buf.Bytes(),
	})
}

func welcomerCustomBackgroundsUploadPNG(ctx context.Context, guildID discord.Snowflake, fileBytes io.ReadSeeker) (*database.WelcomerImages, error) {
	// Validate file and get size
	img, err := png.Decode(fileBytes)
	if err != nil {
		return nil, err
	}

	_, _ = fileBytes.Seek(0, 0)

	// Validate image resolution
	imageSize := img.Bounds().Size()
	if (imageSize.X * imageSize.Y) > MaxFileResolution {
		backend.Logger.Info().
			Int("width", imageSize.X).Int("height", imageSize.Y).
			Int("total", (imageSize.X*imageSize.Y)).Int("max", MaxFileResolution).
			Msg("Rejected image due to resolution")

		return nil, ErrFileSizeTooLarge
	}

	buf := bytes.NewBuffer(nil)

	_, err = buf.ReadFrom(fileBytes)
	if err != nil {
		return nil, err
	}

	var welcomerImageUuid uuid.UUID
	welcomerImageUuid, _ = gen.NewV7()

	return backend.Database.CreateWelcomerImages(ctx, database.CreateWelcomerImagesParams{
		ImageUuid: welcomerImageUuid,
		GuildID:   int64(guildID),
		CreatedAt: time.Now(),
		ImageType: utils.ImageFileTypeImagePng.String(),
		Data:      buf.Bytes(),
	})
}

func welcomerCustomBackgroundsUploadJPG(ctx context.Context, guildID discord.Snowflake, fileBytes io.ReadSeeker) (*database.WelcomerImages, error) {
	// Validate file and get size
	img, err := jpeg.Decode(fileBytes)
	if err != nil {
		return nil, err
	}

	_, _ = fileBytes.Seek(0, 0)

	// Validate image resolution
	imageSize := img.Bounds().Size()
	if (imageSize.X * imageSize.Y) > MaxFileResolution {
		backend.Logger.Info().
			Int("width", imageSize.X).Int("height", imageSize.Y).
			Int("total", (imageSize.X*imageSize.Y)).Int("max", MaxFileResolution).
			Msg("Rejected image due to resolution")

		return nil, ErrFileSizeTooLarge
	}

	buf := bytes.NewBuffer(nil)

	_, err = buf.ReadFrom(fileBytes)
	if err != nil {
		return nil, err
	}

	var welcomerImageUuid uuid.UUID
	welcomerImageUuid, _ = gen.NewV7()

	return backend.Database.CreateWelcomerImages(ctx, database.CreateWelcomerImagesParams{
		ImageUuid: welcomerImageUuid,
		GuildID:   int64(guildID),
		CreatedAt: time.Now(),
		ImageType: utils.ImageFileTypeImageJpeg.String(),
		Data:      buf.Bytes(),
	})
}

// Validates welcomer guild settings
func doValidateWelcomer(guildSettings *GuildSettingsWelcomer) error {
	if guildSettings.Text.MessageFormat != "" {
		if !utils.IsValidEmbed(guildSettings.Text.MessageFormat) {
			return fmt.Errorf("text message is invalid: %w", ErrInvalidJSON)
		}
	}

	if guildSettings.DMs.MessageFormat != "" {
		if !utils.IsValidEmbed(guildSettings.DMs.MessageFormat) {
			return fmt.Errorf("dms message is invalid: %w", ErrInvalidJSON)
		}
	}

	if guildSettings.Text.ToggleEnabled {
		if guildSettings.Text.MessageFormat == "" {
			return fmt.Errorf("text message is invalid: %w", ErrRequired)
		}
	}

	if guildSettings.DMs.ToggleEnabled {
		if guildSettings.DMs.ToggleUseTextFormat {
			if guildSettings.Text.MessageFormat == "" {
				return fmt.Errorf("text message is invalid: %w", ErrRequired)
			}
		} else {
			if guildSettings.DMs.MessageFormat == "" {
				return fmt.Errorf("dms message is invalid: %w", ErrRequired)
			}
		}
	}

	if guildSettings.Images.ToggleEnabled || guildSettings.Text.ToggleEnabled {
		if guildSettings.Text.Channel == nil {
			return fmt.Errorf("text channel is invalid: %w", ErrRequired)
		}

		if !utils.IsValidInteger(*guildSettings.Text.Channel) {
			return fmt.Errorf("text channel is invalid: %w", ErrChannelInvalid)
		}
	}

	if guildSettings.Images.ToggleEnabled {
		if !utils.IsValidBackground(guildSettings.Images.BackgroundName) {
			return fmt.Errorf("image background is invalid: %w", ErrInvalidBackground)
		}

		if !utils.IsValidColour(guildSettings.Images.ColourText) {
			return fmt.Errorf("image text colour is invalid: %w", ErrInvalidColour)
		}

		if !utils.IsValidColour(guildSettings.Images.ColourTextBorder) {
			return fmt.Errorf("image text border colour is invalid: %w", ErrInvalidColour)
		}

		if !utils.IsValidColour(guildSettings.Images.ColourImageBorder) {
			return fmt.Errorf("image border colour is invalid: %w", ErrInvalidColour)
		}

		if !utils.IsValidColour(guildSettings.Images.ColourProfileBorder) {
			return fmt.Errorf("image profile border colour is invalid: %w", ErrInvalidColour)
		}

		if !utils.IsValidImageAlignment(guildSettings.Images.ImageAlignment) {
			return fmt.Errorf("image ImageAlignment is invalid: %w", ErrInvalidImageAlignment)
		}

		if !utils.IsValidImageProfileBorderType(guildSettings.Images.ImageProfileBorderType) {
			return fmt.Errorf("image ImageProfileBorderType is invalid: %w", ErrInvalidProfileBorderType)
		}

		if !utils.IsValidImageTheme(guildSettings.Images.ImageTheme) {
			return fmt.Errorf("image ImageTheme is invalid: %w", ErrInvalidImageTheme)
		}
	}

	return nil
}

func registerGuildSettingsWelcomerRoutes(g *gin.Engine) {
	g.GET("/api/guild/:guildID/welcomer", getGuildSettingsWelcomer)
	g.POST("/api/guild/:guildID/welcomer", setGuildSettingsWelcomer)

	g.GET("/api/welcomer/preview/:key", getGuildWelcomerPreview)
}
