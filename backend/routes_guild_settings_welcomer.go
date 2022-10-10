package backend

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/WelcomerTeam/Welcomer/welcomer"
	"github.com/WelcomerTeam/Welcomer/welcomer/database"
	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
)

// GET /api/guild/:guildID/welcomer
func getGuildSettingsWelcomer(ctx *gin.Context) {
	requireOAuthAuthorization(ctx, func(ctx *gin.Context) {
		requireGuildElevation(ctx, func(ctx *gin.Context) {
			guildID := tryGetGuildID(ctx)

			welcomerText, err := backend.Database.GetWelcomerTextGuildSettings(ctx, int64(guildID))
			if err != nil {
				backend.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to get guild welcomer text settings")
			}

			welcomerImages, err := backend.Database.GetWelcomerImagesGuildSettings(ctx, int64(guildID))
			if err != nil {
				backend.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to get guild welcomer images settings")
			}
			welcomerDMs, err := backend.Database.GetWelcomerDMsGuildSettings(ctx, int64(guildID))
			if err != nil {
				backend.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to get guild welcomer dms settings")
			}

			guildBackgrounds, err := backend.Database.GetWelcomerBackgroundByGuildID(ctx, int64(guildID))

			customIDs := make([]string, 0, len(guildBackgrounds))

			for _, b := range guildBackgrounds {
				customIDs = append(customIDs, b.ImageUuid.String())
			}

			partial := GuildSettingsWelcomerSettingsToPartial(welcomerText, welcomerImages, welcomerDMs, &GuildSettingsWelcomerCustom{
				CustomBackgroundIDs: customIDs,
			})

			ctx.JSON(http.StatusOK, BaseResponse{
				Ok:   true,
				Data: partial,
			})
		})
	})
}

// POST /api/guild/:guildID/welcomer
func setGuildSettingsWelcomer(ctx *gin.Context) {
	requireOAuthAuthorization(ctx, func(ctx *gin.Context) {
		requireGuildElevation(ctx, func(ctx *gin.Context) {
			partial := &GuildSettingsWelcomer{}

			var file *multipart.FileHeader
			var err error

			switch ctx.ContentType() {
			case gin.MIMEMultipartPOSTForm:
				multipart, err := ctx.MultipartForm()
				if err == nil {
					file = multipart.File["file"][0]
					json := multipart.Value["json"][0]

					err = jsoniter.UnmarshalFromString(json, &partial)
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

			err = ensureGuild(ctx, guildID)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, BaseResponse{
					Ok:    false,
					Error: ErrEnsureFailure.Error(),
				})

				return
			}

			welcomerText, welcomerImages, welcomerDMs := PartialToGuildSettingsWelcomerSettings(int64(guildID), partial)

			// if welcomerImages.BackgroundName ==
			println(file)
			if file != nil {
				println(file.Filename, file.Header, file.Size)
			}

			databaseWelcomerTextGuildSettings := database.CreateOrUpdateWelcomerTextGuildSettingsParams(*welcomerText)
			welcomerText, err = backend.Database.CreateOrUpdateWelcomerTextGuildSettings(ctx, &databaseWelcomerTextGuildSettings)
			if err != nil {
				backend.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to create or update guild welcomer text settings")
			}

			databaseWelcomerImagesGuildSettings := database.CreateOrUpdateWelcomerImagesGuildSettingsParams(*welcomerImages)
			welcomerImages, err = backend.Database.CreateOrUpdateWelcomerImagesGuildSettings(ctx, &databaseWelcomerImagesGuildSettings)
			if err != nil {
				backend.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to create or update guild welcomer images settings")
			}

			databaseWelcomerDMsGuildSettings := database.CreateOrUpdateWelcomerDMsGuildSettingsParams(*welcomerDMs)
			welcomerDMs, err = backend.Database.CreateOrUpdateWelcomerDMsGuildSettings(ctx, &databaseWelcomerDMsGuildSettings)
			if err != nil {
				backend.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to create or update guild welcomer dms settings")
			}

			getGuildSettingsWelcomer(ctx)
		})
	})
}

// GET /api/guild/:guildID/welcomer/background/:key
func getGuildWelcomerPreview(ctx *gin.Context) {
	rawKey := ctx.Param(KeyKey)

	if strings.TrimSpace(rawKey) == "" {
		ctx.JSON(http.StatusBadRequest, BaseResponse{
			Ok:    false,
			Error: fmt.Sprintf(ErrMissingParameter.Error(), KeyKey),
			Data:  nil,
		})

		return
	}

	println(rawKey)
}

// Validates welcomer guild settings
func doValidateWelcomer(guildSettings *GuildSettingsWelcomer) error {
	if guildSettings.DMs.MessageFormat != "" {
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
