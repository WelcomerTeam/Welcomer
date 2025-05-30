package backend

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4"
)

// Route GET /api/guild/:guildID/rules.
func getGuildSettingsRules(ctx *gin.Context) {
	requireOAuthAuthorization(ctx, func(ctx *gin.Context) {
		requireGuildElevation(ctx, func(ctx *gin.Context) {
			guildID := tryGetGuildID(ctx)

			rules, err := welcomer.Queries.GetRulesGuildSettings(ctx, int64(guildID))
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					rules = &database.GuildSettingsRules{
						GuildID:          int64(guildID),
						ToggleEnabled:    welcomer.DefaultRules.ToggleEnabled,
						ToggleDmsEnabled: welcomer.DefaultRules.ToggleDmsEnabled,
						Rules:            welcomer.DefaultRules.Rules,
					}
				} else {
					welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to get guild rules settings")

					ctx.JSON(http.StatusInternalServerError, BaseResponse{
						Ok: false,
					})

					return
				}
			}

			partial := GuildSettingsRulesSettingsToPartial(rules)

			ctx.JSON(http.StatusOK, BaseResponse{
				Ok:   true,
				Data: partial,
			})
		})
	})
}

// Route POST /api/guild/:guildID/rules.
func setGuildSettingsRules(ctx *gin.Context) {
	requireOAuthAuthorization(ctx, func(ctx *gin.Context) {
		requireGuildElevation(ctx, func(ctx *gin.Context) {
			partial := &GuildSettingsRules{}

			var err error

			err = ctx.BindJSON(partial)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, BaseResponse{
					Ok:    false,
					Error: err.Error(),
				})

				return
			}

			err = doValidateRules(partial)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, BaseResponse{
					Ok:    false,
					Error: err.Error(),
				})

				return
			}

			guildID := tryGetGuildID(ctx)

			rules := PartialToGuildSettingsRulesSettings(int64(guildID), partial)

			databaseRulesGuildSettings := database.CreateOrUpdateRulesGuildSettingsParams(*rules)

			user := tryGetUser(ctx)
			welcomer.Logger.Info().Int64("guild_id", int64(guildID)).Interface("obj", *rules).Int64("user_id", int64(user.ID)).Msg("Creating or updating guild rule settings")

			err = welcomer.RetryWithFallback(
				func() error {
					_, err = welcomer.Queries.CreateOrUpdateRulesGuildSettings(ctx, databaseRulesGuildSettings)
					return err
				},
				func() error {
					return welcomer.EnsureGuild(ctx, guildID)
				},
				nil,
			)
			if err != nil {
				welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to create or update guild rules settings")

				ctx.JSON(http.StatusInternalServerError, BaseResponse{
					Ok: false,
				})

				return
			}

			getGuildSettingsRules(ctx)
		})
	})
}

// Validates rule settings.
func doValidateRules(guildSettings *GuildSettingsRules) error {
	if len(guildSettings.Rules) > welcomer.MaxRuleCount {
		return fmt.Errorf("too many rules (%d): %w", len(guildSettings.Rules), ErrListTooLong)
	}

	for i, r := range guildSettings.Rules {
		if len(r) > welcomer.MaxRuleLength {
			return fmt.Errorf("rule %d has a length too long: %w", i, ErrStringTooLong)
		}
	}

	return nil
}

func registerGuildSettingsRulesRoutes(g *gin.Engine) {
	g.GET("/api/guild/:guildID/rules", getGuildSettingsRules)
	g.POST("/api/guild/:guildID/rules", setGuildSettingsRules)
}
