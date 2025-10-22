package backend

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	discord "github.com/WelcomerTeam/Discord/discord"
	recoder "github.com/WelcomerTeam/Recoder"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
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

// Route GET /api/guild/:guildID/welcomer.
func getGuildSettingsWelcomer(ctx *gin.Context) {
	requireOAuthAuthorization(ctx, func(ctx *gin.Context) {
		requireGuildElevation(ctx, func(ctx *gin.Context) {
			guildID := tryGetGuildID(ctx)

			welcomerConfig, err := welcomer.Queries.GetWelcomerGuildSettings(ctx, int64(guildID))
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					welcomerConfig = &database.GuildSettingsWelcomer{
						GuildID:                          int64(guildID),
						AutoDeleteWelcomeMessages:        welcomer.DefaultWelcomer.AutoDeleteWelcomeMessages,
						WelcomeMessageLifetime:           welcomer.DefaultWelcomer.WelcomeMessageLifetime,
						AutoDeleteWelcomeMessagesOnLeave: welcomer.DefaultWelcomer.AutoDeleteWelcomeMessagesOnLeave,
					}
				}

				welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to get guild welcomer settings")
			}

			welcomerText, err := welcomer.Queries.GetWelcomerTextGuildSettings(ctx, int64(guildID))
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					welcomerText = &database.GuildSettingsWelcomerText{
						GuildID:       int64(guildID),
						ToggleEnabled: welcomer.DefaultWelcomerText.ToggleEnabled,
						Channel:       welcomer.DefaultWelcomerText.Channel,
						MessageFormat: welcomer.DefaultWelcomerText.MessageFormat,
					}
				}

				welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to get guild welcomer text settings")
			}

			welcomerImages, err := welcomer.Queries.GetWelcomerImagesGuildSettings(ctx, int64(guildID))
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					welcomerImages = &database.GuildSettingsWelcomerImages{
						GuildID:                int64(guildID),
						ToggleEnabled:          welcomer.DefaultWelcomerImages.ToggleEnabled,
						ToggleImageBorder:      welcomer.DefaultWelcomerImages.ToggleImageBorder,
						ToggleShowAvatar:       welcomer.DefaultWelcomerImages.ToggleShowAvatar,
						BackgroundName:         welcomer.DefaultWelcomerImages.BackgroundName,
						ColourText:             welcomer.DefaultWelcomerImages.ColourText,
						ColourTextBorder:       welcomer.DefaultWelcomerImages.ColourTextBorder,
						ColourImageBorder:      welcomer.DefaultWelcomerImages.ColourImageBorder,
						ColourProfileBorder:    welcomer.DefaultWelcomerImages.ColourProfileBorder,
						ImageAlignment:         welcomer.DefaultWelcomerImages.ImageAlignment,
						ImageTheme:             welcomer.DefaultWelcomerImages.ImageTheme,
						ImageMessage:           welcomer.DefaultWelcomerImages.ImageMessage,
						ImageProfileBorderType: welcomer.DefaultWelcomerImages.ImageProfileBorderType,
					}
				}

				welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to get guild welcomer images settings")
			}

			welcomerDMs, err := welcomer.Queries.GetWelcomerDMsGuildSettings(ctx, int64(guildID))
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					welcomerDMs = &database.GuildSettingsWelcomerDms{
						GuildID:             int64(guildID),
						ToggleEnabled:       welcomer.DefaultWelcomerDms.ToggleEnabled,
						ToggleUseTextFormat: welcomer.DefaultWelcomerDms.ToggleUseTextFormat,
						ToggleIncludeImage:  welcomer.DefaultWelcomerDms.ToggleIncludeImage,
						MessageFormat:       welcomer.DefaultWelcomerDms.MessageFormat,
					}
				}

				welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to get guild welcomer dms settings")
			}

			guildBackgrounds, err := welcomer.Queries.GetWelcomerImagesByGuildId(ctx, int64(guildID))
			if err != nil {
				welcomer.Logger.Warn().Err(err).
					Int64("guild_id", int64(guildID)).
					Msg("Failed to get guild welcomer images backgrounds")
			}

			customIDs := make([]string, len(guildBackgrounds))
			for i, b := range guildBackgrounds {
				customIDs[i] = b.ImageUuid.String()
			}

			partial := GuildSettingsWelcomerSettingsToPartial(*welcomerConfig, *welcomerText, *welcomerImages, *welcomerDMs, &GuildSettingsWelcomerCustom{
				CustomBackgroundIDs: customIDs,
			})

			ctx.JSON(http.StatusOK, BaseResponse{
				Ok:   true,
				Data: partial,
			})
		})
	})
}

// Route POST /api/guild/:guildID/welcomer.
func setGuildSettingsWelcomer(ctx *gin.Context) {
	requireOAuthAuthorization(ctx, func(ctx *gin.Context) {
		requireGuildElevation(ctx, func(ctx *gin.Context) {
			userID := tryGetUser(ctx).ID

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

			welcomerConfig, welcomerText, welcomerImages, welcomerDMs := PartialToGuildSettingsWelcomerSettings(int64(guildID), partial)

			if welcomerImages.BackgroundName == welcomer.CustomBackgroundPrefix+"upload" {
				if fileValue != nil {
					hasWelcomerPro, hasCustomBackgrounds, err := getGuildMembership(ctx, guildID)
					if err != nil {
						welcomer.Logger.Warn().Err(err).Int("guildID", int(guildID)).Msg("Exception getting welcomer membership")
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
							res, err = welcomerCustomBackgroundsUploadGIF(ctx, guildID, fileValue, fileOpen, userID)
						case mimeType == MIMEJPEG, mimeType == MIMEPNG, mimeType == MIMEGIF && !hasWelcomerPro:
							// We will still accept GIFs if they do not have Welcomer Pro, however
							// they will be converted to PNG. This saves having to extract the first
							// frame every time we try to generate the resulting welcome image.
							// If you do not like this, get Welcomer Pro :)
							// It helps me out.
							res, err = welcomerCustomBackgroundsUploadStatic(ctx, guildID, fileOpen, userID)
						default:
							ctx.JSON(http.StatusBadRequest, BaseResponse{
								Ok:    false,
								Error: ErrFileNotSupported.Error(),
							})

							return
						}

						if err != nil {
							welcomer.Logger.Error().Err(err).
								Int64("guild_id", int64(guildID)).
								Int64("filesize", fileValue.Size).
								Str("mimetype", mimeType).
								Msg("Failed to upload custom welcomer.background")

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
						welcomerImages.BackgroundName = welcomer.CustomBackgroundPrefix + res.ImageUuid.String()

						// Remove previous welcome images
						backgrounds, err := welcomer.Queries.GetWelcomerImagesByGuildId(ctx, int64(guildID))
						if err == nil {
							for _, background := range backgrounds {
								if background.ImageUuid == res.ImageUuid {
									continue
								}

								_, err = welcomer.Queries.DeleteWelcomerImage(ctx, background.ImageUuid)
								if err != nil {
									welcomer.Logger.Warn().
										Err(err).
										Int64("guild_id", int64(guildID)).
										Str("uuid", background.ImageUuid.String()).
										Msg("Failed to remove background database entry")

									continue
								}
							}
						}
					default:
						welcomer.Logger.Info().
							Int64("guild_id", int64(guildID)).
							Int64("filesize", fileValue.Size).
							Str("mimetype", mimeType).
							Msg("Rejected custom welcomer.background")

						ctx.JSON(http.StatusBadRequest, BaseResponse{
							Ok:    false,
							Error: ErrFileNotSupported.Error(),
						})

						return
					}
				}
			}

			user := tryGetUser(ctx)

			databaseWelcomerGuildSettings := database.CreateOrUpdateWelcomerGuildSettingsParams(*welcomerConfig)

			welcomer.Logger.Info().Int64("guild_id", int64(guildID)).Interface("obj", *welcomerConfig).Int64("user_id", int64(user.ID)).Msg("Creating or updating guild welcomer config settings")

			err = welcomer.RetryWithFallback(
				func() error {
					_, err = welcomer.CreateOrUpdateWelcomerGuildSettingsWithAudit(ctx, databaseWelcomerGuildSettings, discord.Snowflake(user.ID))

					return err
				},
				func() error {
					return welcomer.EnsureGuild(ctx, guildID)
				},
				nil,
			)
			if err != nil {
				welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to create or update guild welcomer config settings")

				ctx.JSON(http.StatusInternalServerError, BaseResponse{
					Ok: false,
				})

				return
			}

			databaseWelcomerTextGuildSettings := database.CreateOrUpdateWelcomerTextGuildSettingsParams(*welcomerText)

			welcomer.Logger.Info().Int64("guild_id", int64(guildID)).Interface("obj", *welcomerText).Int64("user_id", int64(user.ID)).Msg("Creating or updating guild welcomerText settings")

			err = welcomer.RetryWithFallback(
				func() error {
					_, err = welcomer.CreateOrUpdateWelcomerTextGuildSettingsWithAudit(ctx, databaseWelcomerTextGuildSettings, 0)

					return err
				},
				func() error {
					return welcomer.EnsureGuild(ctx, guildID)
				},
				nil,
			)
			if err != nil {
				welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to create or update guild welcomer text settings")

				ctx.JSON(http.StatusInternalServerError, BaseResponse{
					Ok: false,
				})

				return
			}

			databaseWelcomerImagesGuildSettings := database.CreateOrUpdateWelcomerImagesGuildSettingsParams(*welcomerImages)

			welcomer.Logger.Info().Int64("guild_id", int64(guildID)).Interface("obj", *welcomerImages).Int64("user_id", int64(user.ID)).Msg("Creating or updating guild welcomerImages settings")

			_, err = welcomer.CreateOrUpdateWelcomerImagesGuildSettingsWithAudit(ctx, databaseWelcomerImagesGuildSettings, 0)
			if err != nil {
				welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to create or update guild welcomer images settings")

				ctx.JSON(http.StatusInternalServerError, BaseResponse{
					Ok: false,
				})

				return
			}

			databaseWelcomerDMsGuildSettings := database.CreateOrUpdateWelcomerDMsGuildSettingsParams(*welcomerDMs)

			welcomer.Logger.Info().Int64("guild_id", int64(guildID)).Interface("obj", *welcomerDMs).Int64("user_id", int64(user.ID)).Msg("Creating or updating guild welcomerDMs settings")

			_, err = welcomer.CreateOrUpdateWelcomerDMsGuildSettingsWithAudit(ctx, databaseWelcomerDMsGuildSettings, 0)
			if err != nil {
				welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to create or update guild welcomer dms settings")

				ctx.JSON(http.StatusInternalServerError, BaseResponse{
					Ok: false,
				})

				return
			}

			getGuildSettingsWelcomer(ctx)
		})
	})
}

// Route GET /api/welcomer/preview/:key.
func getGuildWelcomerPreview(ctx *gin.Context) {
	key := ctx.Param(KeyKey)

	uuid := uuid.UUID{}

	err := uuid.UnmarshalText(gotils_strconv.S2B(key))
	if err != nil {
		welcomer.Logger.Info().
			Str("key", key).Msg("Failed to unmarshal key to uuid")

		ctx.Data(http.StatusNotFound, "image/png", imageFailure)

		return
	}

	background, err := welcomer.Queries.GetWelcomerImages(ctx, uuid)
	if err != nil || background == nil {
		welcomer.Logger.Info().Str("key", key).Msg("Failed to find welcomer.background with key")

		ctx.Data(http.StatusNotFound, "image/png", imageFailure)

		return
	}

	ctx.Data(http.StatusOK, background.ImageType, background.Data)
}

func welcomerCustomBackgroundsUploadGIF(ctx context.Context, guildID discord.Snowflake, file *multipart.FileHeader, fileBytes io.ReadSeeker, userID discord.Snowflake) (*database.WelcomerImages, error) {
	start := time.Now()

	welcomer.Logger.Info().Int64("size", file.Size).Msg("Recoding image")

	recoderResult, err := recoder.RecodeImage(fileBytes, RecoderQuantizationAttributes)
	if err != nil {
		return nil, err
	}

	welcomer.Logger.Info().Dur("time", time.Since(start)).Msg("Recoded image successfully")

	buf := bytes.NewBuffer(nil)

	_, err = buf.ReadFrom(recoderResult)
	if err != nil {
		return nil, err
	}

	var welcomerImageUUID uuid.UUID
	welcomerImageUUID, _ = gen.NewV7()

	return welcomer.CreateWelcomerImagesWithAudit(ctx, database.CreateWelcomerImagesParams{
		ImageUuid: welcomerImageUUID,
		GuildID:   int64(guildID),
		CreatedAt: time.Now(),
		ImageType: welcomer.ImageFileTypeImageGif.String(),
		Data:      buf.Bytes(),
	}, userID)
}

func welcomerCustomBackgroundsUploadStatic(ctx context.Context, guildID discord.Snowflake, fileBytes io.ReadSeeker, userID discord.Snowflake) (*database.WelcomerImages, error) {
	// Validate file and get size
	img, _, err := image.Decode(fileBytes)
	if err != nil {
		welcomer.Logger.Info().Err(err).Msg("Failed to decode image")

		return nil, ErrFileNotSupported
	}

	_, _ = fileBytes.Seek(0, 0)

	// Validate image resolution
	imageSize := img.Bounds().Size()
	if (imageSize.X * imageSize.Y) > MaxFileResolution {
		welcomer.Logger.Info().
			Int("width", imageSize.X).Int("height", imageSize.Y).
			Int("total", (imageSize.X*imageSize.Y)).Int("max", MaxFileResolution).
			Msg("Rejected image due to resolution")

		return nil, ErrFileSizeTooLarge
	}

	buf := bytes.NewBuffer(nil)

	err = png.Encode(buf, img)
	if err != nil {
		return nil, err
	}

	var welcomerImageUUID uuid.UUID
	welcomerImageUUID, _ = gen.NewV7()

	return welcomer.CreateWelcomerImagesWithAudit(ctx, database.CreateWelcomerImagesParams{
		ImageUuid: welcomerImageUUID,
		GuildID:   int64(guildID),
		CreatedAt: time.Now(),
		ImageType: welcomer.ImageFileTypeImagePng.String(),
		Data:      buf.Bytes(),
	}, userID)
}

// Validates welcomer guild settings.
func doValidateWelcomer(guildSettings *GuildSettingsWelcomer) error {
	if guildSettings.Text.MessageFormat != "" {
		if !welcomer.IsValidEmbed(guildSettings.Text.MessageFormat) {
			return fmt.Errorf("text message is invalid: %w", ErrInvalidJSON)
		}
	}

	if guildSettings.DMs.MessageFormat != "" {
		if !welcomer.IsValidEmbed(guildSettings.DMs.MessageFormat) {
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

		if !welcomer.IsValidInteger(*guildSettings.Text.Channel) {
			return fmt.Errorf("text channel is invalid: %w", ErrChannelInvalid)
		}
	}

	if guildSettings.Images.ToggleEnabled {
		if !welcomer.IsValidBackground(guildSettings.Images.BackgroundName) {
			return fmt.Errorf("image background is invalid: %w", ErrInvalidBackground)
		}

		if !welcomer.IsValidColour(guildSettings.Images.ColourText) {
			return fmt.Errorf("image text colour is invalid: %w", ErrInvalidColour)
		}

		if !welcomer.IsValidColour(guildSettings.Images.ColourTextBorder) {
			return fmt.Errorf("image text border colour is invalid: %w", ErrInvalidColour)
		}

		if !welcomer.IsValidColour(guildSettings.Images.ColourImageBorder) {
			return fmt.Errorf("image border colour is invalid: %w", ErrInvalidColour)
		}

		if !welcomer.IsValidColour(guildSettings.Images.ColourProfileBorder) {
			return fmt.Errorf("image profile border colour is invalid: %w", ErrInvalidColour)
		}

		if !welcomer.IsValidImageAlignment(guildSettings.Images.ImageAlignment) {
			return fmt.Errorf("image ImageAlignment is invalid: %w", ErrInvalidImageAlignment)
		}

		if !welcomer.IsValidImageProfileBorderType(guildSettings.Images.ImageProfileBorderType) {
			return fmt.Errorf("image ImageProfileBorderType is invalid: %w", ErrInvalidProfileBorderType)
		}

		if !welcomer.IsValidImageTheme(guildSettings.Images.ImageTheme) {
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
