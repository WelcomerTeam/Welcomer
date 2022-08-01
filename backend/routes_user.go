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
	LazyRefreshFrequency    = time.Hour
)

func (b *Backend) GetUserGuilds(session sessions.Session) (guilds []*SessionGuild, err error) {
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
		welcomerPresence, _, err := hasWelcomerPresence(discordGuild.ID)
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

// GET /users/@me
func usersMe(ctx *gin.Context) {
	requireOAuthAuthorization(ctx, func(ctx *gin.Context) {
		session := sessions.Default(ctx)

		user, _ := GetUserSession(session)

		ctx.JSON(http.StatusOK, BaseResponse{
			Ok:   true,
			Data: user,
		})
	})
}

func usersGuild(ctx *gin.Context) {
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
			guilds, err = backend.GetUserGuilds(session)
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
}

func registerUserRoutes(g *gin.Engine) {
	g.GET("/api/users/@me", usersMe)
	g.GET("/api/users/guilds", usersGuild)
}
