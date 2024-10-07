package welcomer

import (
	"context"
	"errors"
	"time"

	"github.com/WelcomerTeam/Discord/discord"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	utils "github.com/WelcomerTeam/Welcomer/welcomer-utils"
	"github.com/rs/zerolog"
)

var (
	ErrMembershipAlreadyInUse = errors.New("membership is already in use")
	ErrMembershipInvalid      = errors.New("membership is invalid")
	ErrMembershipExpired      = errors.New("membership has expired")
	ErrMembershipNotInUse     = errors.New("membership is not in use")
	ErrUnhandledMembership    = errors.New("membership type is not handled")
	ErrTransactionNotComplete = errors.New("transaction not completed")
)

// Membership Events

// Triggers when a patreon account has been linked.
func onPatreonLinked(ctx context.Context, logger zerolog.Logger, queries *database.Queries, patreonUser database.PatreonUsers) error {
	// TODO: implement action
	return nil
}

// Triggers when a patreon tier has changed.
func onPatreonTierChanged(ctx context.Context, logger zerolog.Logger, queries *database.Queries, beforePatreonUser, patreonUser database.PatreonUsers) error {
	// TODO: implement action
	return nil
}

// Triggers when a membership has been added to the server.
func onMembershipAdded(ctx context.Context, logger zerolog.Logger, queries *database.Queries, beforeMembership, membership database.GetUserMembershipsByUserIDRow) error {
	// TODO: implement action
	return nil
}

// Triggers when a membership has had its guild transferred.
func onMembershipUpdated(ctx context.Context, logger zerolog.Logger, queries *database.Queries, beforeMembership, membership database.GetUserMembershipsByUserIDRow) error {
	// TODO: implement action
	return nil
}

// Triggers when a membership has been removed from the server.
func onMembershipRemoved(ctx context.Context, logger zerolog.Logger, queries *database.Queries, membership database.GetUserMembershipsByUserIDRow) error {
	// TODO: implement action
	return nil
}

// Triggers when a membership has expired and not renewed.
func onMembershipExpired(ctx context.Context, logger zerolog.Logger, queries *database.Queries, membership database.GetUserMembershipsByUserIDRow) error {
	// TODO: implement action
	return nil
}

// Triggers when a membership has been renewed on expiry.
func onMembershipRenewed(ctx context.Context, logger zerolog.Logger, queries *database.Queries, membership database.GetUserMembershipsByUserIDRow) error {
	// TODO: implement action
	return nil
}

func AddMembershipToServer(ctx context.Context, logger zerolog.Logger, queries *database.Queries, membership database.GetUserMembershipsByUserIDRow, guildID discord.Snowflake) (newMembership database.UpdateUserMembershipParams, err error) {
	// if membership.GuildID != 0 {
	// 	logger.Error().
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
			logger.Error().
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

		_, err = queries.UpdateUserMembership(ctx, newMembership)
		if err != nil {
			logger.Error().Err(err).
				Str("membership_uuid", membership.MembershipUuid.String()).
				Msg("Failed to update user membership")

			return database.UpdateUserMembershipParams{}, err
		}

		return newMembership, nil
	default:
		logger.Error().
			Str("membership_uuid", membership.MembershipUuid.String()).
			Int32("status", membership.Status).
			Msg("Unhandled membership status")

		return database.UpdateUserMembershipParams{}, ErrUnhandledMembership
	}
}

func RemoveMembershipFromServer(ctx context.Context, logger zerolog.Logger, queries *database.Queries, membership database.GetUserMembershipsByUserIDRow) (newMembership database.UpdateUserMembershipParams, err error) {
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

	_, err = queries.UpdateUserMembership(ctx, newMembership)
	if err != nil {
		logger.Error().Err(err).
			Str("membership_uuid", membership.MembershipUuid.String()).
			Msg("Failed to update user membership")

		return database.UpdateUserMembershipParams{}, err
	}

	return newMembership, nil
}
