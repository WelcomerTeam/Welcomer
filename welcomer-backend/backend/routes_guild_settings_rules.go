package backend

import (
	_ "embed"
	"fmt"
	"net/http"

	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	"github.com/gin-gonic/gin"
)

const (
	MaxRuleCount  = 25
	MaxRuleLength = 250
)

// Route GET /api/guild/:guildID/rules
func getGuildSettingsRules(ctx *gin.Context) {
	requireOAuthAuthorization(ctx, func(ctx *gin.Context) {
		requireGuildElevation(ctx, func(ctx *gin.Context) {
			guildID := tryGetGuildID(ctx)

			rules, err := backend.Database.GetRulesGuildSettings(ctx, int64(guildID))
			if err != nil {
				backend.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to get guild rules settings")
			}

			partial := GuildSettingsRulesSettingsToPartial(rules)

			ctx.JSON(http.StatusOK, BaseResponse{
				Ok:   true,
				Data: partial,
			})
		})
	})
}

// Route POST /api/guild/:guildID/rules
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

			err = ensureGuild(ctx, guildID)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, BaseResponse{
					Ok: false,
				})

				return
			}

			rules := PartialToGuildSettingsRulesSettings(int64(guildID), partial)

			databaseRulesGuildSettings := database.CreateOrUpdateRulesGuildSettingsParams(*rules)
			_, err = backend.Database.CreateOrUpdateRulesGuildSettings(ctx, &databaseRulesGuildSettings)
			if err != nil {
				backend.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to create or update guild rules settings")
			}

			getGuildSettingsRules(ctx)
		})
	})
}

// Validates rule settings
func doValidateRules(guildSettings *GuildSettingsRules) error {
	if len(guildSettings.Rules) > MaxRuleCount {
		return fmt.Errorf("too many rules (%d): %w", len(guildSettings.Rules), ErrListTooLong)
	}

	for i, r := range guildSettings.Rules {
		if len(r) > MaxRuleLength {
			return fmt.Errorf("rule %d has a length too long: %w", i, ErrStringTooLong)
		}
	}

	return nil
}

func registerGuildSettingsRulesRoutes(g *gin.Engine) {
	g.GET("/api/guild/:guildID/rules", getGuildSettingsRules)
	g.POST("/api/guild/:guildID/rules", setGuildSettingsRules)
}
