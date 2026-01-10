package backend

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"net/http"
	"strings"

	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/webp"

	"github.com/WelcomerTeam/Discord/discord"
	sandwich "github.com/WelcomerTeam/Sandwich-Daemon/proto"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	"github.com/gin-gonic/gin"
)

const (
	MaxAvatarSize = 5_000_000  // 5MB file size.
	MaxBannerSize = 10_000_000 // 10MB file size.
)

// Route GET /api/guild/:guildID/customisation
func getGuildSettingsCustomisation(ctx *gin.Context) {
	requireOAuthAuthorization(ctx, func(ctx *gin.Context) {
		requireGuildElevation(ctx, func(ctx *gin.Context) {
			guildID := tryGetGuildID(ctx)

			managerName := welcomer.DefaultManagerName

			// Fetch bot application
			applications, err := welcomer.SandwichClient.FetchApplication(ctx, &sandwich.ApplicationIdentifier{
				ApplicationIdentifier: managerName,
			})
			if err != nil || len(applications.GetApplications()) == 0 {
				welcomer.Logger.Error().Err(err).
					Int64("guild_id", int64(guildID)).
					Str("manager_name", managerName).
					Msg("Failed to fetch application for guild customisation")

				ctx.JSON(http.StatusInternalServerError, NewGenericErrorWithLineNumber())

				return
			}

			botID := discord.Snowflake(applications.GetApplications()[managerName].GetUserId())
			if botID.IsNil() {
				welcomer.Logger.Error().
					Int64("guild_id", int64(guildID)).
					Str("manager_name", managerName).
					Msg("Bot ID is nil for fetched application for guild customisation")

				ctx.JSON(http.StatusInternalServerError, NewGenericErrorWithLineNumber())

				return
			}

			member, err := discord.GetGuildMember(ctx, backend.BotSession, guildID, botID)
			if err != nil {
				welcomer.Logger.Error().Err(err).
					Int64("guild_id", int64(guildID)).
					Msg("Failed to get guild member for guild customisation")

				ctx.JSON(http.StatusInternalServerError, NewGenericErrorWithLineNumber())

				return
			}

			guildSettings, err := welcomer.Queries.GetGuild(ctx, int64(guildID))
			if err != nil {
				welcomer.Logger.Warn().Err(err).
					Int64("guild_id", int64(guildID)).
					Msg("Failed to get guild settings for guild customisation")
			}

			ctx.JSON(http.StatusOK, BaseResponse{
				Ok: true,
				Data: GuildSettingsCustomisation{
					Nickname: &member.Nick,
					Avatar:   &member.Avatar,
					Banner:   &member.Banner,
					Bio:      &guildSettings.Bio,
					UserID:   botID,
				},
			})
		})
	})
}

// Route POST /api/guild/:guildID/customisation
func setGuildSettingsCustomisation(ctx *gin.Context) {
	requireOAuthAuthorization(ctx, func(ctx *gin.Context) {
		requireGuildElevation(ctx, func(ctx *gin.Context) {
			guildID := tryGetGuildID(ctx)
			userID := tryGetUser(ctx).ID

			partial := &GuildSettingsCustomisation{}

			hasWelcomerPro, _, _, err := welcomer.CheckGuildMemberships(ctx, guildID)
			if err != nil {
				welcomer.Logger.Error().Err(err).
					Int64("guild_id", int64(guildID)).
					Msg("Failed to check guild memberships for guild customisation")

				ctx.JSON(http.StatusInternalServerError, NewGenericErrorWithLineNumber())

				return
			}

			if !hasWelcomerPro {
				ctx.JSON(http.StatusPaymentRequired, BaseResponse{
					Ok:    false,
					Error: ErrMissingMembership.Error(),
				})

				return
			}

			if err := ctx.BindJSON(partial); err != nil {
				ctx.JSON(http.StatusBadRequest, BaseResponse{
					Ok:    false,
					Error: err.Error(),
				})

				return
			}

			err = doValidateCustomisation(partial)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, BaseResponse{
					Ok:    false,
					Error: err.Error(),
				})

				return
			}

			var modifyCurrentMemberParams discord.ModifyCurrentMemberParams

			if partial.Nickname != nil {
				modifyCurrentMemberParams.Nick = partial.Nickname
			}

			if partial.Avatar != nil {
				if *partial.Avatar != "" {
					avatarData, err := decodeBase64Image(*partial.Avatar)
					if err != nil {
						ctx.JSON(http.StatusBadRequest, BaseResponse{
							Ok:    false,
							Error: "Invalid avatar image data",
						})

						return
					}

					if len(avatarData) > MaxAvatarSize {
						ctx.JSON(http.StatusBadRequest, BaseResponse{
							Ok:    false,
							Error: ErrFileSizeTooLarge.Error(),
						})

						return
					}

					if err := doValidateImageForCustomisation(avatarData, 1024, 1024); err != nil {
						ctx.JSON(http.StatusBadRequest, BaseResponse{
							Ok:    false,
							Error: err.Error(),
						})

						return
					}
				}

				modifyCurrentMemberParams.Avatar = partial.Avatar
			}

			if partial.Banner != nil {
				if *partial.Banner != "" {
					bannerData, err := decodeBase64Image(*partial.Banner)
					if err != nil {
						ctx.JSON(http.StatusBadRequest, BaseResponse{
							Ok:    false,
							Error: "Invalid banner image data",
						})

						return
					}

					if len(bannerData) > MaxBannerSize {
						ctx.JSON(http.StatusBadRequest, BaseResponse{
							Ok:    false,
							Error: ErrFileSizeTooLarge.Error(),
						})

						return
					}

					if err := doValidateImageForCustomisation(bannerData, 1024, 256); err != nil {
						ctx.JSON(http.StatusBadRequest, BaseResponse{
							Ok:    false,
							Error: err.Error(),
						})

						return
					}
				}

				modifyCurrentMemberParams.Banner = partial.Banner
			}

			if partial.Bio != nil {
				modifyCurrentMemberParams.Bio = partial.Bio
			}

			_, err = discord.ModifyCurrentMember(ctx,
				backend.BotSession,
				guildID,
				modifyCurrentMemberParams,
				welcomer.ToPointer("Updated bot customisation"),
			)
			if err != nil {
				welcomer.Logger.Error().Err(err).
					Int64("guild_id", int64(guildID)).
					Msg("Failed to modify current member for guild customisation")

				ctx.JSON(http.StatusInternalServerError, NewGenericErrorWithLineNumber())

				return
			}

			welcomer.AuditChange(ctx, guildID, userID, nil, partial, database.AuditTypeBotCustomisation)

			if partial.Bio != nil {
				_, err = welcomer.UpdateBioWithAudit(ctx, database.UpdateGuildBioParams{
					GuildID: int64(guildID),
					Bio:     *partial.Bio,
				}, userID)
				if err != nil {
					welcomer.Logger.Error().Err(err).
						Int64("guild_id", int64(guildID)).
						Msg("Failed to update guild bio for guild customisation")
				}
			}

			getGuildSettingsCustomisation(ctx)
		})
	})
}

func doValidateCustomisation(partial *GuildSettingsCustomisation) error {
	if partial == nil {
		return nil
	}

	if partial.Bio != nil {
		if len(*partial.Bio) > 190 {
			return fmt.Errorf("bio exceeds maximum length of 190 characters")
		}
	}

	if partial.Nickname != nil {
		if len(*partial.Nickname) > 32 {
			return fmt.Errorf("nickname exceeds maximum length of 32 characters")
		}
	}

	return nil
}

func doValidateImageForCustomisation(data []byte, maxWidth, maxHeight int) error {
	mimeType := http.DetectContentType(data)

	switch mimeType {
	case MIMEPNG, MIMEJPEG, MIMEWEBP:
		// Supported formats
	default:
		return ErrFileNotSupported
	}

	im, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	bounds := im.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Check dimensions
	if width > maxWidth || height > maxHeight {
		return ErrResolutionTooHigh
	}

	return nil
}

func decodeBase64Image(data string) ([]byte, error) {
	commaIndex := strings.Index(data, ",")
	if commaIndex == -1 {
		return nil, fmt.Errorf("invalid base64 image data")
	}

	decodedData, err := base64.StdEncoding.DecodeString(data[commaIndex+1:])
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 image data: %w", err)
	}

	return decodedData, nil
}

func registerGuildSettingsCustomisationRoutes(g *gin.Engine) {
	g.GET("/api/guild/:guildID/customisation", getGuildSettingsCustomisation)
	g.POST("/api/guild/:guildID/customisation", setGuildSettingsCustomisation)
}
