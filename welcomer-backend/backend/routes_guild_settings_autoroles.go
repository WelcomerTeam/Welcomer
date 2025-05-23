package backend

import (
	"errors"
	"net/http"

	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4"
)

// Route GET /api/guild/:guildID/autoroles.
func getGuildSettingsAutoRoles(ctx *gin.Context) {
	requireOAuthAuthorization(ctx, func(ctx *gin.Context) {
		requireGuildElevation(ctx, func(ctx *gin.Context) {
			guildID := tryGetGuildID(ctx)

			autoroles, err := welcomer.Queries.GetAutoRolesGuildSettings(ctx, int64(guildID))
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					autoroles = &database.GuildSettingsAutoroles{
						GuildID:       int64(guildID),
						ToggleEnabled: welcomer.DefaultAutoroles.ToggleEnabled,
						Roles:         welcomer.DefaultAutoroles.Roles,
					}
				} else {
					welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to get guild autoroles settings")

					ctx.JSON(http.StatusInternalServerError, BaseResponse{
						Ok: false,
					})

					return
				}
			}

			partial := GuildSettingsAutoRolesSettingsToPartial(autoroles)

			ctx.JSON(http.StatusOK, BaseResponse{
				Ok:   true,
				Data: partial,
			})
		})
	})
}

// Route POST /api/guild/:guildID/autoroles.
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

			autoroles := PartialToGuildSettingsAutoRolesSettings(int64(guildID), partial)

			databaseAutoRolesGuildSettings := database.CreateOrUpdateAutoRolesGuildSettingsParams(*autoroles)

			user := tryGetUser(ctx)
			welcomer.Logger.Info().Int64("guild_id", int64(guildID)).Interface("obj", *autoroles).Int64("user_id", int64(user.ID)).Msg("Creating or updating guild autoroles settings")

			err = welcomer.RetryWithFallback(
				func() error {
					_, err = welcomer.Queries.CreateOrUpdateAutoRolesGuildSettings(ctx, databaseAutoRolesGuildSettings)

					return err
				},
				func() error {
					return welcomer.EnsureGuild(ctx, guildID)
				},
				nil,
			)
			if err != nil {
				welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to create or update guild autoroles settings")

				ctx.JSON(http.StatusInternalServerError, BaseResponse{
					Ok: false,
				})

				return
			}

			getGuildSettingsAutoRoles(ctx)
		})
	})
}

// Validates autorole settings.
func doValidateAutoRoles(guildSettings *GuildSettingsAutoRoles) error {
	// TODO: validate autoroles

	return nil
}

func registerGuildSettingsAutoRolesRoutes(g *gin.Engine) {
	g.GET("/api/guild/:guildID/autoroles", getGuildSettingsAutoRoles)
	g.POST("/api/guild/:guildID/autoroles", setGuildSettingsAutoRoles)
}
