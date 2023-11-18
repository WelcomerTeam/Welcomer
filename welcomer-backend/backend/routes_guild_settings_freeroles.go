package backend

import (
	_ "embed"
	"net/http"

	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	"github.com/gin-gonic/gin"
)

// Route GET /api/guild/:guildID/freeroles
func getGuildSettingsFreeRoles(ctx *gin.Context) {
	requireOAuthAuthorization(ctx, func(ctx *gin.Context) {
		requireGuildElevation(ctx, func(ctx *gin.Context) {
			guildID := tryGetGuildID(ctx)

			freeroles, err := backend.Database.GetFreeRolesGuildSettings(ctx, int64(guildID))
			if err != nil {
				backend.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to get guild freeroles settings")
			}

			partial := GuildSettingsFreeRolesSettingsToPartial(freeroles)

			ctx.JSON(http.StatusOK, BaseResponse{
				Ok:   true,
				Data: partial,
			})
		})
	})
}

// Route POST /api/guild/:guildID/freeroles
func setGuildSettingsFreeRoles(ctx *gin.Context) {
	requireOAuthAuthorization(ctx, func(ctx *gin.Context) {
		requireGuildElevation(ctx, func(ctx *gin.Context) {
			partial := &GuildSettingsFreeRoles{}

			var err error

			err = ctx.BindJSON(partial)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, BaseResponse{
					Ok:    false,
					Error: err.Error(),
				})

				return
			}

			err = doValidateFreeRoles(partial)
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

			freeroles := PartialToGuildSettingsFreeRolesSettings(int64(guildID), partial)

			databaseFreeRolesGuildSettings := database.CreateOrUpdateFreeRolesGuildSettingsParams(*freeroles)
			_, err = backend.Database.CreateOrUpdateFreeRolesGuildSettings(ctx, &databaseFreeRolesGuildSettings)
			if err != nil {
				backend.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to create or update guild freeroles settings")
			}

			getGuildSettingsFreeRoles(ctx)
		})
	})
}

// Validates freerole settings
func doValidateFreeRoles(guildSettings *GuildSettingsFreeRoles) error {
	// validate freeroles

	return nil
}

func registerGuildSettingsFreeRolesRoutes(g *gin.Engine) {
	g.GET("/api/guild/:guildID/freeroles", getGuildSettingsFreeRoles)
	g.POST("/api/guild/:guildID/freeroles", setGuildSettingsFreeRoles)
}
