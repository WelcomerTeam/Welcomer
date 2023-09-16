package backend

import (
	_ "embed"
	"net/http"

	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	"github.com/gin-gonic/gin"
)

// Route GET /api/guild/:guildID/leaver
func getGuildSettingsLeaver(ctx *gin.Context) {
	requireOAuthAuthorization(ctx, func(ctx *gin.Context) {
		requireGuildElevation(ctx, func(ctx *gin.Context) {
			guildID := tryGetGuildID(ctx)

			leaver, err := backend.Database.GetLeaverGuildSettings(ctx, int64(guildID))
			if err != nil {
				backend.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to get guild leaver settings")
			}

			partial := GuildSettingsLeaverSettingsToPartial(leaver)

			ctx.JSON(http.StatusOK, BaseResponse{
				Ok:   true,
				Data: partial,
			})
		})
	})
}

// Route POST /api/guild/:guildID/leaver
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

			err = ensureGuild(ctx, guildID)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, BaseResponse{
					Ok:    false,
					Error: ErrEnsureFailure.Error(),
				})

				return
			}

			leaver := PartialToGuildSettingsLeaverSettings(int64(guildID), partial)

			databaseLeaverGuildSettings := database.CreateOrUpdateLeaverGuildSettingsParams(*leaver)
			_, err = backend.Database.CreateOrUpdateLeaverGuildSettings(ctx, &databaseLeaverGuildSettings)
			if err != nil {
				backend.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to create or update guild leaver settings")
			}

			getGuildSettingsLeaver(ctx)
		})
	})
}

// Validates leaver settings
func doValidateLeaver(guildSettings *GuildSettingsLeaver) error {
	// validate leaver

	return nil
}

func registerGuildSettingsLeaverRoutes(g *gin.Engine) {
	g.GET("/api/guild/:guildID/leaver", getGuildSettingsLeaver)
	g.POST("/api/guild/:guildID/leaver", setGuildSettingsLeaver)
}
