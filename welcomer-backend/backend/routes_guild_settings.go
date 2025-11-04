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

// Route GET /api/guild/:guildID/settings.
func getGuildSettingsSettings(ctx *gin.Context) {
	requireOAuthAuthorization(ctx, func(ctx *gin.Context) {
		requireGuildElevation(ctx, func(ctx *gin.Context) {
			guildID := tryGetGuildID(ctx)

			settings, err := welcomer.Queries.GetGuild(ctx, int64(guildID))
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					settings = &database.Guilds{
						GuildID:          int64(guildID),
						EmbedColour:      welcomer.DefaultGuild.EmbedColour,
						SiteSplashUrl:    welcomer.DefaultGuild.SiteSplashUrl,
						SiteStaffVisible: welcomer.DefaultGuild.SiteStaffVisible,
						SiteGuildVisible: welcomer.DefaultGuild.SiteGuildVisible,
						SiteAllowInvites: welcomer.DefaultGuild.SiteAllowInvites,
						MemberCount:      0,
						NumberLocale:     welcomer.DefaultGuild.NumberLocale,
					}
				} else {
					welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to get guild settings settings")

					ctx.JSON(http.StatusInternalServerError, BaseResponse{
						Ok: false,
					})

					return
				}
			}

			partial := GuildSettingsToPartial(settings)

			ctx.JSON(http.StatusOK, BaseResponse{
				Ok:   true,
				Data: partial,
			})
		})
	})
}

// Route POST /api/guild/:guildID/settings.
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

			databaseGuildSettings := *settings

			user := tryGetUser(ctx)
			welcomer.Logger.Info().Int64("guild_id", int64(guildID)).Interface("obj", *settings).Int64("user_id", int64(user.ID)).Msg("Creating or updating guild settings settings")

			_, err = welcomer.UpdateGuildWithAudit(ctx, databaseGuildSettings, user.ID)
			if err != nil {
				welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to update guild settings")

				ctx.JSON(http.StatusInternalServerError, NewBaseResponse(NewGenericErrorWithLineNumber(), nil))

				return
			}

			getGuildSettingsSettings(ctx)
		})
	})
}

// Route POST /api/guild/:guildID/settings/update-member-count
func updateGuildSettingsMemberCount(ctx *gin.Context) {
	requireOAuthAuthorization(ctx, func(ctx *gin.Context) {
		requireGuildElevation(ctx, func(ctx *gin.Context) {
			partial := &GuildSettingsUpdateMemberCount{}

			var err error

			err = ctx.BindJSON(partial)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, BaseResponse{
					Ok:    false,
					Error: err.Error(),
				})

				return
			}

			guildID := tryGetGuildID(ctx)

			user := tryGetUser(ctx)
			welcomer.Logger.Info().Int64("guild_id", int64(guildID)).Int32("member_count", partial.Value).Int64("user_id", int64(user.ID)).Msg("Updating guild member count")

			_, err = welcomer.Queries.SetGuildMemberCount(ctx, database.SetGuildMemberCountParams{
				GuildID:     int64(guildID),
				MemberCount: partial.Value,
			})
			if err != nil {
				welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to update guild member count")

				ctx.JSON(http.StatusInternalServerError, NewBaseResponse(NewGenericErrorWithLineNumber(), nil))

				return
			}

			ctx.JSON(http.StatusOK, NewBaseResponse(nil, nil))
		})
	})
}

// Validates settings.
func doValidateSettings(guildSettings *GuildSettingsSettings) error {
	// TODO: validate settings

	if _, err := database.ParseNumberLocale(guildSettings.NumberLocale); err != nil {
		return NewInvalidParameterError("number_locale")
	}

	return nil
}

func registerGuildSettingsRoutes(g *gin.Engine) {
	g.GET("/api/guild/:guildID/settings", getGuildSettingsSettings)
	g.POST("/api/guild/:guildID/settings", setGuildSettingsSettings)
	g.POST("/api/guild/:guildID/settings/update-member-count", updateGuildSettingsMemberCount)
}
