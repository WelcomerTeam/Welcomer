package backend

import (
	_ "embed"
	"errors"
	"net/http"

	discord "github.com/WelcomerTeam/Discord/discord"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	utils "github.com/WelcomerTeam/Welcomer/welcomer-utils"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4"
)

// Route GET /api/guild/:guildID/freeroles
func getGuildSettingsFreeRoles(ctx *gin.Context) {
	requireOAuthAuthorization(ctx, func(ctx *gin.Context) {
		requireGuildElevation(ctx, func(ctx *gin.Context) {
			guildID := tryGetGuildID(ctx)

			freeroles, err := backend.Database.GetFreeRolesGuildSettings(ctx, int64(guildID))
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					freeroles = &database.GuildSettingsFreeroles{
						GuildID:       int64(guildID),
						ToggleEnabled: database.DefaultFreeRoles.ToggleEnabled,
						Roles:         database.DefaultFreeRoles.Roles,
					}
				} else {
					backend.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to get guild freeroles settings")

					ctx.JSON(http.StatusInternalServerError, BaseResponse{
						Ok: false,
					})

					return
				}
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

			freeroles := PartialToGuildSettingsFreeRolesSettings(int64(guildID), partial)

			databaseFreeRolesGuildSettings := database.CreateOrUpdateFreeRolesGuildSettingsParams(*freeroles)

			user := tryGetUser(ctx)
			backend.Logger.Info().Int64("guild_id", int64(guildID)).Interface("obj", *freeroles).Int64("user_id", int64(user.ID)).Msg("Creating or updating guild freerole settings")

			err = utils.RetryWithFallback(
				func() error {
					_, err = backend.Database.CreateOrUpdateFreeRolesGuildSettings(ctx, databaseFreeRolesGuildSettings)
					return err
				},
				func() error {
					return welcomer.EnsureGuild(ctx, backend.Database, discord.Snowflake(guildID))
				},
				nil,
			)
			if err != nil {
				backend.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to create or update guild freeroles settings")

				ctx.JSON(http.StatusInternalServerError, BaseResponse{
					Ok: false,
				})

				return
			}

			getGuildSettingsFreeRoles(ctx)
		})
	})
}

// Validates freerole settings
func doValidateFreeRoles(guildSettings *GuildSettingsFreeRoles) error {
	// TODO: validate freeroles

	return nil
}

func registerGuildSettingsFreeRolesRoutes(g *gin.Engine) {
	g.GET("/api/guild/:guildID/freeroles", getGuildSettingsFreeRoles)
	g.POST("/api/guild/:guildID/freeroles", setGuildSettingsFreeRoles)
}
