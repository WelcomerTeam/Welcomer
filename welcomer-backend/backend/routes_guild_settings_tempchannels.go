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

// Route GET /api/guild/:guildID/tempchannels.
func getGuildSettingsTempChannels(ctx *gin.Context) {
	requireOAuthAuthorization(ctx, func(ctx *gin.Context) {
		requireGuildElevation(ctx, func(ctx *gin.Context) {
			guildID := tryGetGuildID(ctx)

			tempchannels, err := welcomer.Queries.GetTempChannelsGuildSettings(ctx, int64(guildID))
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					tempchannels = &database.GuildSettingsTempchannels{
						GuildID:          int64(guildID),
						ToggleEnabled:    welcomer.DefaultTempChannels.ToggleEnabled,
						ToggleAutopurge:  welcomer.DefaultTempChannels.ToggleAutopurge,
						ChannelLobby:     welcomer.DefaultTempChannels.ChannelLobby,
						ChannelCategory:  welcomer.DefaultTempChannels.ChannelCategory,
						DefaultUserCount: welcomer.DefaultTempChannels.DefaultUserCount,
					}
				} else {
					welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to get guild tempchannels settings")

					ctx.JSON(http.StatusInternalServerError, BaseResponse{
						Ok: false,
					})

					return
				}
			}

			partial := GuildSettingsTempChannelsSettingsToPartial(tempchannels)

			ctx.JSON(http.StatusOK, BaseResponse{
				Ok:   true,
				Data: partial,
			})
		})
	})
}

// Route POST /api/guild/:guildID/tempchannels.
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

			tempchannels := PartialToGuildSettingsTempChannelsSettings(int64(guildID), partial)

			databaseTempChannelsGuildSettings := database.CreateOrUpdateTempChannelsGuildSettingsParams(*tempchannels)

			user := tryGetUser(ctx)
			welcomer.Logger.Info().Int64("guild_id", int64(guildID)).Interface("obj", *tempchannels).Int64("user_id", int64(user.ID)).Msg("Creating or updating guild tempchannel settings")

			err = welcomer.RetryWithFallback(
				func() error {
					_, err = welcomer.Queries.CreateOrUpdateTempChannelsGuildSettings(ctx, databaseTempChannelsGuildSettings)
					return err
				},
				func() error {
					return welcomer.EnsureGuild(ctx, guildID)
				},
				nil,
			)
			if err != nil {
				welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to create or update guild tempchannels settings")

				ctx.JSON(http.StatusInternalServerError, BaseResponse{
					Ok: false,
				})

				return
			}

			getGuildSettingsTempChannels(ctx)
		})
	})
}

// Validates tempchannel settings.
func doValidateTempChannels(guildSettings *GuildSettingsTempChannels) error {
	// TODO: validate tempchannels

	return nil
}

func registerGuildSettingsTempChannelsRoutes(g *gin.Engine) {
	g.GET("/api/guild/:guildID/tempchannels", getGuildSettingsTempChannels)
	g.POST("/api/guild/:guildID/tempchannels", setGuildSettingsTempChannels)
}
