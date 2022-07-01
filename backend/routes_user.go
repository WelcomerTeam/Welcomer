package backend

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	discord "github.com/WelcomerTeam/Discord/discord"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

const (
	RefreshFrequency        = time.Minute * 15
	MinimumRefreshFrequency = time.Second * 30
)

func hasWelcomerPresence(guildID discord.Snowflake) (ok bool, err error) {
	guild, err := backend.GRPCInterface.FetchGuildByID(backend.GetBasicEventContext(), guildID)

	if err != nil {
		backend.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to get welcomer presence")

		return false, nil
	}

	if guild == nil {
		return false, nil
	}

	return true, nil
}

func hasWelcomerMembership(guildID discord.Snowflake) (ok bool, err error) {
	var sqlGuildID sql.NullInt64

	sqlGuildID.Int64 = int64(guildID)
	sqlGuildID.Valid = true

	memberships, err := backend.Database.GetUserMembershipsByGuildID(backend.ctx, sqlGuildID)

	if err != nil {
		backend.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to get welcomer memberships")

		return false, nil
	}

	if len(memberships) == 0 {
		return false, nil
	}

	return true, nil
}

// GET /users/@me
// GET /users/guilds?refresh=1
func (b *Backend) GetUserGuilds(ctx *gin.Context, session sessions.Session) (guilds []*SessionGuild, err error) {
	token, ok := GetTokenSession(session)
	if !ok {
		return nil, ErrMissingToken
	}

	httpInterface := discord.NewBaseInterface()

	discordSession := discord.NewSession(backend.ctx, token.TokenType+" "+token.AccessToken, httpInterface, backend.Logger)

	discordGuilds, err := discord.GetCurrentUserGuilds(discordSession)
	if err != nil {
		return nil, fmt.Errorf("failed to get current user guilds: %w", err)
	}

	for _, discordGuild := range discordGuilds {
		welcomerPresence, err := hasWelcomerPresence(discordGuild.ID)
		if err != nil {
			b.Logger.Warn().Err(err).Int("guildID", int(discordGuild.ID)).Msg("Exception getting welcomer presence")
		}

		welcomerMembership, err := hasWelcomerMembership(discordGuild.ID)
		if err != nil {
			b.Logger.Warn().Err(err).Int("guildID", int(discordGuild.ID)).Msg("Exception getting welcomer membership")
		}

		guilds = append(guilds, &SessionGuild{
			ID:            discordGuild.ID,
			Name:          discordGuild.Name,
			Icon:          discordGuild.Icon,
			HasWelcomer:   welcomerPresence,
			HasMembership: welcomerMembership,
		})
	}

	return guilds, nil
}

func registerUserRoutes(g *gin.Engine) {
	g.GET("/api/users/@me", func(ctx *gin.Context) {
		requireOAuthAuthorization(ctx, func(ctx *gin.Context) {
			ctx.Status(http.StatusOK)

			session := sessions.Default(ctx)

			user, _ := GetUserSession(session)

			ctx.JSON(http.StatusOK, BaseResponse{
				Ok:   true,
				Data: user,
			})
		})
	})

	g.GET("/api/users/guilds", func(ctx *gin.Context) {
		requireOAuthAuthorization(ctx, func(ctx *gin.Context) {
			var refreshFrequency time.Duration

			if ctx.Query("refresh") == "1" {
				refreshFrequency = MinimumRefreshFrequency
			} else {
				refreshFrequency = RefreshFrequency
			}

			session := sessions.Default(ctx)

			user, _ := GetUserSession(session)

			refresh := time.Since(user.GuildsLastRequestedAt) > refreshFrequency

			var guilds []*SessionGuild
			var err error

			if refresh {
				guilds, err = backend.GetUserGuilds(ctx, session)
				if err != nil {
					ctx.JSON(http.StatusInternalServerError, BaseResponse{
						Ok:    false,
						Error: err.Error(),
					})

					return
				}

				user.Guilds = guilds
				user.GuildsLastRequestedAt = time.Now()

				SetUserSession(session, user)
			} else {
				guilds = user.Guilds
			}

			err = session.Save()
			if err != nil {
				backend.Logger.Warn().Err(err).Msg("Failed to save session")
			}

			ctx.JSON(http.StatusOK, BaseResponse{
				Ok:   true,
				Data: guilds,
			})
		})
	})
}
