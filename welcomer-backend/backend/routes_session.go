package backend

import (
	"github.com/WelcomerTeam/Discord/discord"
	utils "github.com/WelcomerTeam/Welcomer/welcomer-utils"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"time"
)

// Send user to OAuth2 Authorize URL.
func doDiscordOAuthAuthorize(session sessions.Session, ctx *gin.Context) {
	state := utils.RandStringBytesRmndr(StateStringLength)

	SetStateSession(session, state)

	if err := session.Save(); err != nil {
		backend.Logger.Warn().Err(err).Msg("Failed to save session")
	}

	ctx.Redirect(http.StatusTemporaryRedirect, DiscordOAuth2Config.AuthCodeURL(state))
}

// Route GET /login
func login(ctx *gin.Context) {
	session := sessions.Default(ctx)

	queryPath := ctx.Query("path")

	SetPreviousPathSession(session, queryPath)

	doDiscordOAuthAuthorize(session, ctx)
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
		doDiscordOAuthAuthorize(session, ctx)

		return
	}

	token, err := DiscordOAuth2Config.Exchange(ctx, queryCode)
	if err != nil {
		backend.Logger.Warn().Err(err).Msg("Failed to exchange code for token")

		doDiscordOAuthAuthorize(session, ctx)

		return
	}

	SetTokenSession(session, *token)

	httpInterface := discord.NewBaseInterface()

	discordSession := discord.NewSession(ctx, token.Type()+" "+token.AccessToken, httpInterface)

	authorizationInformation, err := discord.GetCurrentAuthorizationInformation(discordSession)
	if err != nil || authorizationInformation == nil {
		doDiscordOAuthAuthorize(session, ctx)

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
