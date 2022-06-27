package backend

import (
	"encoding/gob"
	"net/http"

	"github.com/WelcomerTeam/Discord/discord"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

var OAuth2Config = &oauth2.Config{
	ClientID:     "342685807221407744",
	ClientSecret: "anJWheCbfseLaGPhZ2p-K7JgMCGLA-c4",
	Endpoint: oauth2.Endpoint{
		AuthURL:   discord.EndpointDiscord + discord.EndpointOAuth2Authorize + "?prompt=none",
		TokenURL:  discord.EndpointDiscord + "/api/v10" + discord.EndpointOAuth2Token,
		AuthStyle: oauth2.AuthStyleInParams,
	},
	RedirectURL: "https://beta-d53e2274.welcomer.gg/callback",
	Scopes:      []string{"identify", "guilds"},
}

func init() {
	gob.Register(oauth2.Token{})
	gob.Register(SessionUser{})
}

// Send user to OAuth2 Authorize URL.
func doOAuthAuthorize(session sessions.Session, ctx *gin.Context) {
	state := RandStringBytesRmndr(16)

	SetStateSession(session, state)

	err := session.Save()
	if err != nil {
		backend.Logger.Warn().Err(err).Msg("Failed to save session")
	}

	ctx.Redirect(http.StatusTemporaryRedirect, OAuth2Config.AuthCodeURL(state))
}

// Returns Unauthorized if user is not logged in, else runs handler.
func requireOAuthAuthorization(ctx *gin.Context, handler gin.HandlerFunc) {
	session := sessions.Default(ctx)

	user, ok := GetUserSession(session)
	if !ok {
		ctx.Status(http.StatusUnauthorized)

		return
	}

	ctx.Set(UserKey, user)

	handler(ctx)
}

func registerSessionRoutes(g *gin.Engine) {
	g.GET("/login", func(ctx *gin.Context) {
		session := sessions.Default(ctx)

		doOAuthAuthorize(session, ctx)
	})

	g.GET("/callback", func(ctx *gin.Context) {
		queryCode := ctx.Query("code")
		queryState := ctx.Query("state")

		session := sessions.Default(ctx)

		sessionState, ok := GetStateSession(session)
		if queryCode == "" || queryState == "" || !ok || (sessionState != queryState) {
			doOAuthAuthorize(session, ctx)

			return
		}

		token, err := OAuth2Config.Exchange(backend.ctx, queryCode)
		if err != nil {
			doOAuthAuthorize(session, ctx)

			return
		}

		SetTokenSession(session, *token)

		httpInterface := discord.NewBaseInterface()

		discordSession := discord.NewSession(backend.ctx, token.TokenType+" "+token.AccessToken, httpInterface, backend.Logger)

		authorizationInformation, err := discord.GetCurrentAuthorizationInformation(discordSession)
		if err != nil {
			doOAuthAuthorize(session, ctx)

			return
		}

		user := authorizationInformation.User

		sessionUser := SessionUser{
			ID:            user.ID,
			Username:      user.Username,
			Discriminator: user.Discriminator,
			Avatar:        user.Avatar,
		}

		SetUserSession(session, sessionUser)

		err = session.Save()
		if err != nil {
			backend.Logger.Warn().Err(err).Msg("Failed to save session")
		}

		ctx.Redirect(http.StatusTemporaryRedirect, "/")
	})

	g.GET("/logout", func(ctx *gin.Context) {
		session := sessions.Default(ctx)

		session.Clear()
		_ = session.Save()

		ctx.Redirect(http.StatusTemporaryRedirect, "/")
	})
}
