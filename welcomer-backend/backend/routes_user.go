package backend

import (
	"errors"
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

func (b *Backend) GetUserGuilds(session sessions.Session) (guilds map[discord.Snowflake]*SessionGuild, err error) {
	token, ok := GetTokenSession(session)
	if !ok {
		return nil, ErrMissingToken
	}

	guilds = make(map[discord.Snowflake]*SessionGuild)

	user, ok := GetUserSession(session)
	if !ok {
		return nil, ErrMissingUser
	}

	httpInterface := discord.NewBaseInterface()

	discordSession := discord.NewSession(backend.ctx, token.TokenType+" "+token.AccessToken, httpInterface, backend.Logger)

	discordGuilds, err := discord.GetCurrentUserGuilds(discordSession)
	if err != nil {
		// If we have received Unauthorized, clear their token.
		if errors.Is(err, discord.ErrUnauthorized) {
			ClearTokenSession(session)

			err = session.Save()
			if err != nil {
				backend.Logger.Warn().Err(err).Msg("Failed to save session")
			}

			return nil, discord.ErrUnauthorized
		}

		return nil, fmt.Errorf("failed to get current user guilds: %w", err)
	}

	for _, discordGuild := range discordGuilds {
		welcomerPresence, _, _, err := hasWelcomerPresence(discordGuild.ID, false)
		if err != nil {
			b.Logger.Warn().Err(err).Int("guildID", int(discordGuild.ID)).Msg("Exception getting welcomer presence")
		}

		hasWelcomerPro, hasCustomBackgrounds, err := getGuildMembership(discordGuild.ID)
		if err != nil {
			b.Logger.Warn().Err(err).Int("guildID", int(discordGuild.ID)).Msg("Exception getting welcomer membership")
		}

		guilds[discordGuild.ID] = &SessionGuild{
			ID:          discordGuild.ID,
			Name:        discordGuild.Name,
			Icon:        discordGuild.Icon,
			HasWelcomer: welcomerPresence,

			HasWelcomerPro:       hasWelcomerPro,
			HasCustomBackgrounds: hasCustomBackgrounds,
			HasElevation:         hasElevation(discordGuild, user),
			IsOwner:              discordGuild.Owner,
		}
	}

	return guilds, nil
}

func (b *Backend) GetUserMemberships(session sessions.Session) (memberships []*Membership, err error) {
	user, ok := GetUserSession(session)
	if !ok {
		return nil, ErrMissingUser
	}

	userMemberships, err := backend.Database.GetUserMembershipsByUserID(backend.ctx, int64(user.ID))
	if err != nil {
		return nil, fmt.Errorf("failed to get user memberships: %w", err)
	}

	for _, userMembership := range userMemberships {
		membership := &Membership{
			MembershipUuid: userMembership.MembershipUuid,
			CreatedAt:      userMembership.CreatedAt,
			UpdatedAt:      userMembership.UpdatedAt,
			StartedAt:      userMembership.StartedAt,
			ExpiresAt:      userMembership.ExpiresAt,
			Status:         userMembership.Status,
			MembershipType: userMembership.MembershipType,
			GuildID:        discord.Snowflake(userMembership.GuildID),
		}

		if userMembership.GuildID != 0 {
			ok, guild, _, err := hasWelcomerPresence(discord.Snowflake(userMembership.GuildID), false)
			if err != nil {
				backend.Logger.Warn().Err(err).Int64("guildID", userMembership.GuildID).Msg("Exception getting guild info")
			}

			if ok {
				membership.Guild = GuildToMinimal(guild)
			}
		}

		memberships = append(memberships, membership)
	}

	return memberships, nil
}

// Route GET /api/users/@me
func usersMe(ctx *gin.Context) {
	requireOAuthAuthorization(ctx, func(ctx *gin.Context) {
		session := sessions.Default(ctx)

		user, _ := GetUserSession(session)

		refresh := time.Since(user.MembershipsLastRequestedAt) > RefreshFrequency

		if refresh {
			memberships, err := backend.GetUserMemberships(session)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, BaseResponse{
					Ok:    false,
					Error: err.Error(),
				})

				return
			}

			user.Memberships = memberships
			user.MembershipsLastRequestedAt = time.Now()

			SetUserSession(session, user)
		}

		ctx.JSON(http.StatusOK, BaseResponse{
			Ok:   true,
			Data: SessionUserToMinimal(&user),
		})
	})
}

// Route GET /api/users/@me/memberships
func usersMeMemberships(ctx *gin.Context) {
	requireOAuthAuthorization(ctx, func(ctx *gin.Context) {
		var refreshFrequency time.Duration

		if ctx.Query("refresh") == "1" {
			refreshFrequency = MinimumRefreshFrequency
		} else {
			refreshFrequency = RefreshFrequency
		}

		session := sessions.Default(ctx)

		user, _ := GetUserSession(session)

		refresh := time.Since(user.MembershipsLastRequestedAt) > refreshFrequency

		var memberships []*Membership
		var err error

		if refresh {
			memberships, err = backend.GetUserMemberships(session)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, BaseResponse{
					Ok:    false,
					Error: err.Error(),
				})

				return
			}

			user.Memberships = memberships
			user.MembershipsLastRequestedAt = time.Now()

			SetUserSession(session, user)
		} else {
			memberships = user.Memberships
		}

		err = session.Save()
		if err != nil {
			backend.Logger.Warn().Err(err).Msg("Failed to save session")
		}

		ctx.JSON(http.StatusOK, BaseResponse{
			Ok:   true,
			Data: memberships,
		})
	})
}

// ROUTE GET /api/users/guilds
func usersGuilds(ctx *gin.Context) {
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

		var mappedGuilds map[discord.Snowflake]*SessionGuild
		var err error

		if refresh {
			mappedGuilds, err = backend.GetUserGuilds(session)

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

			user.Guilds = mappedGuilds
			user.GuildsLastRequestedAt = time.Now()

			SetUserSession(session, user)

			err = session.Save()
			if err != nil {
				backend.Logger.Warn().Err(err).Msg("Failed to save session")
			}
		} else {
			mappedGuilds = user.Guilds
		}

		guilds := make([]*SessionGuild, 0, len(mappedGuilds))

		for _, guild := range mappedGuilds {
			guilds = append(guilds, guild)
		}

		ctx.JSON(http.StatusOK, BaseResponse{
			Ok:   true,
			Data: guilds,
		})
	})
}

func registerUserRoutes(g *gin.Engine) {
	g.GET("/api/users/@me", usersMe)
	g.GET("/api/users/@me/memberships", usersMeMemberships)
	g.GET("/api/users/guilds", usersGuilds)
}
