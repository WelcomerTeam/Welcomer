package backend

import (
	"net/http"
	"strings"
	"time"

	"github.com/WelcomerTeam/Discord/discord"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// GET /login
func login(ctx *gin.Context) {
	session := sessions.Default(ctx)

	queryPath := ctx.Query("path")

	SetPreviousPathSession(session, queryPath)

	doOAuthAuthorize(session, ctx)
}

// GET /logout
func logout(ctx *gin.Context) {
	session := sessions.Default(ctx)

	session.Clear()
	_ = session.Save()

	ctx.Redirect(http.StatusTemporaryRedirect, "/")
}

// POST /callback
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

		Guilds:                make([]*SessionGuild, 0),
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
