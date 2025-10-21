package backend

import (
	_ "embed"
	"errors"
	"net/http"

	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4"
)

// Route GET /api/guild/:guildID/leaver.
func getGuildSettingsLeaver(ctx *gin.Context) {
	requireOAuthAuthorization(ctx, func(ctx *gin.Context) {
		requireGuildElevation(ctx, func(ctx *gin.Context) {
			guildID := tryGetGuildID(ctx)

			leaver, err := welcomer.Queries.GetLeaverGuildSettings(ctx, int64(guildID))
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					leaver = &database.GuildSettingsLeaver{
						GuildID:                  int64(guildID),
						ToggleEnabled:            welcomer.DefaultLeaver.ToggleEnabled,
						Channel:                  welcomer.DefaultLeaver.Channel,
						MessageFormat:            welcomer.DefaultLeaver.MessageFormat,
						AutoDeleteLeaverMessages: welcomer.DefaultLeaver.AutoDeleteLeaverMessages,
						LeaverMessageLifetime:    welcomer.DefaultLeaver.LeaverMessageLifetime,
					}
				} else {
					welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to get guild leaver settings")

					ctx.JSON(http.StatusInternalServerError, BaseResponse{
						Ok: false,
					})

					return
				}
			}

			partial := GuildSettingsLeaverSettingsToPartial(*leaver)

			ctx.JSON(http.StatusOK, BaseResponse{
				Ok:   true,
				Data: partial,
			})
		})
	})
}

// Route POST /api/guild/:guildID/leaver.
func setGuildSettingsLeaver(ctx *gin.Context) {
	requireOAuthAuthorization(ctx, func(ctx *gin.Context) {
		requireGuildElevation(ctx, func(ctx *gin.Context) {
			partial := &GuildSettingsLeaver{}

			var err error

			err = ctx.BindJSON(partial)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, BaseResponse{
					Ok:    false,
					Error: err.Error(),
				})

				return
			}

			err = doValidateLeaver(partial)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, BaseResponse{
					Ok:    false,
					Error: err.Error(),
				})

				return
			}

			guildID := tryGetGuildID(ctx)

			leaver := PartialToGuildSettingsLeaverSettings(int64(guildID), partial)

			databaseLeaverGuildSettings := database.CreateOrUpdateLeaverGuildSettingsParams(*leaver)

			user := tryGetUser(ctx)
			welcomer.Logger.Info().Int64("guild_id", int64(guildID)).Interface("obj", *leaver).Int64("user_id", int64(user.ID)).Msg("Creating or updating guild leaver settings")

			err = welcomer.RetryWithFallback(
				func() error {
					_, err = welcomer.Queries.CreateOrUpdateLeaverGuildSettings(ctx, databaseLeaverGuildSettings)

					return err
				},
				func() error {
					return welcomer.EnsureGuild(ctx, guildID)
				},
				nil,
			)
			if err != nil {
				welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to create or update guild leaver settings")

				ctx.JSON(http.StatusInternalServerError, BaseResponse{
					Ok: false,
				})

				return
			}

			getGuildSettingsLeaver(ctx)
		})
	})
}

// Validates leaver settings.
func doValidateLeaver(guildSettings *GuildSettingsLeaver) error {
	// TODO: validate leaver

	return nil
}

func registerGuildSettingsLeaverRoutes(g *gin.Engine) {
	g.GET("/api/guild/:guildID/leaver", getGuildSettingsLeaver)
	g.POST("/api/guild/:guildID/leaver", setGuildSettingsLeaver)
}
