package welcomer

import (
	"context"
	"errors"
	"time"

	"github.com/WelcomerTeam/Discord/discord"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	"github.com/gofrs/uuid"
)

var (
	ErrMembershipAlreadyInUse = errors.New("membership is already in use")
	ErrMembershipInvalid      = errors.New("membership is invalid")
	ErrMembershipExpired      = errors.New("membership has expired")
	ErrMembershipNotInUse     = errors.New("membership is not in use")
	ErrUnhandledMembership    = errors.New("membership type is not handled")
	ErrTransactionNotComplete = errors.New("transaction not completed")

	ErrMissingPaypalTransaction = errors.New("missing paypal subscription transaction")
	ErrMissingPaypalUser        = errors.New("missing user for paypal subscription transaction")
)

func CreateTransactionForUser(ctx context.Context, userID discord.Snowflake, platformType database.PlatformType, transactionStatus database.TransactionStatus, transactionID, currencyCode, amount string) (*database.UserTransactions, error) {
	return Queries.CreateUserTransaction(ctx, database.CreateUserTransactionParams{
		UserID:            int64(userID),
		PlatformType:      int32(platformType),
		TransactionID:     transactionID,
		TransactionStatus: int32(transactionStatus),
		CurrencyCode:      currencyCode,
		Amount:            amount,
	})
}

func CreateMembershipForUser(ctx context.Context, userID discord.Snowflake, transactionUUID uuid.UUID, membershipType database.MembershipType, expiresAt time.Time, guildID *discord.Snowflake) error {
	var guildIDInt int64
	if guildID != nil {
		guildIDInt = int64(*guildID)
	}

	// Create a new membership for the user.
	membership, err := Queries.CreateNewMembership(ctx, database.CreateNewMembershipParams{
		StartedAt:       time.Time{},
		ExpiresAt:       expiresAt,
		Status:          int32(database.MembershipStatusIdle),
		MembershipType:  int32(membershipType),
		TransactionUuid: transactionUUID,
		UserID:          int64(userID),
		GuildID:         guildIDInt,
	})
	if err != nil {
		Logger.Error().Err(err).
			Int64("user_id", int64(userID)).
			Str("transaction_uuid", transactionUUID.String()).
			Msg("Failed to create new membership")

		return err
	}

	session, err := AcquireSession(ctx, GetManagerNameFromContext(ctx))
	if err != nil {
		Logger.Error().Err(err).
			Str("membership_uuid", membership.MembershipUuid.String()).
			Msg("Failed to acquire session")

		return err
	}

	err = NotifyMembershipCreated(ctx, session, *membership)

	if guildIDInt != 0 {
		var userMembership *database.GetUserMembershipsByUserIDRow

		userMemberships, err := Queries.GetUserMembershipsByUserID(ctx, int64(userID))
		if err != nil {
			Logger.Error().Err(err).
				Int64("user_id", int64(userID)).
				Str("membership_uuid", membership.MembershipUuid.String()).
				Msg("Failed to get user memberships")

			return err
		}

		for _, m := range userMemberships {
			if m.MembershipUuid == membership.MembershipUuid {
				userMembership = m

				break
			}
		}

		if userMembership != nil {
			_, err = AddMembershipToServer(ctx, *userMembership, discord.Snowflake(guildIDInt))
			if err != nil {
				Logger.Error().Err(err).
					Str("membership_uuid", membership.MembershipUuid.String()).
					Msg("Failed to add membership to server")
			}
		}
	}

	return err
}

// Triggers when a membership has been added to the server.
func onMembershipAdded(ctx context.Context, beforeMembership database.GetUserMembershipsByUserIDRow, membership database.UpdateUserMembershipParams) error {
	println("HandleMembershipAdded", beforeMembership.MembershipUuid.String(), beforeMembership.UserID, beforeMembership.GuildID, beforeMembership.ExpiresAt.String(), membership.GuildID, membership.ExpiresAt.String())

	// Notify of addition?

	return nil
}

// Triggers when a membership has had its guild transferred.
func onMembershipUpdated(ctx context.Context, beforeMembership database.GetUserMembershipsByUserIDRow, membership database.UpdateUserMembershipParams) error {
	println("HandleMembershipUpdated", beforeMembership.MembershipUuid.String(), beforeMembership.UserID, beforeMembership.GuildID, beforeMembership.ExpiresAt.String(), membership.GuildID, membership.ExpiresAt.String())

	// Notify of transfer?

	return nil
}

// Triggers when a membership has been removed from the server.
func onMembershipRemoved(ctx context.Context, membership database.GetUserMembershipsByUserIDRow) error {
	println("HandleMembershipRemoved", membership.MembershipUuid.String(), membership.UserID, membership.GuildID, membership.ExpiresAt.String())

	// Notify of removal?

	return nil
}

// Triggers when a membership has expired and not renewed.
func OnMembershipExpired(ctx context.Context, membership database.GetUserMembershipsByUserIDRow) error {
	// TODO: implement action

	println("HandleMembershipExpired", membership.MembershipUuid.String(), membership.UserID, membership.GuildID, membership.ExpiresAt.String())

	// Notify of new membership?

	session, err := AcquireSession(ctx, GetManagerNameFromContext(ctx))
	if err != nil {
		Logger.Error().Err(err).
			Str("membership_uuid", membership.MembershipUuid.String()).
			Msg("Failed to acquire session")

		return err
	}

	err = NotifyMembershipExpired(ctx, session, database.UserMemberships{
		MembershipUuid:  membership.MembershipUuid,
		CreatedAt:       membership.CreatedAt,
		UpdatedAt:       membership.UpdatedAt,
		StartedAt:       membership.StartedAt,
		ExpiresAt:       membership.ExpiresAt,
		Status:          membership.Status,
		MembershipType:  membership.MembershipType,
		TransactionUuid: membership.TransactionUuid,
		UserID:          membership.UserID,
		GuildID:         membership.GuildID,
	})
	if err != nil {
		Logger.Error().Err(err).
			Str("membership_uuid", membership.MembershipUuid.String()).
			Msg("Failed to trigger onMembershipAdded")
	}

	return nil
}

// Triggers when a membership has been renewed on expiry.
func onMembershipRenewed(ctx context.Context, membership database.GetUserMembershipsByUserIDRow) error {
	// TODO: implement action

	println("HandleMembershipRenewed", membership.MembershipUuid.String(), membership.UserID, membership.GuildID, membership.ExpiresAt.String())

	return nil
}

func AddMembershipToServer(ctx context.Context, membership database.GetUserMembershipsByUserIDRow, guildID discord.Snowflake) (newMembership database.UpdateUserMembershipParams, err error) {
	// if membership.GuildID != 0 {
	// 	Logger.Error().
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
			Logger.Error().
				Str("membership_uuid", membership.MembershipUuid.String()).
				Int32("transaction_status", membership.TransactionStatus.Int32).
				Msg("Membership transaction is not completed")

			return database.UpdateUserMembershipParams{}, ErrTransactionNotComplete
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

		_, err = Queries.UpdateUserMembership(ctx, newMembership)
		if err != nil {
			Logger.Error().Err(err).
				Str("membership_uuid", membership.MembershipUuid.String()).
				Msg("Failed to update user membership")

			return database.UpdateUserMembershipParams{}, err
		}

		if database.MembershipStatus(membership.Status) == database.MembershipStatusIdle {
			err = onMembershipAdded(ctx, membership, newMembership)
			if err != nil {
				Logger.Error().Err(err).
					Str("membership_uuid", membership.MembershipUuid.String()).
					Msg("Failed to trigger onMembershipAdded")
			}
		} else if database.MembershipStatus(membership.Status) == database.MembershipStatusActive {
			// Triggered when a membership has been transferred to a new guild.
			err = onMembershipUpdated(ctx, membership, newMembership)
			if err != nil {
				Logger.Error().Err(err).
					Str("membership_uuid", membership.MembershipUuid.String()).
					Msg("Failed to trigger onMembershipUpdated")
			}
		}

		return newMembership, nil
	default:
		Logger.Error().
			Str("membership_uuid", membership.MembershipUuid.String()).
			Int32("status", membership.Status).
			Msg("Unhandled membership status")

		return database.UpdateUserMembershipParams{}, ErrUnhandledMembership
	}
}

func RemoveMembershipFromServer(ctx context.Context, membership database.GetUserMembershipsByUserIDRow) (newMembership database.UpdateUserMembershipParams, err error) {
	if membership.GuildID == 0 {
		return database.UpdateUserMembershipParams{}, ErrMembershipNotInUse
	}

	// Only set the status to Idle if the membership is currently Active.
	membership.Status = If(membership.Status == int32(database.MembershipStatusActive), int32(database.MembershipStatusIdle), membership.Status)
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

	_, err = Queries.UpdateUserMembership(ctx, newMembership)
	if err != nil {
		Logger.Error().Err(err).
			Str("membership_uuid", membership.MembershipUuid.String()).
			Msg("Failed to update user membership")

		return database.UpdateUserMembershipParams{}, err
	}

	err = onMembershipRemoved(ctx, membership)
	if err != nil {
		Logger.Error().Err(err).
			Str("membership_uuid", membership.MembershipUuid.String()).
			Msg("Failed to trigger onMembershipRemoved")
	}

	return newMembership, nil
}
