package backend

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	core "github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	utils "github.com/WelcomerTeam/Welcomer/welcomer-utils"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4"
)

// Send user to OAuth2 Authorize URL.
func doPatreonOAuthAuthorize(session sessions.Session, ctx *gin.Context) {
	state := utils.RandStringBytesRmndr(StateStringLength)

	SetStateSession(session, state)

	if err := session.Save(); err != nil {
		backend.Logger.Warn().Err(err).Msg("Failed to save session")
	}

	ctx.Redirect(http.StatusTemporaryRedirect, PatreonOAuth2Config.AuthCodeURL(state))
}

// Route GET /patreon_link
func getPatreonLink(ctx *gin.Context) {
	err := checkOAuthAuthorization(ctx)
	if err != nil {
		ctx.Redirect(http.StatusTemporaryRedirect, "/login?path="+ctx.Request.URL.EscapedPath())

		return
	}

	session := sessions.Default(ctx)

	queryPath := ctx.Query("path")

	SetPreviousPathSession(session, queryPath)

	doPatreonOAuthAuthorize(session, ctx)
}

// Route POST /patreon_callback
func getPatreonCallback(ctx *gin.Context) {
	requireOAuthAuthorization(ctx, func(ctx *gin.Context) {
		queryCode := ctx.Query("code")
		queryState := ctx.Query("state")

		session := sessions.Default(ctx)

		sessionState, ok := GetStateSession(session)
		if !ok || (sessionState != queryState) || queryCode == "" || queryState == "" {
			doPatreonOAuthAuthorize(session, ctx)

			return
		}

		token, err := PatreonOAuth2Config.Exchange(backend.ctx, queryCode)
		if err != nil {
			backend.Logger.Warn().Err(err).Msg("Failed to exchange code for token")

			doPatreonOAuthAuthorize(session, ctx)

			return
		}

		patreonUser, err := core.IdentifyPatreonMember(backend.ctx, token.Type()+" "+token.AccessToken)
		if err != nil {
			backend.Logger.Warn().Err(err).Msg("Failed to identify member")

			doPatreonOAuthAuthorize(session, ctx)

			return
		}

		patreonUsers, err := core.GetAllPatreonMembers(ctx, token.Type()+" "+token.AccessToken, nil, "")
		if err != nil {
			backend.Logger.Warn().Err(err).Msg("Failed to get all patreon members")

			doPatreonOAuthAuthorize(session, ctx)

			return
		}

		var patreonMember *core.PatreonMember

		for _, puser := range patreonUsers {
			if puser.PatreonUserID == patreonUser.ID {
				patreonMember = &puser
			}
		}

		user := tryGetUser(ctx)

		databasePatreonUser, err := backend.Database.GetPatreonUser(ctx, int64(patreonUser.ID))
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			backend.Logger.Warn().Err(err).Msg("Failed to get patreon user")

			doPatreonOAuthAuthorize(session, ctx)

			return
		}

		var pledgeCreatedAt time.Time
		var pledgeEndedAt time.Time
		var tierID int64

		pledgeCreatedAt = databasePatreonUser.PledgeCreatedAt
		pledgeEndedAt = databasePatreonUser.PledgeEndedAt

		if patreonMember != nil && len(patreonMember.EntitledTiers) > 0 {
			if pledgeCreatedAt.IsZero() {
				pledgeCreatedAt = time.Now()
			}

			pledgeEndedAt = time.Time{}
			tierID = int64(patreonMember.EntitledTiers[0])
		} else {
			pledgeCreatedAt = time.Time{}

			if pledgeEndedAt.IsZero() && !databasePatreonUser.CreatedAt.IsZero() {
				pledgeEndedAt = time.Now()
			}
		}

		_, err = backend.Database.CreateOrUpdatePatreonUser(ctx, database.CreateOrUpdatePatreonUserParams{
			PatreonUserID:   int64(patreonUser.ID),
			UserID:          int64(user.ID),
			FullName:        patreonUser.FullName,
			Email:           patreonUser.Email,
			ThumbUrl:        patreonUser.ThumbURL,
			PledgeCreatedAt: pledgeCreatedAt,
			PledgeEndedAt:   pledgeEndedAt,
			TierID:          tierID,
		})
		if err != nil {
			backend.Logger.Warn().Err(err).Msg("Failed to create or update patreon user")

			doPatreonOAuthAuthorize(session, ctx)

			return
		}

		if tierID != databasePatreonUser.TierID {
			println(user.ID, "tier changed from", databasePatreonUser.TierID, "to", tierID)
			// TODO: Handle tier change
		}

		queryPath, ok := GetPreviousPathSession(session)
		if !ok || !strings.HasPrefix(queryPath, "/") {
			queryPath = "/"
		}

		ctx.Redirect(http.StatusTemporaryRedirect, queryPath)
	})
}

// Route DELETE /api/patreon/link/:patreonID
func deletePatreonLink(ctx *gin.Context) {
	requireOAuthAuthorization(ctx, func(ctx *gin.Context) {
		rawPatreonID := ctx.Param(PatreonIDKey)

		patreonIDInt, err := utils.Atoi(rawPatreonID)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, BaseResponse{
				Ok:    false,
				Error: fmt.Sprintf(ErrMissingParameter.Error(), patreonIDInt),
				Data:  nil,
			})

			return
		}

		_, err = backend.Database.DeletePatreonUser(ctx, database.DeletePatreonUserParams{
			PatreonUserID: int64(patreonIDInt),
			UserID:        int64(tryGetUser(ctx).ID),
		})
		if err != nil {
			backend.Logger.Warn().Err(err).Msg("Failed to delete patreon user")

			ctx.JSON(http.StatusInternalServerError, BaseResponse{
				Ok: false,
			})

			return
		}

		ctx.JSON(http.StatusOK, BaseResponse{
			Ok: true,
		})
	})
}

func registerPatreonRoutes(g *gin.Engine) {
	g.GET("/patreon_link", getPatreonLink)
	g.GET("/patreon_callback", getPatreonCallback)

	g.DELETE("/api/patreon/link/:patreonID", deletePatreonLink)
}
