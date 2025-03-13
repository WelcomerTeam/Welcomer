package welcomer

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"

	"github.com/WelcomerTeam/Discord/discord"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	utils "github.com/WelcomerTeam/Welcomer/welcomer-utils"
	"github.com/gofrs/uuid"
	"github.com/plutov/paypal/v4"
	"github.com/rs/zerolog"
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

func CreateTransactionForUser(ctx context.Context, queries *database.Queries, userID discord.Snowflake, platformType database.PlatformType, transactionStatus database.TransactionStatus, transactionID, currencyCode, amount string) (*database.UserTransactions, error) {
	return queries.CreateUserTransaction(ctx, database.CreateUserTransactionParams{
		UserID:            int64(userID),
		PlatformType:      int32(platformType),
		TransactionID:     transactionID,
		TransactionStatus: int32(transactionStatus),
		CurrencyCode:      currencyCode,
		Amount:            amount,
	})
}

func CreateMembershipForUser(ctx context.Context, queries *database.Queries, userID discord.Snowflake, transactionUUID uuid.UUID, membershipType database.MembershipType, expiresAt time.Time, guildID *discord.Snowflake) error {
	var guildIDInt int64
	if guildID != nil {
		guildIDInt = int64(*guildID)
	}

	// Create a new membership for the user.
	_, err := queries.CreateNewMembership(ctx, database.CreateNewMembershipParams{
		StartedAt:       time.Time{},
		ExpiresAt:       expiresAt,
		Status:          int32(database.MembershipStatusIdle),
		MembershipType:  int32(membershipType),
		TransactionUuid: transactionUUID,
		UserID:          int64(userID),
		GuildID:         guildIDInt,
	})

	return err
}

// Membership Events

// Triggers when a patreon tier has changed.
func OnPatreonTierChanged(ctx context.Context, logger zerolog.Logger, queries *database.Queries, beforePatreonUser *database.PatreonUsers, patreonUser database.CreateOrUpdatePatreonUserParams) error {
	println("HandlePatreonTierChanged", beforePatreonUser.PatreonUserID, beforePatreonUser.TierID, patreonUser.TierID)

	eligibleMemberships := 0

	membershipType := database.MembershipTypeUnknown

	switch PatreonTier(patreonUser.TierID) {
	case PatreonTierUnpublishedWelcomerDonator:
		membershipType = database.MembershipTypeLegacyCustomBackgrounds
	case PatreonTierUnpublishedWelcomerPro1:
		eligibleMemberships = 1
		membershipType = database.MembershipTypeLegacyWelcomerPro
	case PatreonTierUnpublishedWelcomerPro3:
		eligibleMemberships = 3
		membershipType = database.MembershipTypeLegacyWelcomerPro
	case PatreonTierUnpublishedWelcomerPro5:
		eligibleMemberships = 5
		membershipType = database.MembershipTypeLegacyWelcomerPro
	case PatreonTierWelcomerPro:
		eligibleMemberships = 1
		membershipType = database.MembershipTypeWelcomerPro
	case PatreonTierFree:
		eligibleMemberships = 0
	}

	txs, err := queries.GetUserTransactionsByUserID(ctx, int64(patreonUser.UserID))
	if err != nil {
		logger.Error().Err(err).
			Int64("user_id", int64(patreonUser.UserID)).
			Msg("Failed to get user transactions")

		return err
	}

	var patreonTxs *database.UserTransactions

	for _, tx := range txs {
		if tx.PlatformType == int32(database.PlatformTypePatreon) {
			patreonTxs = tx

			break
		}
	}

	// If no patreon transaction is found, create one.
	if patreonTxs == nil {
		logger.Warn().
			Int64("user_id", int64(patreonUser.UserID)).
			Msg("No patreon transaction found")

		tx, err := CreateTransactionForUser(ctx, queries, discord.Snowflake(patreonUser.UserID), database.PlatformTypePatreon, database.TransactionStatusCompleted, "", "", "")
		if err != nil {
			logger.Error().Err(err).
				Int64("user_id", int64(patreonUser.UserID)).
				Msg("Failed to create patreon transaction")

			return err
		}

		patreonTxs = tx
	}

	memberships, err := queries.GetUserMembershipsByUserID(ctx, int64(patreonUser.UserID))
	if err != nil {
		logger.Error().Err(err).
			Int64("user_id", int64(patreonUser.UserID)).
			Msg("Failed to get user memberships")

		return err
	}

	patreonMemberships := make([]*database.GetUserMembershipsByUserIDRow, 0, len(memberships))

	for _, membership := range memberships {
		if database.MembershipType(membership.PlatformType.Int32) == database.MembershipType(database.PlatformTypePatreon) &&
			membership.TransactionUuid == patreonTxs.TransactionUuid &&
			(membership.Status == int32(database.MembershipStatusActive) || membership.Status == int32(database.MembershipStatusIdle)) &&
			membership.ExpiresAt.After(time.Now()) {
			patreonMemberships = append(patreonMemberships, membership)
		}
	}

	var membershipExpiration time.Time
	if LastChargeStatus(patreonUser.LastChargeStatus) == LastChargeStatusPaid {
		membershipExpiration = time.Now().AddDate(0, 1, 0)
	} else {
		membershipExpiration = time.Now().AddDate(0, 0, 7)
	}

	if len(patreonMemberships) > eligibleMemberships {
		logger.Warn().
			Int64("user_id", int64(patreonUser.UserID)).
			Int("current_memberships", len(patreonMemberships)).
			Int("eligible_memberships", eligibleMemberships).
			Msg("User has too many memberships")

		// Remove memberships. Let them expire naturally.
	} else if len(patreonMemberships) < eligibleMemberships {
		logger.Info().
			Int64("user_id", int64(patreonUser.UserID)).
			Int("current_memberships", len(patreonMemberships)).
			Int("eligible_memberships", eligibleMemberships).
			Msg("User has too few memberships")

		// Add memberships

		for i := len(patreonMemberships); i < eligibleMemberships; i++ {
			err = CreateMembershipForUser(ctx, queries, discord.Snowflake(patreonUser.UserID), patreonTxs.TransactionUuid, membershipType, membershipExpiration, nil)
			if err != nil {
				logger.Error().Err(err).
					Int64("user_id", int64(patreonUser.UserID)).
					Msg("Failed to create membership")

				return err
			}
		}

		for _, membership := range patreonMemberships {
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
					Int64("user_id", int64(patreonUser.UserID)).
					Msg("Failed to update membership")

				return err
			}
		}
	} else {
		logger.Info().
			Int64("user_id", int64(patreonUser.UserID)).
			Int("current_memberships", len(patreonMemberships)).
			Int("eligible_memberships", eligibleMemberships).
			Msg("User has correct number of memberships")

		// No changes
	}

	// Update pledge status
	if patreonUser.TierID == 0 {
		patreonUser.PledgeEndedAt = time.Time{}
	} else {
		patreonUser.PledgeCreatedAt = time.Now()
		patreonUser.PledgeEndedAt = time.Time{}
	}

	_, err = queries.CreateOrUpdatePatreonUser(ctx, database.CreateOrUpdatePatreonUserParams{
		PatreonUserID:    patreonUser.PatreonUserID,
		UserID:           patreonUser.UserID,
		ThumbUrl:         patreonUser.ThumbUrl,
		PledgeCreatedAt:  patreonUser.PledgeCreatedAt,
		PledgeEndedAt:    patreonUser.PledgeEndedAt,
		TierID:           patreonUser.TierID,
		FullName:         patreonUser.FullName,
		Email:            patreonUser.Email,
		LastChargeStatus: patreonUser.LastChargeStatus,
		PatronStatus:     patreonUser.PatronStatus,
	})
	if err != nil {
		logger.Error().Err(err).
			Int64("user_id", int64(patreonUser.UserID)).
			Msg("Failed to create or update patreon user")

		return err
	}

	return nil
}

// Triggers when a patreon tier has changed. Ran if OnPatreonTierChange failed to run.
func OnPatreonTierChanged_Fallback(ctx context.Context, logger zerolog.Logger, queries *database.Queries, beforePatreonUser *database.PatreonUsers, patreonUser database.CreateOrUpdatePatreonUserParams, e error) error {
	println("HandlePatreonTierChanged_Fallback", beforePatreonUser.PatreonUserID, beforePatreonUser.TierID, patreonUser.TierID, e.Error())

	// TODO: Log to file/webhook

	return nil
}

// Triggers when a patron is no longer pledging
func OnPatreonNoLongerPledging(ctx context.Context, logger zerolog.Logger, queries *database.Queries, patreonUser database.PatreonUsers, patreonMember PatreonMember) error {
	println("HandlePatreonNoLongerPledging", patreonUser.PatreonUserID, patreonUser.UserID, patreonUser.TierID, patreonMember.Attributes.PatronStatus)

	memberships, err := queries.GetUserMembershipsByUserID(ctx, int64(patreonUser.UserID))
	if err != nil {
		logger.Error().Err(err).
			Int64("user_id", int64(patreonUser.UserID)).
			Msg("Failed to get user memberships")

		return err
	}

	var membershipExpiration time.Time
	if patreonMember.Attributes.LastChargeStatus == LastChargeStatusPaid {
		membershipExpiration = patreonMember.Attributes.LastChargeDate.AddDate(0, 1, 0)
	} else {
		membershipExpiration = time.Now().AddDate(0, 0, 7)
	}

	// If no longer pledging, ensure membership lasts a month after last purchase.
	for _, membership := range memberships {
		if database.MembershipType(membership.PlatformType.Int32) == database.MembershipType(database.PlatformTypePatreon) && (membership.Status == int32(database.MembershipStatusActive) || membership.Status == int32(database.MembershipStatusIdle)) && membership.ExpiresAt.After(time.Now()) {
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
					Int64("user_id", int64(patreonUser.UserID)).
					Msg("Failed to update membership")

				return err
			}
		}
	}

	// Update pledge ended at.
	_, err = queries.CreateOrUpdatePatreonUser(ctx, database.CreateOrUpdatePatreonUserParams{
		PatreonUserID:    patreonUser.PatreonUserID,
		UserID:           patreonUser.UserID,
		PledgeCreatedAt:  patreonUser.PledgeCreatedAt,
		PledgeEndedAt:    time.Now(),
		TierID:           patreonUser.TierID,
		FullName:         utils.Coalesce(patreonMember.Attributes.FullName, patreonUser.FullName),
		Email:            utils.Coalesce(patreonMember.Attributes.Email, patreonUser.Email),
		ThumbUrl:         utils.Coalesce(patreonMember.Attributes.ThumbUrl, patreonUser.ThumbUrl),
		LastChargeStatus: string(patreonMember.Attributes.LastChargeStatus),
		PatronStatus:     string(patreonMember.Attributes.PatronStatus),
	})
	if err != nil {
		logger.Error().Err(err).
			Int64("user_id", int64(patreonUser.UserID)).
			Msg("Failed to create or update patreon user")

		return err
	}

	return nil
}

// Triggered when a patron is still pledging.
func OnPatreonActive(ctx context.Context, logger zerolog.Logger, queries *database.Queries, patreonUser database.PatreonUsers, patreonMember PatreonMember) error {
	println("HandlePatreonActive", patreonUser.PatreonUserID, patreonUser.UserID, patreonUser.TierID, patreonMember.Attributes.PatronStatus, patreonMember.Attributes.LastChargeStatus)

	memberships, err := queries.GetUserMembershipsByUserID(ctx, int64(patreonUser.UserID))
	if err != nil {
		logger.Error().Err(err).
			Int64("user_id", int64(patreonUser.UserID)).
			Msg("Failed to get user memberships")

		return err
	}

	var membershipExpiration time.Time
	if patreonMember.Attributes.LastChargeStatus == LastChargeStatusPaid {
		membershipExpiration = patreonMember.Attributes.LastChargeDate.AddDate(0, 1, 0)
	} else {
		membershipExpiration = time.Now().AddDate(0, 0, 7)
	}

	// If no longer pledging, ensure membership lasts a month after last purchase.
	for _, membership := range memberships {
		if database.MembershipType(membership.PlatformType.Int32) == database.MembershipType(database.PlatformTypePatreon) {
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
					Int64("user_id", int64(patreonUser.UserID)).
					Msg("Failed to update membership")

				return err
			}
		}
	}

	// Update patreon user
	_, err = queries.CreateOrUpdatePatreonUser(ctx, database.CreateOrUpdatePatreonUserParams{
		PatreonUserID:    patreonUser.PatreonUserID,
		UserID:           patreonUser.UserID,
		PledgeCreatedAt:  patreonUser.PledgeCreatedAt,
		PledgeEndedAt:    patreonUser.PledgeEndedAt,
		TierID:           patreonUser.TierID,
		FullName:         utils.Coalesce(patreonMember.Attributes.FullName, patreonUser.FullName),
		Email:            utils.Coalesce(patreonMember.Attributes.Email, patreonUser.Email),
		ThumbUrl:         utils.Coalesce(patreonMember.Attributes.ThumbUrl, patreonUser.ThumbUrl),
		LastChargeStatus: string(patreonMember.Attributes.LastChargeStatus),
		PatronStatus:     string(patreonMember.Attributes.PatronStatus),
	})
	if err != nil {
		logger.Error().Err(err).
			Int64("user_id", int64(patreonUser.UserID)).
			Msg("Failed to create or update patreon user")

		return err
	}

	return nil
}

// Triggers when a patreon account has been linked.
func OnPatreonLinked(ctx context.Context, logger zerolog.Logger, queries *database.Queries, patreonUser PatreonUser, automatic bool) error {
	println("HandlePatreonLinked", patreonUser.SocialConnections.Discord.UserID.String(), patreonUser.ID, automatic)

	// TODO: Notify of linked if not automatic or automatic and has tier.

	return nil
}

// Triggers when a patreon account has been unlinked.
func OnPatreonUnlinked(ctx context.Context, logger zerolog.Logger, queries *database.Queries, patreonUser *database.PatreonUsers) error {
	println("HandlePatreonUnlinked", patreonUser.PatreonUserID, patreonUser.UserID)

	// TODO: Figure out what to do. Let people know when it is unlinked?

	return nil
}

// Paypal Handlers

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
		if database.MembershipType(membership.PlatformType) == database.MembershipType(database.PlatformTypePaypalSubscription) &&
			(membership.Status == int32(database.MembershipStatusActive) || membership.Status == int32(database.MembershipStatusIdle)) &&
			membership.ExpiresAt.After(time.Now()) {
			paypalMemberships = append(paypalMemberships, membership)
		}
	}

	membershipExpiration := time.Now().AddDate(0, 1, 0)

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
		if tx.PlatformType == int32(database.PlatformTypePaypalSubscription) {
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
		if database.MembershipType(membership.PlatformType.Int32) == database.MembershipType(database.PlatformTypePaypalSubscription) &&
			membership.TransactionUuid == paypalTxs.TransactionUuid &&
			(membership.Status == int32(database.MembershipStatusActive) || membership.Status == int32(database.MembershipStatusIdle)) &&
			membership.ExpiresAt.After(time.Now()) {
			paypalMemberships = append(paypalMemberships, membership)
		}
	}

	var membershipExpiration time.Time
	if paypalSubscription.BillingInfo.FailedPaymentsCount == 0 {
		membershipExpiration = time.Now().AddDate(0, 1, 0)
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

// Meta Actions

// Triggers when a membership has been added to the server.
func onMembershipAdded(ctx context.Context, logger zerolog.Logger, queries *database.Queries, beforeMembership database.GetUserMembershipsByUserIDRow, membership database.UpdateUserMembershipParams) error {
	println("HandleMembershipAdded", beforeMembership.MembershipUuid.String(), beforeMembership.UserID, beforeMembership.GuildID, beforeMembership.ExpiresAt.String(), membership.GuildID, membership.ExpiresAt.String())

	// Notify of new membership?

	return nil
}

// Triggers when a membership has had its guild transferred.
func onMembershipUpdated(ctx context.Context, logger zerolog.Logger, queries *database.Queries, beforeMembership database.GetUserMembershipsByUserIDRow, membership database.UpdateUserMembershipParams) error {
	println("HandleMembershipUpdated", beforeMembership.MembershipUuid.String(), beforeMembership.UserID, beforeMembership.GuildID, beforeMembership.ExpiresAt.String(), membership.GuildID, membership.ExpiresAt.String())

	// Notify of transfer?

	return nil
}

// Triggers when a membership has been removed from the server.
func onMembershipRemoved(ctx context.Context, logger zerolog.Logger, queries *database.Queries, membership database.GetUserMembershipsByUserIDRow) error {
	println("HandleMembershipRemoved", membership.MembershipUuid.String(), membership.UserID, membership.GuildID, membership.ExpiresAt.String())

	// Notify of removal?

	return nil
}

// Triggers when a membership has expired and not renewed.
func onMembershipExpired(ctx context.Context, logger zerolog.Logger, queries *database.Queries, membership database.GetUserMembershipsByUserIDRow) error {
	// TODO: implement action

	println("HandleMembershipExpired", membership.MembershipUuid.String(), membership.UserID, membership.GuildID, membership.ExpiresAt.String())

	return nil
}

// Triggers when a membership has been renewed on expiry.
func onMembershipRenewed(ctx context.Context, logger zerolog.Logger, queries *database.Queries, membership database.GetUserMembershipsByUserIDRow) error {
	// TODO: implement action

	println("HandleMembershipRenewed", membership.MembershipUuid.String(), membership.UserID, membership.GuildID, membership.ExpiresAt.String())

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

		if database.MembershipStatus(membership.Status) == database.MembershipStatusIdle {
			err = onMembershipAdded(ctx, logger, queries, membership, newMembership)
			if err != nil {
				logger.Error().Err(err).
					Str("membership_uuid", membership.MembershipUuid.String()).
					Msg("Failed to trigger onMembershipAdded")
			}
		} else if database.MembershipStatus(membership.Status) == database.MembershipStatusActive {
			// Triggered when a membership has been transferred to a new guild.
			err = onMembershipUpdated(ctx, logger, queries, membership, newMembership)
			if err != nil {
				logger.Error().Err(err).
					Str("membership_uuid", membership.MembershipUuid.String()).
					Msg("Failed to trigger onMembershipUpdated")
			}
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

	err = onMembershipRemoved(ctx, logger, queries, membership)
	if err != nil {
		logger.Error().Err(err).
			Str("membership_uuid", membership.MembershipUuid.String()).
			Msg("Failed to trigger onMembershipRemoved")
	}

	return newMembership, nil
}
