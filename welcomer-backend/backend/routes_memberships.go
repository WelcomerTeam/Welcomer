package backend

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/WelcomerTeam/Discord/discord"
	pb "github.com/WelcomerTeam/Sandwich-Daemon/protobuf"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	utils "github.com/WelcomerTeam/Welcomer/welcomer-utils"
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
}

func getAccounts(ctx context.Context, userID discord.Snowflake) []userAccount {
	accounts := make([]userAccount, 0)

	// Patreon

	patreonAccounts, err := backend.Database.GetPatreonUsersByUserID(ctx, int64(userID))
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		backend.Logger.Error().Err(err).
			Msg("Failed to get patreon users by user ID")
	}

	for _, patreonAccount := range patreonAccounts {
		accounts = append(accounts, userAccount{
			UserID:     utils.Itoa(patreonAccount.PatreonUserID),
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

// Route GET /api/memberships
func getMemberships(ctx *gin.Context) {
	requireOAuthAuthorization(ctx, func(ctx *gin.Context) {
		userMemberships := make([]userMembership, 0)

		userID := tryGetUser(ctx).ID

		memberships, err := backend.Database.GetUserMembershipsByUserID(ctx, int64(userID))
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
			guildResponse, err := backend.SandwichClient.FetchGuild(ctx, &pb.FetchGuildRequest{
				GuildIDs: guildIDs,
			})
			if err != nil {
				backend.Logger.Error().Err(err).
					Msg("Failed to fetch guilds via GRPC")

				guilds = map[int64]*pb.Guild{}
			} else {
				guilds = guildResponse.Guilds
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
				guildName = guild.Name
				guildIcon = guild.Icon
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

// Route POST /api/membership/:membershipID/subscribe
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
			guildID = discord.Snowflake(utils.TryParseInt(*body.GuildID))
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

		memberships, err := backend.Database.GetUserMembershipsByUserID(ctx, int64(userID))
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			ctx.JSON(http.StatusInternalServerError, BaseResponse{
				Ok: false,
			})

			return
		}

		for _, membership := range memberships {
			if membership.MembershipUuid == membershipID {
				var err error
				var newMembership database.UpdateUserMembershipParams

				if !guildID.IsNil() {
					newMembership, err = addMembershipToServer(ctx, *membership, guildID)
				} else {
					newMembership, err = removeMembershipFromServer(ctx, *membership)
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

func addMembershipToServer(ctx context.Context, membership database.GetUserMembershipsByUserIDRow, guildID discord.Snowflake) (newMembership database.UpdateUserMembershipParams, err error) {
	// if membership.GuildID != 0 {
	// 	backend.Logger.Error().
	// 		Str("membership_uuid", membership.MembershipUuid.String()).
	// 		Int64("guild_id", membership.GuildID).
	// 		Int64("new_guild_id", int64(guildID)).
	// 		Msg("Membership is already in use")

	// 	return database.UpdateUserMembershipParams{}, ErrMembershipAlreadyInUse
	// }

	switch database.MembershipStatus(membership.Status) {
	case database.MembershipStatusUnknown,
		database.MembershipStatusRefunded:
		return database.UpdateUserMembershipParams{}, ErrMembershipInvalid
	case database.MembershipStatusExpired:
		return database.UpdateUserMembershipParams{}, ErrMembershipExpired
	case database.MembershipStatusActive,
		database.MembershipStatusIdle:

		// Check if the transaction is completed.
		if database.TransactionStatus(membership.TransactionStatus.Int32) != database.TransactionStatusCompleted {
			backend.Logger.Error().
				Str("membership_uuid", membership.MembershipUuid.String()).
				Int32("transaction_status", membership.TransactionStatus.Int32).
				Msg("Membership transaction is not completed")

			return database.UpdateUserMembershipParams{}, utils.ErrTransactionNotComplete
		}

		membership.Status = int32(database.MembershipStatusActive)
		membership.GuildID = int64(guildID)

		if membership.StartedAt.IsZero() {
			// If StartedAt is zero, the membership has not started yet.
			// Compare ExpiresAt against StartedAt to find the length of the membership.
			duration := membership.ExpiresAt.Sub(membership.StartedAt)

			membership.StartedAt = time.Now()
			membership.ExpiresAt = membership.StartedAt.Add(duration)
		} else {
			// If StartedAt is not zero, the membership has already started.
			// Do not modify the ExpiresAt, but reset the StartedAt.
			membership.StartedAt = time.Now()
		}

		newMembership := database.UpdateUserMembershipParams{
			MembershipUuid:  membership.MembershipUuid,
			StartedAt:       membership.StartedAt,
			ExpiresAt:       membership.ExpiresAt,
			Status:          membership.Status,
			TransactionUuid: membership.TransactionUuid,
			UserID:          membership.UserID,
			GuildID:         membership.GuildID,
		}

		_, err = backend.Database.UpdateUserMembership(ctx, newMembership)
		if err != nil {
			backend.Logger.Error().Err(err).
				Str("membership_uuid", membership.MembershipUuid.String()).
				Msg("Failed to update user membership")

			return database.UpdateUserMembershipParams{}, err
		}

		return newMembership, nil
	default:
		backend.Logger.Error().
			Str("membership_uuid", membership.MembershipUuid.String()).
			Int32("status", membership.Status).
			Msg("Unhandled membership status")

		return database.UpdateUserMembershipParams{}, ErrUnhandledMembership
	}
}

func removeMembershipFromServer(ctx context.Context, membership database.GetUserMembershipsByUserIDRow) (newMembership database.UpdateUserMembershipParams, err error) {
	if membership.GuildID == 0 {
		return database.UpdateUserMembershipParams{}, ErrMembershipNotInUse
	}

	// Only set the status to Idle if the membership is currently Active.
	membership.Status = utils.If(membership.Status == int32(database.MembershipStatusActive), int32(database.MembershipStatusIdle), membership.Status)
	membership.GuildID = 0

	newMembership = database.UpdateUserMembershipParams{
		MembershipUuid:  membership.MembershipUuid,
		StartedAt:       membership.StartedAt,
		ExpiresAt:       membership.ExpiresAt,
		Status:          membership.Status,
		TransactionUuid: membership.TransactionUuid,
		UserID:          membership.UserID,
		GuildID:         membership.GuildID,
	}

	_, err = backend.Database.UpdateUserMembership(ctx, newMembership)
	if err != nil {
		backend.Logger.Error().Err(err).
			Str("membership_uuid", membership.MembershipUuid.String()).
			Msg("Failed to update user membership")

		return database.UpdateUserMembershipParams{}, err
	}

	return newMembership, nil
}

func registerMembershipsRoutes(r *gin.Engine) {
	r.GET("/api/memberships", getMemberships)
	r.POST("/api/memberships/:membershipID/subscribe", postMembershipSubscribe)
}
