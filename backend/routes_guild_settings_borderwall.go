package backend

import (
	_ "embed"
	"net/http"

	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	"github.com/gin-gonic/gin"
)

// Route GET /api/guild/:guildID/borderwall
func getGuildSettingsBorderwall(ctx *gin.Context) {
	requireOAuthAuthorization(ctx, func(ctx *gin.Context) {
		requireGuildElevation(ctx, func(ctx *gin.Context) {
			guildID := tryGetGuildID(ctx)

			borderwall, err := backend.Database.GetBorderwallGuildSettings(ctx, int64(guildID))
			if err != nil {
				backend.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to get guild borderwall settings")
			}

			partial := GuildSettingsBorderwallSettingsToPartial(borderwall)

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

			err = ensureGuild(ctx, guildID)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, BaseResponse{
					Ok:    false,
					Error: ErrEnsureFailure.Error(),
				})

				return
			}

			borderwall := PartialToGuildSettingsBorderwallSettings(int64(guildID), partial)

			databaseBorderwallGuildSettings := database.CreateOrUpdateBorderwallGuildSettingsParams(*borderwall)
			_, err = backend.Database.CreateOrUpdateBorderwallGuildSettings(ctx, &databaseBorderwallGuildSettings)
			if err != nil {
				backend.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to create or update guild borderwall settings")
			}

			getGuildSettingsBorderwall(ctx)
		})
	})
}

// Validates borderwall settings
func doValidateBorderwall(guildSettings *GuildSettingsBorderwall) error {
	// validate borderwall

	return nil
}

func registerGuildSettingsBorderwallRoutes(g *gin.Engine) {
	g.GET("/api/guild/:guildID/borderwall", getGuildSettingsBorderwall)
	g.POST("/api/guild/:guildID/borderwall", setGuildSettingsBorderwall)
}
