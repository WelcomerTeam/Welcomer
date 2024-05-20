package backend

import (
	_ "embed"
	"net/http"

	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	"github.com/gin-gonic/gin"
)

// Route GET /api/guild/:guildID/settings
func getGuildSettingsSettings(ctx *gin.Context) {
	requireOAuthAuthorization(ctx, func(ctx *gin.Context) {
		requireGuildElevation(ctx, func(ctx *gin.Context) {
			guildID := tryGetGuildID(ctx)

			settings, err := backend.Database.GetGuild(ctx, int64(guildID))
			if err != nil {
				backend.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to get guild settings settings")
			}

			partial := GuildSettingsToPartial(settings)

			ctx.JSON(http.StatusOK, BaseResponse{
				Ok:   true,
				Data: partial,
			})
		})
	})
}

// Route POST /api/guild/:guildID/settings
func setGuildSettingsSettings(ctx *gin.Context) {
	requireOAuthAuthorization(ctx, func(ctx *gin.Context) {
		requireGuildElevation(ctx, func(ctx *gin.Context) {
			partial := &GuildSettingsSettings{}

			var err error

			err = ctx.BindJSON(partial)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, BaseResponse{
					Ok:    false,
					Error: err.Error(),
				})

				return
			}

			err = doValidateSettings(partial)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, BaseResponse{
					Ok:    false,
					Error: err.Error(),
				})

				return
			}

			guildID := tryGetGuildID(ctx)

			settings := PartialToGuildSettings(int64(guildID), partial)

			databaseGuildSettings := database.CreateOrUpdateGuildParams(*settings)

			_, err = backend.Database.CreateOrUpdateGuild(ctx, databaseGuildSettings)
			if err != nil {
				backend.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to create or update guild settings settings")
			}

			getGuildSettingsSettings(ctx)
		})
	})
}

// Validates settings
func doValidateSettings(guildSettings *GuildSettingsSettings) error {
	// TODO: validate settings

	return nil
}

func registerGuildSettingsRoutes(g *gin.Engine) {
	g.GET("/api/guild/:guildID/settings", getGuildSettingsSettings)
	g.POST("/api/guild/:guildID/settings", setGuildSettingsSettings)
}
