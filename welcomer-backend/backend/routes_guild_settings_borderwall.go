package backend

import (
	_ "embed"
	"errors"
	discord "github.com/WelcomerTeam/Discord/discord"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	utils "github.com/WelcomerTeam/Welcomer/welcomer-utils"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4"
	"net/http"
)

// Route GET /api/guild/:guildID/borderwall
func getGuildSettingsBorderwall(ctx *gin.Context) {
	requireOAuthAuthorization(ctx, func(ctx *gin.Context) {
		requireGuildElevation(ctx, func(ctx *gin.Context) {
			guildID := tryGetGuildID(ctx)

			borderwall, err := backend.Database.GetBorderwallGuildSettings(ctx, int64(guildID))
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					borderwall = &database.GuildSettingsBorderwall{
						GuildID:         int64(guildID),
						ToggleEnabled:   database.DefaultBorderwall.ToggleEnabled,
						ToggleSendDm:    database.DefaultBorderwall.ToggleSendDm,
						Channel:         database.DefaultBorderwall.Channel,
						MessageVerify:   database.DefaultBorderwall.MessageVerify,
						MessageVerified: database.DefaultBorderwall.MessageVerified,
						RolesOnJoin:     database.DefaultBorderwall.RolesOnJoin,
						RolesOnVerify:   database.DefaultBorderwall.RolesOnVerify,
					}
				} else {
					backend.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to get guild borderwall settings")

					ctx.JSON(http.StatusInternalServerError, BaseResponse{
						Ok: false,
					})

					return
				}
			}

			partial := GuildSettingsBorderwallSettingsToPartial(*borderwall)

			ctx.JSON(http.StatusOK, BaseResponse{
				Ok:   true,
				Data: partial,
			})
		})
	})
}

// Route POST /api/guild/:guildID/borderwall
func setGuildSettingsBorderwall(ctx *gin.Context) {
	requireOAuthAuthorization(ctx, func(ctx *gin.Context) {
		requireGuildElevation(ctx, func(ctx *gin.Context) {
			partial := &GuildSettingsBorderwall{}

			var err error

			err = ctx.BindJSON(partial)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, BaseResponse{
					Ok:    false,
					Error: err.Error(),
				})

				return
			}

			err = doValidateBorderwall(partial)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, BaseResponse{
					Ok:    false,
					Error: err.Error(),
				})

				return
			}

			guildID := tryGetGuildID(ctx)

			borderwall := PartialToGuildSettingsBorderwallSettings(int64(guildID), partial)

			databaseBorderwallGuildSettings := database.CreateOrUpdateBorderwallGuildSettingsParams(*borderwall)

			user := tryGetUser(ctx)
			backend.Logger.Info().Int64("guild_id", int64(guildID)).Interface("obj", *borderwall).Int64("user_id", int64(user.ID)).Msg("Creating or updating guild borderwall settings")

			err = utils.RetryWithFallback(
				func() error {
					_, err = backend.Database.CreateOrUpdateBorderwallGuildSettings(ctx, databaseBorderwallGuildSettings)
					return err
				},
				func() error {
					return welcomer.EnsureGuild(ctx, backend.Database, discord.Snowflake(guildID))
				},
				nil,
			)
			if err != nil {
				backend.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to create or update guild borderwall settings")

				ctx.JSON(http.StatusInternalServerError, BaseResponse{
					Ok: false,
				})

				return
			}

			getGuildSettingsBorderwall(ctx)
		})
	})
}

// Validates borderwall settings
func doValidateBorderwall(guildSettings *GuildSettingsBorderwall) error {
	// TODO: validate borderwall

	return nil
}

func registerGuildSettingsBorderwallRoutes(g *gin.Engine) {
	g.GET("/api/guild/:guildID/borderwall", getGuildSettingsBorderwall)
	g.POST("/api/guild/:guildID/borderwall", setGuildSettingsBorderwall)
}
