package backend

import (
	"net/http"
	"strings"
	"time"

	"github.com/WelcomerTeam/Discord/discord"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// Route GET /login
func login(ctx *gin.Context) {
	session := sessions.Default(ctx)

	queryPath := ctx.Query("path")

	SetPreviousPathSession(session, queryPath)

	doOAuthAuthorize(session, ctx)
}

// Route GET /logout
func logout(ctx *gin.Context) {
	session := sessions.Default(ctx)

	session.Clear()
	_ = session.Save()

	ctx.Redirect(http.StatusTemporaryRedirect, "/")
}

// Route POST /callback
func callback(ctx *gin.Context) {
	queryCode := ctx.Query("code")
	queryState := ctx.Query("state")

	session := sessions.Default(ctx)

	sessionState, ok := GetStateSession(session)
	if !ok || (sessionState != queryState) || queryCode == "" || queryState == "" {
		doOAuthAuthorize(session, ctx)

		return
	}

	token, err := OAuth2Config.Exchange(backend.ctx, queryCode)
	if err != nil {
		backend.Logger.Warn().Err(err).Msg("Failed to exchange code for token")

		doOAuthAuthorize(session, ctx)

		return
	}

	SetTokenSession(session, *token)

	httpInterface := discord.NewBaseInterface()

	discordSession := discord.NewSession(backend.ctx, token.TokenType+" "+token.AccessToken, httpInterface)

	authorizationInformation, err := discord.GetCurrentAuthorizationInformation(discordSession)
	if err != nil || authorizationInformation == nil {
		doOAuthAuthorize(session, ctx)

		return
	}

	sessionUser := SessionUser{
		ID:            authorizationInformation.User.ID,
		Username:      authorizationInformation.User.Username,
		GlobalName:    authorizationInformation.User.GlobalName,
		Discriminator: authorizationInformation.User.Discriminator,
		Avatar:        authorizationInformation.User.Avatar,

		Guilds:                make(map[discord.Snowflake]*SessionGuild),
		GuildsLastRequestedAt: time.Time{},
	}

	SetUserSession(session, sessionUser)

	err = session.Save()
	if err != nil {
		backend.Logger.Warn().Err(err).Msg("Failed to save session")
	}

	queryPath, ok := GetPreviousPathSession(session)
	if !ok || !strings.HasPrefix(queryPath, "/") {
		queryPath = "/"
	}

	ctx.Redirect(http.StatusTemporaryRedirect, queryPath)
}

func registerSessionRoutes(g *gin.Engine) {
	g.GET("/login", login)
	g.GET("/logout", logout)
	g.GET("/callback", callback)
}
