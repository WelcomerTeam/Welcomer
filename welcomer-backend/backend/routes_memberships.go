package backend

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/WelcomerTeam/Discord/discord"
	pb "github.com/WelcomerTeam/Sandwich-Daemon/proto"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	core "github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
)

type getUserMembershipResponse struct {
	Memberships []userMembership `json:"memberships"`
	Accounts    []userAccount    `json:"accounts"`
}

type userAccount struct {
	UserID   string                `json:"user_id"`
	Platform database.PlatformType `json:"platform"`

	Name      string `json:"name"`
	Thumbnail string `json:"thumb_url"`

	Membership *userMembership `json:"membership"`

	CreatedAt time.Time `json:"created_at"`
	EndedAt   time.Time `json:"ended_at"`
	TierID    int64     `json:"tier_id"`
}

type userMembership struct {
	MembershipUUID   uuid.UUID                 `json:"membership_uuid"`
	GuildID          discord.Snowflake         `json:"guild_id"`
	GuildName        string                    `json:"guild_name"`
	GuildIcon        string                    `json:"guild_icon"`
	ExpiresAt        time.Time                 `json:"expires_at"`
	MembershipStatus database.MembershipStatus `json:"membership_status"`
	MembershipType   database.MembershipType   `json:"membership_type"`
	PlatformType     database.PlatformType     `json:"platform_type"`
}

func getAccounts(ctx context.Context, userID discord.Snowflake) []userAccount {
	accounts := make([]userAccount, 0)

	// Patreon

	patreonAccounts, err := welcomer.Queries.GetPatreonUsersByUserID(ctx, int64(userID))
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		welcomer.Logger.Error().Err(err).
			Msg("Failed to get patreon users by user ID")
	}

	for _, patreonAccount := range patreonAccounts {
		accounts = append(accounts, userAccount{
			UserID:     welcomer.Itoa(patreonAccount.PatreonUserID),
			Platform:   database.PlatformTypePatreon,
			Name:       patreonAccount.FullName,
			Thumbnail:  patreonAccount.ThumbUrl,
			Membership: nil,
			CreatedAt:  patreonAccount.PledgeCreatedAt,
			EndedAt:    patreonAccount.PledgeEndedAt,
			TierID:     patreonAccount.TierID,
		})
	}

	return accounts
}

// Route GET /api/memberships.
func getMemberships(ctx *gin.Context) {
	requireOAuthAuthorization(ctx, func(ctx *gin.Context) {
		userMemberships := make([]userMembership, 0)

		userID := tryGetUser(ctx).ID

		memberships, err := welcomer.Queries.GetUserMembershipsByUserID(ctx, int64(userID))
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			ctx.JSON(http.StatusInternalServerError, BaseResponse{
				Ok: false,
			})

			return
		}

		var guildIDs []int64

		for _, membership := range memberships {
			if membership.GuildID != 0 {
				guildIDs = append(guildIDs, membership.GuildID)
			}
		}

		var guilds map[int64]*pb.Guild

		// Fetch all guilds in one request.
		if len(guildIDs) > 0 {
			guildResponse, err := welcomer.SandwichClient.FetchGuild(ctx, &pb.FetchGuildRequest{
				GuildIds: guildIDs,
			})
			if err != nil {
				welcomer.Logger.Error().Err(err).
					Msg("Failed to fetch guilds via GRPC")

				guilds = map[int64]*pb.Guild{}
			} else {
				guilds = guildResponse.GetGuilds()
			}
		} else {
			guilds = map[int64]*pb.Guild{}
		}

		for _, membership := range memberships {
			membershipStatus := database.MembershipStatus(membership.Status)
			if !membershipStatus.IsValid() {
				continue
			}

			switch membershipStatus {
			case database.MembershipStatusUnknown,
				database.MembershipStatusRefunded:
				continue
			}

			membershipType := database.MembershipType(membership.MembershipType)
			if !membershipType.IsValid() {
				continue
			}

			var guildName string

			var guildIcon string

			if membership.GuildID == 0 {
				guildName = ""
				guildIcon = ""
			} else if guild, ok := guilds[membership.GuildID]; ok {
				guildName = guild.GetName()
				guildIcon = guild.GetIcon()
			} else {
				guildName = fmt.Sprintf("Unknown Guild %d", membership.GuildID)
				guildIcon = ""
			}

			userMemberships = append(userMemberships, userMembership{
				MembershipUUID:   membership.MembershipUuid,
				GuildID:          discord.Snowflake(membership.GuildID),
				GuildName:        guildName,
				GuildIcon:        guildIcon,
				ExpiresAt:        membership.ExpiresAt,
				MembershipStatus: membershipStatus,
				MembershipType:   membershipType,
				PlatformType:     database.PlatformType(membership.PlatformType.Int32),
			})
		}

		ctx.JSON(http.StatusOK, BaseResponse{
			Ok: true,
			Data: getUserMembershipResponse{
				Accounts:    getAccounts(ctx, userID),
				Memberships: userMemberships,
			},
		})
	})
}

type postMembershipSubscribeBody struct {
	GuildID *string `json:"guild_id"`
}

// Route POST /api/membership/:membershipID/subscribe.
func postMembershipSubscribe(ctx *gin.Context) {
	requireOAuthAuthorization(ctx, func(ctx *gin.Context) {
		body := postMembershipSubscribeBody{}

		err := ctx.BindJSON(&body)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, BaseResponse{
				Ok:    false,
				Error: err.Error(),
			})

			return
		}

		var guildID discord.Snowflake

		if body.GuildID != nil {
			guildID = discord.Snowflake(welcomer.TryParseInt(*body.GuildID))
		}

		rawMembershipID := ctx.Param(MembershipIDKey)

		membershipID, err := uuid.FromString(rawMembershipID)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, BaseResponse{
				Ok:    false,
				Error: fmt.Sprintf(ErrInvalidParameter.Error(), MembershipIDKey),
			})

			return
		}

		userID := tryGetUser(ctx).ID

		memberships, err := welcomer.Queries.GetUserMembershipsByUserID(ctx, int64(userID))
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			ctx.JSON(http.StatusInternalServerError, BaseResponse{
				Ok: false,
			})

			return
		}

		for _, membership := range memberships {
			if membership.MembershipUuid == membershipID {
				if database.PlatformType(membership.PlatformType.Int32) == database.PlatformTypeDiscord {
					// Discord memberships cannot be transferred or modified.

					ctx.JSON(http.StatusBadRequest, NewBaseResponse(ErrCannotTransferMembership, nil))

					return
				}

				var err error

				var newMembership database.UpdateUserMembershipParams

				if !guildID.IsNil() {
					newMembership, err = core.AddMembershipToServer(ctx, *membership, guildID)
				} else {
					newMembership, err = core.RemoveMembershipFromServer(ctx, *membership)
				}

				if err != nil {
					ctx.JSON(http.StatusInternalServerError, BaseResponse{
						Ok: false,
					})

					return
				}

				ctx.JSON(http.StatusOK, BaseResponse{
					Ok:   true,
					Data: newMembership,
				})

				return
			}
		}
	})
}

func registerMembershipsRoutes(r *gin.Engine) {
	r.GET("/api/memberships", getMemberships)
	r.POST("/api/memberships/:membershipID/subscribe", postMembershipSubscribe)
}
