package backend

import (
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
	return true, nil
}

func hasWelcomerMembership(guildID discord.Snowflake) (ok bool, err error) {
	return true, nil
}

// GET /user/@me
// GET /user/guilds?refresh=1
func (b *Backend) GetUserGuilds(ctx *gin.Context, session sessions.Session) (guilds []*SessionGuild, err error) {
	token, ok := GetTokenSession(session)
	if !ok {
		return nil, ErrMissingToken
	}

	httpInterface := discord.NewBaseInterface()
	httpInterface.SetDebug(true)

	discordSession := discord.NewSession(backend.ctx, token.TokenType+" "+token.AccessToken, httpInterface, backend.Logger)

	discordGuilds, err := discord.GetCurrentUserGuilds(discordSession)
	if err != nil {
		return nil, fmt.Errorf("failed to get current user guilds: %w", err)
	}

	for _, discordGuild := range discordGuilds {
		guilds = append(guilds, &SessionGuild{
			ID:            discordGuild.ID,
			Name:          discordGuild.Name,
			Icon:          discordGuild.Icon,
			HasWelcomer:   false,
			HasMembership: false,
		})
	}

	return
}

func registerUserRoutes(g *gin.Engine) {
	g.GET("/api/user/@me", func(ctx *gin.Context) {
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

	g.GET("/api/user/guilds", func(ctx *gin.Context) {
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
			println(time.Since(user.GuildsLastRequestedAt), refreshFrequency, refresh)

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
