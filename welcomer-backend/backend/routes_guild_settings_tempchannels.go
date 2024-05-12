package backend

import (
	_ "embed"
	"net/http"

	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	"github.com/gin-gonic/gin"
)

// Route GET /api/guild/:guildID/tempchannels
func getGuildSettingsTempChannels(ctx *gin.Context) {
	requireOAuthAuthorization(ctx, func(ctx *gin.Context) {
		requireGuildElevation(ctx, func(ctx *gin.Context) {
			guildID := tryGetGuildID(ctx)

			tempchannels, err := backend.Database.GetTempChannelsGuildSettings(ctx, int64(guildID))
			if err != nil {
				backend.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to get guild tempchannels settings")
			}

			partial := GuildSettingsTempChannelsSettingsToPartial(tempchannels)

			ctx.JSON(http.StatusOK, BaseResponse{
				Ok:   true,
				Data: partial,
			})
		})
	})
}

// Route POST /api/guild/:guildID/tempchannels
func setGuildSettingsTempChannels(ctx *gin.Context) {
	requireOAuthAuthorization(ctx, func(ctx *gin.Context) {
		requireGuildElevation(ctx, func(ctx *gin.Context) {
			partial := &GuildSettingsTempChannels{}

			var err error

			err = ctx.BindJSON(partial)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, BaseResponse{
					Ok:    false,
					Error: err.Error(),
				})

				return
			}

			err = doValidateTempChannels(partial)
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

			tempchannels := PartialToGuildSettingsTempChannelsSettings(int64(guildID), partial)

			databaseTempChannelsGuildSettings := database.CreateOrUpdateTempChannelsGuildSettingsParams(*tempchannels)
			_, err = backend.Database.CreateOrUpdateTempChannelsGuildSettings(ctx, databaseTempChannelsGuildSettings)
			if err != nil {
				backend.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to create or update guild tempchannels settings")
			}

			getGuildSettingsTempChannels(ctx)
		})
	})
}

// Validates tempchannel settings
func doValidateTempChannels(guildSettings *GuildSettingsTempChannels) error {
	// TODO: validate tempchannels

	return nil
}

func registerGuildSettingsTempChannelsRoutes(g *gin.Engine) {
	g.GET("/api/guild/:guildID/tempchannels", getGuildSettingsTempChannels)
	g.POST("/api/guild/:guildID/tempchannels", setGuildSettingsTempChannels)
}
