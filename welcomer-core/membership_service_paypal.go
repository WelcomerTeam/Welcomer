package welcomer

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/WelcomerTeam/Discord/discord"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	"github.com/plutov/paypal/v4"
)

var DiscordPaymentsWebhookURL = os.Getenv("DISCORD_PAYMENTS_WEBHOOK_URL")

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
func HandlePaypalSale(ctx context.Context, paypalSale PaypalSale) error {
	println("HandlePaypalSale", paypalSale.ID, paypalSale.State, paypalSale.Amount.Total, paypalSale.Amount.Currency, paypalSale.BillingAgreementID)

	memberships, err := Queries.GetUserMembershipsByTransactionID(ctx, paypalSale.BillingAgreementID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		Logger.Error().Err(err).
			Str("transaction_id", paypalSale.BillingAgreementID).
			Msg("Failed to get user memberships")

		return err
	}

	paypalMemberships := make([]*database.GetUserMembershipsByTransactionIDRow, 0, len(memberships))

	var userID discord.Snowflake

	for _, membership := range memberships {
		if database.PlatformType(membership.PlatformType) == database.PlatformTypePaypalSubscription &&
			membership.TransactionID == paypalSale.BillingAgreementID &&
			(membership.Status == int32(database.MembershipStatusActive) || membership.Status == int32(database.MembershipStatusIdle) || membership.Status == int32(database.MembershipStatusExpired)) {
			paypalMemberships = append(paypalMemberships, membership)
			userID = discord.Snowflake(membership.UserID)
		}
	}

	if DiscordPaymentsWebhookURL != "" {
		err := SendWebhookMessage(ctx, DiscordPaymentsWebhookURL, discord.WebhookMessageParams{
			Embeds: []discord.Embed{
				{
					Title:     "Paypal Sale",
					Color:     0x009CDE,
					Timestamp: ToPointer(time.Now()),
					Fields: []discord.EmbedField{
						{
							Name:   "ID",
							Value:  paypalSale.ID,
							Inline: true,
						},
						{
							Name:   "Billing Agreement ID",
							Value:  paypalSale.BillingAgreementID,
							Inline: true,
						},
						{
							Name:   "User",
							Value:  "<@" + userID.String() + "> " + userID.String(),
							Inline: true,
						},
						{
							Name:   "Amount",
							Value:  paypalSale.Amount.Total + " " + paypalSale.Amount.Currency,
							Inline: true,
						},
						{
							Name:   "State",
							Value:  paypalSale.State,
							Inline: true,
						},
						{
							Name:   "Created",
							Value:  "<t:" + Itoa(time.Time(paypalSale.CreateTime).Unix()) + ":R>",
							Inline: true,
						},
						{
							Name:   "Cleared",
							Value:  "<t:" + Itoa(time.Time(paypalSale.ClearingTime).Unix()) + ":R>",
							Inline: true,
						},
						{
							Name:   "Updated",
							Value:  "<t:" + Itoa(time.Time(paypalSale.UpdateTime).Unix()) + ":R>",
							Inline: true,
						},
					},
				},
			},
		})
		if err != nil {
			Logger.Error().Err(err).Msg("Failed to send webhook message")
		}
	}

	if paypalSale.State != "completed" {
		return nil
	}

	if paypalSale.BillingAgreementID == "" {
		return nil
	}

	membershipExpiration := time.Now().AddDate(0, 1, 7)

	for _, membership := range paypalMemberships {
		membership.ExpiresAt = membershipExpiration

		_, err = Queries.UpdateUserMembership(ctx, database.UpdateUserMembershipParams{
			MembershipUuid:  membership.MembershipUuid,
			StartedAt:       membership.StartedAt,
			ExpiresAt:       membership.ExpiresAt,
			Status:          membership.Status,
			TransactionUuid: membership.TransactionUuid,
			UserID:          membership.UserID,
			GuildID:         membership.GuildID,
		})
		if err != nil {
			Logger.Error().Err(err).
				Msg("Failed to update membership")

			return err
		}

		Logger.Info().
			Str("membership_uuid", membership.MembershipUuid.String()).
			Int64("user_id", membership.UserID).
			Int64("guild_id", membership.GuildID).
			Time("expires_at", membership.ExpiresAt).
			Msg("Updated membership expiration due to paypal sale")
	}

	return nil
}

// Handles all subscription events.
func HandlePaypalSubscription(ctx context.Context, action string, paypalSubscription paypal.Subscription) error {
	println("HandlePaypalSubscription", paypalSubscription.ID, paypalSubscription.PlanID, paypalSubscription.SubscriptionStatus, paypalSubscription.Quantity)

	if paypalSubscription.SubscriptionStatus == paypal.SubscriptionStatusApprovalPending ||
		paypalSubscription.SubscriptionStatus == paypal.SubscriptionStatusCancelled ||
		paypalSubscription.SubscriptionStatus == paypal.SubscriptionStatusExpired ||
		paypalSubscription.SubscriptionStatus == paypal.SubscriptionStatusSuspended {
		paypalSubscription.Quantity = "0"
	}

	paypalSub, err := Queries.GetPaypalSubscriptionBySubscriptionID(ctx, paypalSubscription.ID)
	if err != nil {
		Logger.Error().Err(err).
			Str("subscription_id", paypalSubscription.ID).
			Msg("Failed to get paypal subscription")
	}

	var userID discord.Snowflake

	if paypalSub != nil {
		userID = discord.Snowflake(paypalSub.UserID)
	}

	txs, err := Queries.GetUserTransactionsByTransactionID(ctx, paypalSubscription.ID)
	if err != nil {
		Logger.Error().Err(err).
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

	if userID.IsNil() && paypalTxs != nil {
		userID = discord.Snowflake(paypalTxs.UserID)

		Logger.Warn().
			Str("subscription_id", paypalSubscription.ID).
			Int64("user_id", int64(userID)).
			Msg("Falling back to transaction user ID")
	}

	// If no paypal transaction is found, error.
	if paypalTxs == nil {
		Logger.Error().
			Str("subscription_id", paypalSubscription.ID).
			Msg("No paypal transaction found")

		return ErrMissingPaypalTransaction
	}

	if DiscordPaymentsWebhookURL != "" {
		var startTime time.Time

		if paypalSubscription.StartTime != nil {
			startTime = time.Time(*paypalSubscription.StartTime)
		}

		err := SendWebhookMessage(ctx, DiscordPaymentsWebhookURL, discord.WebhookMessageParams{
			Embeds: []discord.Embed{
				{
					Title:     "Paypal Subscription " + action,
					Color:     0x009CDE,
					Timestamp: ToPointer(time.Now()),
					Fields: []discord.EmbedField{
						{
							Name:   "ID",
							Value:  paypalSubscription.ID,
							Inline: true,
						},
						{
							Name:   "CID",
							Value:  paypalSubscription.CustomID,
							Inline: true,
						},
						{
							Name:   "Plan ID",
							Value:  paypalSubscription.PlanID,
							Inline: true,
						},
						{
							Name:   "User",
							Value:  "<@" + userID.String() + "> " + userID.String(),
							Inline: true,
						},
						{
							Name:   "Amount",
							Value:  paypalSubscription.BillingInfo.LastPayment.Amount.Value + " " + paypalSubscription.BillingInfo.LastPayment.Amount.Currency,
							Inline: true,
						},
						{
							Name:   "Quantity",
							Value:  paypalSubscription.Quantity,
							Inline: true,
						},
						{
							Name:   "State",
							Value:  string(paypalSubscription.SubscriptionStatus),
							Inline: true,
						},
						{
							Name:   "Created",
							Value:  "<t:" + Itoa(time.Time(TryParseTime(paypalSubscription.CreateTime)).Unix()) + ":R>",
							Inline: true,
						},
						{
							Name:   "Started",
							Value:  "<t:" + Itoa(time.Time(startTime).Unix()) + ":R>",
							Inline: true,
						},
						{
							Name:   "Updated",
							Value:  "<t:" + Itoa(time.Time(TryParseTime(paypalSubscription.UpdateTime)).Unix()) + ":R>",
							Inline: true,
						},
						{
							Name:   "Next Billing",
							Value:  "<t:" + Itoa(time.Time(paypalSubscription.BillingInfo.NextBillingTime).Unix()) + ":R>",
							Inline: true,
						},
					},
				},
			},
		})
		if err != nil {
			Logger.Error().Err(err).Msg("Failed to send webhook message")
		}
	}

	// Update paypal subscription entry.
	_, err = Queries.CreateOrUpdatePaypalSubscription(ctx, database.CreateOrUpdatePaypalSubscriptionParams{
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
		Logger.Error().Err(err).
			Str("subscription_id", paypalSubscription.ID).
			Msg("Failed to create or update paypal subscription")
	}

	if userID.IsNil() {
		Logger.Error().
			Str("subscription_id", paypalSubscription.ID).
			Msg("No user ID found")

		return ErrMissingPaypalUser
	}

	memberships, err := Queries.GetUserMembershipsByUserID(ctx, int64(userID))
	if err != nil {
		Logger.Error().Err(err).
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
		Logger.Warn().Err(err).
			Str("subscription_id", paypalSubscription.ID).
			Str("quantity", paypalSubscription.Quantity).
			Msg("Failed to parse quantity")

		eligibleMemberships = 1
	}

	if len(paypalMemberships) > eligibleMemberships {
		Logger.Warn().
			Int64("user_id", int64(userID)).
			Int("current_memberships", len(paypalMemberships)).
			Int("eligible_memberships", eligibleMemberships).
			Msg("User has too many memberships")

		// Remove memberships. Let them expire naturally.
	} else if len(paypalMemberships) < eligibleMemberships {
		Logger.Info().
			Int64("user_id", int64(userID)).
			Int("current_memberships", len(paypalMemberships)).
			Int("eligible_memberships", eligibleMemberships).
			Msg("User has too few memberships")

		// Add memberships

		for i := len(paypalMemberships); i < eligibleMemberships; i++ {
			err = CreateMembershipForUser(ctx, userID, paypalTxs.TransactionUuid, database.MembershipTypeWelcomerPro, membershipExpiration, nil)
			if err != nil {
				Logger.Error().Err(err).
					Int64("user_id", int64(userID)).
					Msg("Failed to create membership")

				return err
			}
		}

		for _, membership := range paypalMemberships {
			membership.ExpiresAt = membershipExpiration

			_, err = Queries.UpdateUserMembership(ctx, database.UpdateUserMembershipParams{
				MembershipUuid:  membership.MembershipUuid,
				StartedAt:       membership.StartedAt,
				ExpiresAt:       membership.ExpiresAt,
				Status:          membership.Status,
				TransactionUuid: membership.TransactionUuid,
				UserID:          membership.UserID,
				GuildID:         membership.GuildID,
			})
			if err != nil {
				Logger.Error().Err(err).
					Int64("user_id", int64(userID)).
					Msg("Failed to update membership")

				return err
			}

			Logger.Info().
				Str("membership_uuid", membership.MembershipUuid.String()).
				Int64("user_id", membership.UserID).
				Int64("guild_id", membership.GuildID).
				Time("expires_at", membership.ExpiresAt).
				Msg("Updated membership expiration paypal subscription")
		}
	} else {
		Logger.Info().
			Int64("user_id", int64(userID)).
			Int("current_memberships", len(paypalMemberships)).
			Int("eligible_memberships", eligibleMemberships).
			Msg("User has correct number of memberships")

		// No changes
	}

	return nil
}

// Triggers when a subscription is created.
func OnPaypalSubscriptionCreated(ctx context.Context, paypalSubscription paypal.Subscription) error {
	println("HandlePaypalSubscriptionCreated", paypalSubscription.ID)

	return HandlePaypalSubscription(ctx, "created", paypalSubscription)
}

// Triggers when a subscription has been activated.
func OnPaypalSubscriptionActivated(ctx context.Context, paypalSubscription paypal.Subscription) error {
	println("HandlePaypalSubscriptionActivated", paypalSubscription.ID)

	return HandlePaypalSubscription(ctx, "activated", paypalSubscription)
}

// Triggers when a subscription has been updated.
func OnPaypalSubscriptionUpdated(ctx context.Context, paypalSubscription paypal.Subscription) error {
	println("HandlePaypalSubscriptionUpdated", paypalSubscription.ID)

	return HandlePaypalSubscription(ctx, "updated", paypalSubscription)
}

// Triggers when a subscription has been re-activated.
func OnPaypalSubscriptionReactivated(ctx context.Context, paypalSubscription paypal.Subscription) error {
	println("HandlePaypalSubscriptionReactivated", paypalSubscription.ID)

	return HandlePaypalSubscription(ctx, "reactivated", paypalSubscription)
}

// Triggers when a subscription has expired.
func OnPaypalSubscriptionExpired(ctx context.Context, paypalSubscription paypal.Subscription) error {
	println("HandlePaypalSubscriptionExpired", paypalSubscription.ID)

	paypalSubscription.Quantity = "0"

	return HandlePaypalSubscription(ctx, "expired", paypalSubscription)
}

// Triggers when a subscription has been cancelled.
func OnPaypalSubscriptionCancelled(ctx context.Context, paypalSubscription paypal.Subscription) error {
	println("HandlePaypalSubscriptionCancelled", paypalSubscription.ID)

	paypalSubscription.Quantity = "0"

	return HandlePaypalSubscription(ctx, "cancelled", paypalSubscription)
}

// Triggers when a subscription has been suspended.
func OnPaypalSubscriptionSuspended(ctx context.Context, paypalSubscription paypal.Subscription) error {
	println("HandlePaypalSubscriptionSuspended", paypalSubscription.ID)

	paypalSubscription.Quantity = "0"

	return HandlePaypalSubscription(ctx, "suspended", paypalSubscription)
}

// Triggers when a subscription payment has failed.
func OnPaypalSubscriptionPaymentFailed(ctx context.Context, paypalSubscription paypal.Subscription) error {
	println("HandlePaypalSubscriptionPaymentFailed", paypalSubscription.ID)

	return HandlePaypalSubscription(ctx, "payment failed", paypalSubscription)
}
