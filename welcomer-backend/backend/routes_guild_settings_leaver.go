package backend

import (
	_ "embed"
	"net/http"

	discord "github.com/WelcomerTeam/Discord/discord"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
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
				leaver = &database.GuildSettingsLeaver{}
			}

			partial := GuildSettingsLeaverSettingsToPartial(*leaver)

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

			leaver := PartialToGuildSettingsLeaverSettings(int64(guildID), partial)

			databaseLeaverGuildSettings := database.CreateOrUpdateLeaverGuildSettingsParams(*leaver)

			err = welcomer.RetryWithFallback(
				func() error {
					_, err = backend.Database.CreateOrUpdateLeaverGuildSettings(ctx, databaseLeaverGuildSettings)
					return err
				},
				func() error {
					return welcomer.EnsureGuild(ctx, backend.Database, discord.Snowflake(guildID))
				},
				nil,
			)
			if err != nil {
				backend.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to create or update guild leaver settings")

				ctx.JSON(http.StatusInternalServerError, BaseResponse{
					Ok: false,
				})

				return
			}

			getGuildSettingsLeaver(ctx)
		})
	})
}

// Validates leaver settings
func doValidateLeaver(guildSettings *GuildSettingsLeaver) error {
	// TODO: validate leaver

	return nil
}

func registerGuildSettingsLeaverRoutes(g *gin.Engine) {
	g.GET("/api/guild/:guildID/leaver", getGuildSettingsLeaver)
	g.POST("/api/guild/:guildID/leaver", setGuildSettingsLeaver)
}
