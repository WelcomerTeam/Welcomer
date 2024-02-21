package backend

import (
	_ "embed"
	"net/http"

	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	"github.com/gin-gonic/gin"
)

// Route GET /api/guild/:guildID/timeroles
func getGuildSettingsTimeRoles(ctx *gin.Context) {
	requireOAuthAuthorization(ctx, func(ctx *gin.Context) {
		requireGuildElevation(ctx, func(ctx *gin.Context) {
			guildID := tryGetGuildID(ctx)

			timeroles, err := backend.Database.GetTimeRolesGuildSettings(ctx, int64(guildID))
			if err != nil {
				backend.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to get guild timeroles settings")
			}

			partial := GuildSettingsTimeRolesSettingsToPartial(timeroles)

			ctx.JSON(http.StatusOK, BaseResponse{
				Ok:   true,
				Data: partial,
			})
		})
	})
}

// Route POST /api/guild/:guildID/timeroles
func setGuildSettingsTimeRoles(ctx *gin.Context) {
	requireOAuthAuthorization(ctx, func(ctx *gin.Context) {
		requireGuildElevation(ctx, func(ctx *gin.Context) {
			partial := &GuildSettingsTimeRoles{}

			var err error

			err = ctx.BindJSON(partial)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, BaseResponse{
					Ok:    false,
					Error: err.Error(),
				})

				return
			}

			err = doValidateTimeRoles(partial)
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
					Ok: false,
				})

				return
			}

			timeroles := PartialToGuildSettingsTimeRolesSettings(int64(guildID), partial)

			databaseTimeRolesGuildSettings := database.CreateOrUpdateTimeRolesGuildSettingsParams(*timeroles)
			_, err = backend.Database.CreateOrUpdateTimeRolesGuildSettings(ctx, &databaseTimeRolesGuildSettings)
			if err != nil {
				backend.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to create or update guild timeroles settings")
			}

			getGuildSettingsTimeRoles(ctx)
		})
	})
}

// Validates timerole settings
func doValidateTimeRoles(guildSettings *GuildSettingsTimeRoles) error {
	// TODO: validate timeroles

	return nil
}

func registerGuildSettingsTimeRolesRoutes(g *gin.Engine) {
	g.GET("/api/guild/:guildID/timeroles", getGuildSettingsTimeRoles)
	g.POST("/api/guild/:guildID/timeroles", setGuildSettingsTimeRoles)
}
