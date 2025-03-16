package welcomer

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"

	"github.com/WelcomerTeam/Discord/discord"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	"github.com/plutov/paypal/v4"
	"github.com/rs/zerolog"
)

type PaypalSale struct {
	ID string `json:"id"`

	Amount paypal.Amount `json:"amount"`
	State  string        `json:"state"`

	CreateTime   time.Time `json:"create_time"`
	ClearingTime time.Time `json:"clearing_time"`
	UpdateTime   time.Time `json:"update_time"`

	BillingAgreementID string `json:"billing_agreement_id"`
}

// Handles all paypal sales.
func HandlePaypalSale(ctx context.Context, logger zerolog.Logger, queries *database.Queries, paypalSale PaypalSale) error {
	println("HandlePaypalSale", paypalSale.ID, paypalSale.State, paypalSale.Amount.Total, paypalSale.Amount.Currency, paypalSale.BillingAgreementID)

	if paypalSale.State != "completed" {
		return nil
	}

	if paypalSale.BillingAgreementID == "" {
		return nil
	}

	memberships, err := queries.GetUserMembershipsByTransactionID(ctx, paypalSale.BillingAgreementID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Error().Err(err).
			Str("transaction_id", paypalSale.BillingAgreementID).
			Msg("Failed to get user memberships")

		return err
	}

	paypalMemberships := make([]*database.GetUserMembershipsByTransactionIDRow, 0, len(memberships))

	for _, membership := range memberships {
		if database.PlatformType(membership.PlatformType) == database.PlatformTypePaypalSubscription &&
			membership.TransactionID == paypalSale.BillingAgreementID &&
			(membership.Status == int32(database.MembershipStatusActive) || membership.Status == int32(database.MembershipStatusIdle)) &&
			membership.ExpiresAt.After(time.Now()) {
			paypalMemberships = append(paypalMemberships, membership)
		}
	}

	membershipExpiration := time.Now().AddDate(0, 1, 7)

	for _, membership := range paypalMemberships {
		membership.ExpiresAt = membershipExpiration

		_, err = queries.UpdateUserMembership(ctx, database.UpdateUserMembershipParams{
			MembershipUuid:  membership.MembershipUuid,
			StartedAt:       membership.StartedAt,
			ExpiresAt:       membership.ExpiresAt,
			Status:          membership.Status,
			TransactionUuid: membership.TransactionUuid,
			UserID:          membership.UserID,
			GuildID:         membership.GuildID,
		})
		if err != nil {
			logger.Error().Err(err).
				Msg("Failed to update membership")

			return err
		}

		logger.Info().
			Str("membership_uuid", membership.MembershipUuid.String()).
			Int64("user_id", membership.UserID).
			Int64("guild_id", membership.GuildID).
			Time("expires_at", membership.ExpiresAt).
			Msg("Updated membership expiration due to paypal sale")
	}

	return nil
}

// Handles all subscription events.
func HandlePaypalSubscription(ctx context.Context, logger zerolog.Logger, queries *database.Queries, paypalSubscription paypal.Subscription) error {
	println("HandlePaypalSubscription", paypalSubscription.ID, paypalSubscription.PlanID, paypalSubscription.SubscriptionStatus, paypalSubscription.Quantity)

	if paypalSubscription.SubscriptionStatus == paypal.SubscriptionStatusApprovalPending ||
		paypalSubscription.SubscriptionStatus == paypal.SubscriptionStatusCancelled ||
		paypalSubscription.SubscriptionStatus == paypal.SubscriptionStatusExpired ||
		paypalSubscription.SubscriptionStatus == paypal.SubscriptionStatusSuspended {
		paypalSubscription.Quantity = "0"
	}

	paypalSub, err := queries.GetPaypalSubscriptionBySubscriptionID(ctx, paypalSubscription.ID)
	if err != nil {
		logger.Error().Err(err).
			Str("subscription_id", paypalSubscription.ID).
			Msg("Failed to get paypal subscription")
	}

	var userID discord.Snowflake

	if paypalSub != nil {
		userID = discord.Snowflake(paypalSub.UserID)
	}

	txs, err := queries.GetUserTransactionsByTransactionID(ctx, paypalSubscription.ID)
	if err != nil {
		logger.Error().Err(err).
			Str("subscription_id", paypalSubscription.ID).
			Msg("Failed to get user transactions")

		return err
	}

	var paypalTxs *database.UserTransactions

	for _, tx := range txs {
		if tx.PlatformType == int32(database.PlatformTypePaypalSubscription) &&
			tx.TransactionID == paypalSubscription.ID {
			paypalTxs = tx

			break
		}
	}

	// If no paypal transaction is found, error.
	if paypalTxs == nil {
		logger.Error().
			Str("subscription_id", paypalSubscription.ID).
			Msg("No paypal transaction found")

		return ErrMissingPaypalTransaction
	}

	if userID.IsNil() {
		userID = discord.Snowflake(paypalTxs.UserID)

		logger.Warn().
			Str("subscription_id", paypalSubscription.ID).
			Int64("user_id", int64(userID)).
			Msg("Falling back to transaction user ID")
	}

	// Update paypal subscription entry.
	_, err = queries.CreateOrUpdatePaypalSubscription(ctx, database.CreateOrUpdatePaypalSubscriptionParams{
		SubscriptionID:     paypalSubscription.ID,
		UserID:             int64(userID),
		PayerID:            paypalSubscription.Subscriber.PayerID,
		LastBilledAt:       paypalSubscription.BillingInfo.LastPayment.Time,
		NextBillingAt:      paypalSubscription.BillingInfo.NextBillingTime,
		SubscriptionStatus: string(paypalSubscription.SubscriptionStatus),
		PlanID:             paypalSubscription.PlanID,
		Quantity:           paypalSubscription.Quantity,
	})
	if err != nil {
		logger.Error().Err(err).
			Str("subscription_id", paypalSubscription.ID).
			Msg("Failed to create or update paypal subscription")
	}

	if userID.IsNil() {
		logger.Error().
			Str("subscription_id", paypalSubscription.ID).
			Msg("No user ID found")

		return ErrMissingPaypalUser
	}

	memberships, err := queries.GetUserMembershipsByUserID(ctx, int64(userID))
	if err != nil {
		logger.Error().Err(err).
			Int64("user_id", int64(userID)).
			Msg("Failed to get user memberships")

		return err
	}

	paypalMemberships := make([]*database.GetUserMembershipsByUserIDRow, 0, len(memberships))

	for _, membership := range memberships {
		if database.PlatformType(membership.PlatformType.Int32) == database.PlatformTypePaypalSubscription &&
			membership.TransactionUuid == paypalTxs.TransactionUuid &&
			(membership.Status == int32(database.MembershipStatusActive) || membership.Status == int32(database.MembershipStatusIdle)) &&
			membership.ExpiresAt.After(time.Now()) {
			paypalMemberships = append(paypalMemberships, membership)
		}
	}

	var membershipExpiration time.Time
	if paypalSubscription.BillingInfo.FailedPaymentsCount == 0 {
		membershipExpiration = time.Now().AddDate(0, 1, 7)
	} else {
		// Payment failed
		membershipExpiration = time.Now().AddDate(0, 0, 7)
	}

	eligibleMemberships, err := strconv.Atoi(paypalSubscription.Quantity)
	if err != nil {
		logger.Warn().Err(err).
			Str("subscription_id", paypalSubscription.ID).
			Str("quantity", paypalSubscription.Quantity).
			Msg("Failed to parse quantity")

		eligibleMemberships = 1
	}

	if len(paypalMemberships) > eligibleMemberships {
		logger.Warn().
			Int64("user_id", int64(userID)).
			Int("current_memberships", len(paypalMemberships)).
			Int("eligible_memberships", eligibleMemberships).
			Msg("User has too many memberships")

		// Remove memberships. Let them expire naturally.
	} else if len(paypalMemberships) < eligibleMemberships {
		logger.Info().
			Int64("user_id", int64(userID)).
			Int("current_memberships", len(paypalMemberships)).
			Int("eligible_memberships", eligibleMemberships).
			Msg("User has too few memberships")

		// Add memberships

		for i := len(paypalMemberships); i < eligibleMemberships; i++ {
			err = CreateMembershipForUser(ctx, queries, userID, paypalTxs.TransactionUuid, database.MembershipTypeWelcomerPro, membershipExpiration, nil)
			if err != nil {
				logger.Error().Err(err).
					Int64("user_id", int64(userID)).
					Msg("Failed to create membership")

				return err
			}
		}

		for _, membership := range paypalMemberships {
			membership.ExpiresAt = membershipExpiration

			_, err = queries.UpdateUserMembership(ctx, database.UpdateUserMembershipParams{
				MembershipUuid:  membership.MembershipUuid,
				StartedAt:       membership.StartedAt,
				ExpiresAt:       membership.ExpiresAt,
				Status:          membership.Status,
				TransactionUuid: membership.TransactionUuid,
				UserID:          membership.UserID,
				GuildID:         membership.GuildID,
			})
			if err != nil {
				logger.Error().Err(err).
					Int64("user_id", int64(userID)).
					Msg("Failed to update membership")

				return err
			}

			logger.Info().
				Str("membership_uuid", membership.MembershipUuid.String()).
				Int64("user_id", membership.UserID).
				Int64("guild_id", membership.GuildID).
				Time("expires_at", membership.ExpiresAt).
				Msg("Updated membership expiration paypal subscription")
		}
	} else {
		logger.Info().
			Int64("user_id", int64(userID)).
			Int("current_memberships", len(paypalMemberships)).
			Int("eligible_memberships", eligibleMemberships).
			Msg("User has correct number of memberships")

		// No changes
	}

	return nil
}

// Triggers when a subscription is created.
func OnPaypalSubscriptionCreated(ctx context.Context, logger zerolog.Logger, queries *database.Queries, paypalSubscription paypal.Subscription) error {
	println("HandlePaypalSubscriptionCreated", paypalSubscription.ID)

	return HandlePaypalSubscription(ctx, logger, queries, paypalSubscription)
}

// Triggers when a subscription has been activated.
func OnPaypalSubscriptionActivated(ctx context.Context, logger zerolog.Logger, queries *database.Queries, paypalSubscription paypal.Subscription) error {
	println("HandlePaypalSubscriptionActivated", paypalSubscription.ID)

	return HandlePaypalSubscription(ctx, logger, queries, paypalSubscription)
}

// Triggers when a subscription has been updated.
func OnPaypalSubscriptionUpdated(ctx context.Context, logger zerolog.Logger, queries *database.Queries, paypalSubscription paypal.Subscription) error {
	println("HandlePaypalSubscriptionUpdated", paypalSubscription.ID)

	return HandlePaypalSubscription(ctx, logger, queries, paypalSubscription)
}

// Triggers when a subscription has been re-activated.
func OnPaypalSubscriptionReactivated(ctx context.Context, logger zerolog.Logger, queries *database.Queries, paypalSubscription paypal.Subscription) error {
	println("HandlePaypalSubscriptionReactivated", paypalSubscription.ID)

	return HandlePaypalSubscription(ctx, logger, queries, paypalSubscription)
}

// Triggers when a subscription has expired.
func OnPaypalSubscriptionExpired(ctx context.Context, logger zerolog.Logger, queries *database.Queries, paypalSubscription paypal.Subscription) error {
	println("HandlePaypalSubscriptionExpired", paypalSubscription.ID)

	paypalSubscription.Quantity = "0"

	return HandlePaypalSubscription(ctx, logger, queries, paypalSubscription)
}

// Triggers when a subscription has been cancelled.
func OnPaypalSubscriptionCancelled(ctx context.Context, logger zerolog.Logger, queries *database.Queries, paypalSubscription paypal.Subscription) error {
	println("HandlePaypalSubscriptionCancelled", paypalSubscription.ID)

	paypalSubscription.Quantity = "0"

	return HandlePaypalSubscription(ctx, logger, queries, paypalSubscription)
}

// Triggers when a subscription has been suspended.
func OnPaypalSubscriptionSuspended(ctx context.Context, logger zerolog.Logger, queries *database.Queries, paypalSubscription paypal.Subscription) error {
	println("HandlePaypalSubscriptionSuspended", paypalSubscription.ID)

	paypalSubscription.Quantity = "0"

	return HandlePaypalSubscription(ctx, logger, queries, paypalSubscription)
}

// Triggers when a subscription payment has failed.
func OnPaypalSubscriptionPaymentFailed(ctx context.Context, logger zerolog.Logger, queries *database.Queries, paypalSubscription paypal.Subscription) error {
	println("HandlePaypalSubscriptionPaymentFailed", paypalSubscription.ID)

	return HandlePaypalSubscription(ctx, logger, queries, paypalSubscription)
}
