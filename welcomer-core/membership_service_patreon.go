package welcomer

import (
	"context"
	"time"

	"github.com/WelcomerTeam/Discord/discord"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
)

// Triggers when a patreon tier has changed.
func OnPatreonTierChanged(ctx context.Context, beforePatreonUser *database.PatreonUsers, patreonUser database.CreateOrUpdatePatreonUserParams) error {
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

	txs, err := Queries.GetUserTransactionsByUserID(ctx, int64(patreonUser.UserID))
	if err != nil {
		Logger.Error().Err(err).
			Int64("user_id", int64(patreonUser.UserID)).
			Msg("Failed to get user transactions")

		return err
	}

	var patreonTransaction *database.UserTransactions

	for _, tx := range txs {
		if tx.PlatformType == int32(database.PlatformTypePatreon) {
			patreonTransaction = tx

			break
		}
	}

	// If no patreon transaction is found, create one.
	if patreonTransaction == nil {
		Logger.Warn().
			Int64("user_id", int64(patreonUser.UserID)).
			Msg("No patreon transaction found")

		transaction, err := CreateTransactionForUser(ctx, discord.Snowflake(patreonUser.UserID), database.PlatformTypePatreon, database.TransactionStatusCompleted, "", "", "")
		if err != nil {
			Logger.Error().Err(err).
				Int64("user_id", int64(patreonUser.UserID)).
				Msg("Failed to create patreon transaction")

			return err
		}

		patreonTransaction = transaction
	}

	// Update pledge status
	if patreonUser.TierID == 0 {
		patreonUser.PledgeEndedAt = time.Time{}
	} else {
		patreonUser.PledgeCreatedAt = time.Now()
		patreonUser.PledgeEndedAt = time.Time{}
	}

	_, err = Queries.CreateOrUpdatePatreonUser(ctx, database.CreateOrUpdatePatreonUserParams{
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
		Logger.Error().Err(err).
			Int64("user_id", int64(patreonUser.UserID)).
			Msg("Failed to create or update patreon user")

		return err
	}

	memberships, err := Queries.GetUserMembershipsByUserID(ctx, int64(patreonUser.UserID))
	if err != nil {
		Logger.Error().Err(err).
			Int64("user_id", int64(patreonUser.UserID)).
			Msg("Failed to get user memberships")

		return err
	}

	patreonMemberships := make([]*database.GetUserMembershipsByUserIDRow, 0, len(memberships))

	for _, membership := range memberships {
		if database.PlatformType(membership.PlatformType.Int32) == database.PlatformTypePatreon &&
			membership.TransactionUuid == patreonTransaction.TransactionUuid &&
			(membership.Status == int32(database.MembershipStatusActive) || membership.Status == int32(database.MembershipStatusIdle)) &&
			membership.ExpiresAt.After(time.Now()) {
			patreonMemberships = append(patreonMemberships, membership)
		}
	}

	var membershipExpiration time.Time
	if LastChargeStatus(patreonUser.LastChargeStatus) == LastChargeStatusPaid {
		membershipExpiration = time.Now().AddDate(0, 1, 7)
	} else {
		membershipExpiration = time.Now().AddDate(0, 0, 7)
	}

	if len(patreonMemberships) > eligibleMemberships {
		Logger.Warn().
			Int64("user_id", int64(patreonUser.UserID)).
			Int("current_memberships", len(patreonMemberships)).
			Int("eligible_memberships", eligibleMemberships).
			Msg("User has too many memberships")

		// Remove memberships. Let them expire naturally.
	} else if len(patreonMemberships) < eligibleMemberships {
		Logger.Info().
			Int64("user_id", int64(patreonUser.UserID)).
			Int("current_memberships", len(patreonMemberships)).
			Int("eligible_memberships", eligibleMemberships).
			Msg("User has too few memberships")

		// Add memberships

		for i := len(patreonMemberships); i < eligibleMemberships; i++ {
			err = CreateMembershipForUser(ctx, discord.Snowflake(patreonUser.UserID), patreonTransaction.TransactionUuid, membershipType, membershipExpiration, nil)
			if err != nil {
				Logger.Error().Err(err).
					Int64("user_id", int64(patreonUser.UserID)).
					Msg("Failed to create membership")

				return err
			}
		}

		for _, membership := range patreonMemberships {
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
					Int64("user_id", int64(patreonUser.UserID)).
					Msg("Failed to update membership")

				return err
			}
		}
	} else {
		Logger.Info().
			Int64("user_id", int64(patreonUser.UserID)).
			Int("current_memberships", len(patreonMemberships)).
			Int("eligible_memberships", eligibleMemberships).
			Msg("User has correct number of memberships")

		// No changes
	}

	return nil
}

// Triggers when a patreon tier has changed. Ran if OnPatreonTierChange failed to run.
func OnPatreonTierChanged_Fallback(ctx context.Context, beforePatreonUser *database.PatreonUsers, patreonUser database.CreateOrUpdatePatreonUserParams, e error) error {
	println("HandlePatreonTierChanged_Fallback", beforePatreonUser.PatreonUserID, beforePatreonUser.TierID, patreonUser.TierID, e.Error())

	// TODO: Log to file/webhook

	return nil
}

// Triggers when a patron is no longer pledging
func OnPatreonNoLongerPledging(ctx context.Context, patreonUser database.PatreonUsers, patreonMember PatreonMember) error {
	println("HandlePatreonNoLongerPledging", patreonUser.PatreonUserID, patreonUser.UserID, patreonUser.TierID, patreonMember.Attributes.PatronStatus)

	// Update pledge ended at.
	_, err := Queries.CreateOrUpdatePatreonUser(ctx, database.CreateOrUpdatePatreonUserParams{
		PatreonUserID:    patreonUser.PatreonUserID,
		UserID:           patreonUser.UserID,
		PledgeCreatedAt:  patreonUser.PledgeCreatedAt,
		PledgeEndedAt:    time.Now(),
		TierID:           patreonUser.TierID,
		FullName:         Coalesce(patreonMember.Attributes.FullName, patreonUser.FullName),
		Email:            Coalesce(patreonMember.Attributes.Email, patreonUser.Email),
		ThumbUrl:         Coalesce(patreonMember.Attributes.ThumbUrl, patreonUser.ThumbUrl),
		LastChargeStatus: string(patreonMember.Attributes.LastChargeStatus),
		PatronStatus:     string(patreonMember.Attributes.PatronStatus),
	})
	if err != nil {
		Logger.Error().Err(err).
			Int64("user_id", int64(patreonUser.UserID)).
			Msg("Failed to create or update patreon user")

		return err
	}

	memberships, err := Queries.GetUserMembershipsByUserID(ctx, int64(patreonUser.UserID))
	if err != nil {
		Logger.Error().Err(err).
			Int64("user_id", int64(patreonUser.UserID)).
			Msg("Failed to get user memberships")

		return err
	}

	var membershipExpiration time.Time
	if patreonMember.Attributes.LastChargeStatus == LastChargeStatusPaid {
		membershipExpiration = patreonMember.Attributes.LastChargeDate.AddDate(0, 1, 7)
	} else {
		membershipExpiration = time.Now().AddDate(0, 0, 7)
	}

	// If no longer pledging, ensure membership lasts a month after last purchase.
	for _, membership := range memberships {
		if database.MembershipType(membership.PlatformType.Int32) == database.MembershipType(database.PlatformTypePatreon) && (membership.Status == int32(database.MembershipStatusActive) || membership.Status == int32(database.MembershipStatusIdle)) && membership.ExpiresAt.After(time.Now()) {
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
					Int64("user_id", int64(patreonUser.UserID)).
					Msg("Failed to update membership")

				return err
			}
		}
	}

	return nil
}

// Triggered when a patron is still pledging.
func OnPatreonActive(ctx context.Context, patreonUser database.PatreonUsers, patreonMember PatreonMember) error {
	println("HandlePatreonActive", patreonUser.PatreonUserID, patreonUser.UserID, patreonUser.TierID, patreonMember.Attributes.PatronStatus, patreonMember.Attributes.LastChargeStatus)

	// Update patreon user
	_, err := Queries.CreateOrUpdatePatreonUser(ctx, database.CreateOrUpdatePatreonUserParams{
		PatreonUserID:    patreonUser.PatreonUserID,
		UserID:           patreonUser.UserID,
		PledgeCreatedAt:  patreonUser.PledgeCreatedAt,
		PledgeEndedAt:    patreonUser.PledgeEndedAt,
		TierID:           patreonUser.TierID,
		FullName:         Coalesce(patreonMember.Attributes.FullName, patreonUser.FullName),
		Email:            Coalesce(patreonMember.Attributes.Email, patreonUser.Email),
		ThumbUrl:         Coalesce(patreonMember.Attributes.ThumbUrl, patreonUser.ThumbUrl),
		LastChargeStatus: string(patreonMember.Attributes.LastChargeStatus),
		PatronStatus:     string(patreonMember.Attributes.PatronStatus),
	})
	if err != nil {
		Logger.Error().Err(err).
			Int64("user_id", int64(patreonUser.UserID)).
			Msg("Failed to create or update patreon user")

		return err
	}

	memberships, err := Queries.GetUserMembershipsByUserID(ctx, int64(patreonUser.UserID))
	if err != nil {
		Logger.Error().Err(err).
			Int64("user_id", int64(patreonUser.UserID)).
			Msg("Failed to get user memberships")

		return err
	}

	var membershipExpiration time.Time
	if patreonMember.Attributes.LastChargeStatus == LastChargeStatusPaid {
		membershipExpiration = patreonMember.Attributes.LastChargeDate.AddDate(0, 1, 7)
	} else {
		membershipExpiration = time.Now().AddDate(0, 0, 7)
	}

	// If no longer pledging, ensure membership lasts a month after last purchase.
	for _, membership := range memberships {
		if database.MembershipType(membership.PlatformType.Int32) == database.MembershipType(database.PlatformTypePatreon) {
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
					Int64("user_id", int64(patreonUser.UserID)).
					Msg("Failed to update membership")

				return err
			}
		}
	}

	return nil
}

// Triggers when a patreon account has been linked.
func OnPatreonLinked(ctx context.Context, patreonUser PatreonUser, automatic bool) error {
	println("HandlePatreonLinked", patreonUser.SocialConnections.Discord.UserID.String(), patreonUser.ID, automatic)

	// TODO: Notify of linked if not automatic or automatic and has tier.

	return nil
}

// Triggers when a patreon account has been unlinked.
func OnPatreonUnlinked(ctx context.Context, patreonUser *database.PatreonUsers) error {
	println("HandlePatreonUnlinked", patreonUser.PatreonUserID, patreonUser.UserID)

	// TODO: Figure out what to do. Let people know when it is unlinked?

	return nil
}
