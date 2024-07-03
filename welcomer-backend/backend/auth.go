package backend

import (
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"net/http"
	"time"

	discord "github.com/WelcomerTeam/Discord/discord"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	utils "github.com/WelcomerTeam/Welcomer/welcomer-utils"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

const (
	StateStringLength = 16
)

var OAuth2Config = &oauth2.Config{
	ClientID:     "",
	ClientSecret: "",
	Endpoint: oauth2.Endpoint{
		AuthURL:   discord.EndpointDiscord + discord.EndpointOAuth2Authorize + "?prompt=none",
		TokenURL:  discord.EndpointDiscord + "/api/v10" + discord.EndpointOAuth2Token,
		AuthStyle: oauth2.AuthStyleInParams,
	},
	RedirectURL: "",
	Scopes:      []string{"identify", "guilds"},
}

func init() {
	gob.Register(oauth2.Token{})
	gob.Register(SessionUser{})
}

func hasElevation(discordGuild discord.Guild, user SessionUser) bool {
	return welcomer.MemberHasElevation(discordGuild, discord.GuildMember{
		User: &discord.User{
			ID:            user.ID,
			Username:      user.Username,
			Discriminator: user.Discriminator,
			GlobalName:    user.GlobalName,
			Avatar:        user.Avatar,
		},
		GuildID:     &discordGuild.ID,
		Permissions: discordGuild.Permissions,
	})
}

func userHasElevation(guildID discord.Snowflake, user SessionUser) bool {
	guild, ok := user.Guilds[guildID]
	if !ok {
		return false
	}

	return guild.IsOwner || guild.HasElevation
}

func checkToken(ctx context.Context, config *oauth2.Config, token oauth2.Token) (newToken *oauth2.Token, changed bool, err error) {
	source := config.TokenSource(ctx, &token)

	newToken, err = source.Token()
	if err != nil {
		backend.Logger.Warn().Err(err).Msg("Failed to check token")

		return
	}

	changed = newToken.AccessToken != token.AccessToken

	return
}

// Send user to OAuth2 Authorize URL.
func doOAuthAuthorize(session sessions.Session, ctx *gin.Context) {
	state := utils.RandStringBytesRmndr(StateStringLength)

	SetStateSession(session, state)

	if err := session.Save(); err != nil {
		backend.Logger.Warn().Err(err).Msg("Failed to save session")
	}

	ctx.Redirect(http.StatusTemporaryRedirect, OAuth2Config.AuthCodeURL(state))
}

// Returns Unauthorized if user is not logged in, else runs handler.
func requireOAuthAuthorization(ctx *gin.Context, handler gin.HandlerFunc) {
	session := sessions.Default(ctx)

	user, ok := GetUserSession(session)
	if !ok {
		backend.Logger.Warn().Msg("Failed to get user session")

		ctx.JSON(http.StatusUnauthorized, BaseResponse{
			Ok:    false,
			Error: ErrMissingUser.Error(),
		})

		return
	}

	ctx.Set(UserKey, user)

	token, ok := GetTokenSession(session)
	if !ok {
		backend.Logger.Warn().Msg("Failed to get token session")

		ctx.JSON(http.StatusUnauthorized, BaseResponse{
			Ok:    false,
			Error: ErrMissingToken.Error(),
		})

		return
	}

	newToken, changed, err := checkToken(backend.ctx, OAuth2Config, token)
	if err != nil {
		ClearTokenSession(session)

		err = session.Save()
		if err != nil {
			backend.Logger.Warn().Err(err).Msg("Failed to save session")
		}

		ctx.JSON(http.StatusUnauthorized, BaseResponse{
			Ok:    false,
			Error: ErrOAuthFailure.Error(),
		})

		return
	}

	if changed {
		SetTokenSession(session, *newToken)

		err = session.Save()
		if err != nil {
			backend.Logger.Warn().Err(err).Msg("Failed to save session")
		}
	}

	handler(ctx)
}

// RequireGuildIDKey returns BadRequest if the request does not supply a guildID. Sets GuildID key.
func requireGuildIDKey(ctx *gin.Context, handler gin.HandlerFunc) {
	rawGuildID := ctx.Param(GuildIDKey)

	guildIDInt, err := utils.Atoi(rawGuildID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, BaseResponse{
			Ok:    false,
			Error: fmt.Sprintf(ErrMissingParameter.Error(), GuildIDKey),
			Data:  nil,
		})

		return
	}

	guildID := discord.Snowflake(guildIDInt)
	ctx.Set(GuildIDKey, guildID)

	handler(ctx)
}

// RequireMutualGuild returns Unauthorized if the user is not in the guild. Sets GuildID key.
func requireMutualGuild(ctx *gin.Context, handler gin.HandlerFunc) {
	requireGuildIDKey(ctx, func(ctx *gin.Context) {
		guildID := tryGetGuildID(ctx)

		session := sessions.Default(ctx)

		user, ok := GetUserSession(session)
		if !ok {
			backend.Logger.Warn().Msg("Failed to get user session")

			ctx.JSON(http.StatusUnauthorized, BaseResponse{
				Ok:    false,
				Error: ErrMissingUser.Error(),
			})

			return
		}

		for _, guild := range user.Guilds {
			if guild.ID == guildID {
				handler(ctx)

				return
			}
		}

		refresh := time.Since(user.GuildsLastRequestedAt) > LazyRefreshFrequency

		// Try get up-to-date user guilds.

		if !refresh {
			ctx.JSON(http.StatusForbidden, BaseResponse{
				Ok:    false,
				Error: ErrWelcomerMissing.Error(),
				Data:  nil,
			})
		}

		guilds, err := backend.GetUserGuilds(session)
		if err != nil {
			var statusCode int
			if errors.Is(err, discord.ErrUnauthorized) {
				statusCode = http.StatusUnauthorized
			} else {
				statusCode = http.StatusInternalServerError
			}

			ctx.JSON(statusCode, BaseResponse{
				Ok:    false,
				Error: err.Error(),
			})

			return
		}

		user.Guilds = guilds
		user.GuildsLastRequestedAt = time.Now()

		SetUserSession(session, user)

		err = session.Save()
		if err != nil {
			backend.Logger.Warn().Err(err).Msg("Failed to save session")
		}

		for _, guild := range user.Guilds {
			if guild.ID == guildID {
				ctx.Set(GuildKey, guild)

				handler(ctx)

				return
			}
		}

		ctx.JSON(http.StatusForbidden, BaseResponse{
			Ok:    false,
			Error: ErrWelcomerMissing.Error(),
			Data:  nil,
		})
	})
}

// RequireGuildElevation checks if a user has privileges on a guild.
func requireGuildElevation(ctx *gin.Context, handler gin.HandlerFunc) {
	requireMutualGuild(ctx, func(ctx *gin.Context) {
		session := sessions.Default(ctx)

		user, ok := GetUserSession(session)
		if !ok {
			backend.Logger.Warn().Msg("Failed to get user session")

			ctx.JSON(http.StatusUnauthorized, BaseResponse{
				Ok:    false,
				Error: ErrMissingUser.Error(),
			})

			return
		}

		guildID := tryGetGuildID(ctx)

		if !userHasElevation(guildID, user) {
			backend.Logger.Warn().
				Int64("user_id", int64(user.ID)).
				Int64("guild_id", int64(guildID)).
				Msg("User does not have elevation")

			ctx.JSON(http.StatusUnauthorized, BaseResponse{
				Ok:    false,
				Error: ErrMissingUser.Error(),
			})

			return
		}

		handler(ctx)
	})
}

// tryGetGuildID returns GuildID from context.
func tryGetGuildID(ctx *gin.Context) discord.Snowflake {
	rawGuildID, _ := ctx.Get(GuildIDKey)
	guildID, _ := rawGuildID.(discord.Snowflake)

	return guildID
}

// tryGetUser returns User from context.
func tryGetUser(ctx *gin.Context) SessionUser {
	rawUser, _ := ctx.Get(UserKey)
	user, _ := rawUser.(SessionUser)

	return user
}
