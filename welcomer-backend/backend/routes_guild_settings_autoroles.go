package backend

import (
	_ "embed"
	"net/http"

	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	"github.com/gin-gonic/gin"
)

// Route GET /api/guild/:guildID/autoroles
func getGuildSettingsAutoRoles(ctx *gin.Context) {
	requireOAuthAuthorization(ctx, func(ctx *gin.Context) {
		requireGuildElevation(ctx, func(ctx *gin.Context) {
			guildID := tryGetGuildID(ctx)

			autoroles, err := backend.Database.GetAutoRolesGuildSettings(ctx, int64(guildID))
			if err != nil {
				backend.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to get guild autoroles settings")
			}

			partial := GuildSettingsAutoRolesSettingsToPartial(autoroles)

			ctx.JSON(http.StatusOK, BaseResponse{
				Ok:   true,
				Data: partial,
			})
		})
	})
}

// Route POST /api/guild/:guildID/autoroles
func setGuildSettingsAutoRoles(ctx *gin.Context) {
	requireOAuthAuthorization(ctx, func(ctx *gin.Context) {
		requireGuildElevation(ctx, func(ctx *gin.Context) {
			partial := &GuildSettingsAutoRoles{}

			var err error

			err = ctx.BindJSON(partial)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, BaseResponse{
					Ok:    false,
					Error: err.Error(),
				})

				return
			}

			err = doValidateAutoRoles(partial)
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

			autoroles := PartialToGuildSettingsAutoRolesSettings(int64(guildID), partial)

			databaseAutoRolesGuildSettings := database.CreateOrUpdateAutoRolesGuildSettingsParams(*autoroles)
			_, err = backend.Database.CreateOrUpdateAutoRolesGuildSettings(ctx, &databaseAutoRolesGuildSettings)
			if err != nil {
				backend.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to create or update guild autoroles settings")
			}

			getGuildSettingsAutoRoles(ctx)
		})
	})
}

// Validates autorole settings
func doValidateAutoRoles(guildSettings *GuildSettingsAutoRoles) error {
	// TODO: validate autoroles

	return nil
}

func registerGuildSettingsAutoRolesRoutes(g *gin.Engine) {
	g.GET("/api/guild/:guildID/autoroles", getGuildSettingsAutoRoles)
	g.POST("/api/guild/:guildID/autoroles", setGuildSettingsAutoRoles)
}
