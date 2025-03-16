package welcomer

import (
	"context"
	"database/sql"
	"os"
	"time"

	"github.com/WelcomerTeam/Discord/discord"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	utils "github.com/WelcomerTeam/Welcomer/welcomer-utils"
	"github.com/rs/zerolog"
)

var (
	WelcomerProDiscordSKU               = discord.Snowflake(utils.TryParseInt(os.Getenv("WELCOMER_PRO_DISCORD_SKU_ID")))
	WelcomerCustomBackgroundsDiscordSKU = discord.Snowflake(utils.TryParseInt(os.Getenv("WELCOMER_CUSTOM_BACKGROUNDS_DISCORD_SKU_ID")))
)

func returnSnowflakeIfNotNull(i *discord.Snowflake) discord.Snowflake {
	if i == nil {
		return discord.Snowflake(0)
	}

	return *i
}

func HandleDiscordEntitlement(ctx context.Context, logger zerolog.Logger, queries *database.Queries, entitlement discord.Entitlement) error {
	if entitlement.UserID == nil {
		logger.Error().
			Int64("entitlement_id", int64(entitlement.ID)).
			Msg("entitlement user_id is nil")

		return nil
	}

	giftCodeFlags := sql.NullInt64{Valid: false}
	if entitlement.GiftCodeFlags != nil {
		giftCodeFlags = sql.NullInt64{Valid: true, Int64: int64(*entitlement.GiftCodeFlags)}
	}

	guildID := sql.NullInt64{Valid: false}
	if entitlement.GuildID != nil {
		guildID = sql.NullInt64{Valid: true, Int64: int64(*entitlement.GuildID)}
	}

	startsAt := sql.NullTime{Valid: false}
	if entitlement.StartsAt != nil {
		startsAt = sql.NullTime{Valid: true, Time: *entitlement.StartsAt}
	} else {
		logger.Warn().
			Int64("entitlement_id", int64(entitlement.ID)).
			Msg("entitlement starts_at is nil")

		startsAt = sql.NullTime{Valid: true, Time: time.Now()}
	}

	endsAt := sql.NullTime{Valid: false}
	if entitlement.EndsAt != nil {
		endsAt = sql.NullTime{Valid: true, Time: *entitlement.EndsAt}
	} else {
		endsAt = sql.NullTime{Valid: true, Time: startsAt.Time.AddDate(0, 1, 7)}
	}

	println("HandleDiscordEntitlement", entitlement.ID, entitlement.SkuID, startsAt.Time.String(), endsAt.Time.String(), returnSnowflakeIfNotNull(entitlement.UserID).String(), returnSnowflakeIfNotNull(entitlement.GuildID).String(), entitlement.Consumed, entitlement.Deleted)

	var membershipType database.MembershipType

	switch entitlement.SkuID {
	case WelcomerProDiscordSKU:
		membershipType = database.MembershipTypeWelcomerPro
	case WelcomerCustomBackgroundsDiscordSKU:
		membershipType = database.MembershipTypeCustomBackgrounds
	default:
		logger.Error().
			Int64("entitlement_id", int64(entitlement.ID)).
			Int64("sku_id", int64(entitlement.SkuID)).
			Msg("Unknown SKU")

		return nil
	}

	var quantity int = 1

	if endsAt.Time.Before(time.Now()) && !endsAt.Time.IsZero() {
		quantity = 0
	}

	txs, err := queries.GetUserTransactionsByUserID(ctx, int64(*entitlement.UserID))
	if err != nil {
		logger.Error().Err(err).
			Int64("user_id", int64(*entitlement.UserID)).
			Msg("failed to get user transactions")

		return err
	}

	var discordTxs *database.UserTransactions

	for _, tx := range txs {
		if tx.PlatformType == int32(database.PlatformTypeDiscord) &&
			tx.TransactionID == entitlement.ID.String() {
			discordTxs = tx

			break
		}
	}

	// If no discord transactions exist, create one.
	if discordTxs == nil {
		logger.Info().
			Str("entitlement_id", entitlement.ID.String()).
			Msg("no discord transactions found, creating one")

		discordTxs, err = CreateTransactionForUser(ctx, queries, *entitlement.UserID, database.PlatformTypeDiscord, database.TransactionStatusCompleted, entitlement.ID.String(), "", "")
		if err != nil {
			logger.Error().Err(err).
				Str("entitlement_id", entitlement.ID.String()).
				Msg("failed to create discord transaction")

			return err
		}
	}

	// Update discord subscription entity.
	_, err = queries.CreateOrUpdateDiscordSubscription(ctx, database.CreateOrUpdateDiscordSubscriptionParams{
		SubscriptionID:  entitlement.ID.String(),
		UserID:          int64(*entitlement.UserID),
		GiftCodeFlags:   giftCodeFlags,
		GuildID:         guildID,
		StartsAt:        startsAt,
		EndsAt:          endsAt,
		SkuID:           int64(entitlement.SkuID),
		ApplicationID:   int64(entitlement.ApplicationID),
		EntitlementType: int64(entitlement.Type),
		Deleted:         entitlement.Deleted,
		Consumed:        entitlement.Consumed,
	})
	if err != nil {
		logger.Error().Err(err).
			Str("entitlement_id", entitlement.ID.String()).
			Msg("failed to create or update discord subscription")

		return err
	}

	memberships, err := queries.GetUserMembershipsByTransactionID(ctx, entitlement.ID.String())
	if err != nil {
		logger.Error().Err(err).
			Int64("user_id", int64(*entitlement.UserID)).
			Msg("failed to get user memberships")

		return err
	}

	discordMemberships := make([]*database.GetUserMembershipsByTransactionIDRow, 0, len(memberships))

	for _, membership := range memberships {
		if database.PlatformType(membership.PlatformType) == database.PlatformTypeDiscord &&
			membership.TransactionID == entitlement.ID.String() &&
			(membership.Status == int32(database.MembershipStatusActive) || membership.Status == int32(database.MembershipStatusIdle)) &&
			membership.ExpiresAt.After(time.Now()) {
			discordMemberships = append(discordMemberships, membership)
		}
	}

	if len(discordMemberships) > quantity {
		logger.Warn().
			Int64("user_id", int64(*entitlement.UserID)).
			Int("discord_memberships", len(discordMemberships)).
			Int("eligible_memberships", quantity).
			Msg("User has too many memberships")

		// Remove memberships. Let them expire naturally.
	} else if len(discordMemberships) < quantity {
		logger.Info().
			Int64("user_id", int64(*entitlement.UserID)).
			Int("current_memberships", len(discordMemberships)).
			Int("eligible_memberships", quantity).
			Msg("User has too few memberships")

		// Add memberships.

		for i := len(discordMemberships); i < quantity; i++ {
			err = CreateMembershipForUser(ctx, queries, *entitlement.UserID, discordTxs.TransactionUuid, membershipType, endsAt.Time, entitlement.GuildID)
			if err != nil {
				logger.Error().Err(err).
					Int64("user_id", int64(*entitlement.UserID)).
					Msg("failed to create membership")

				return err
			}
		}

		for _, membership := range discordMemberships {
			membership.ExpiresAt = endsAt.Time

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
					Int64("user_id", int64(*entitlement.UserID)).
					Msg("Failed to update membership")

				return err
			}
		}
	} else {
		logger.Info().
			Int64("user_id", int64(*entitlement.UserID)).
			Int("current_memberships", len(discordMemberships)).
			Int("eligible_memberships", quantity).
			Msg("User has correct number of memberships")

		// No changes
	}

	return nil
}

func OnDiscordEntitlementCreated(ctx context.Context, logger zerolog.Logger, queries *database.Queries, entitlement discord.Entitlement) error {
	return HandleDiscordEntitlement(ctx, logger, queries, entitlement)
}

func OnDiscordEntitlementUpdated(ctx context.Context, logger zerolog.Logger, queries *database.Queries, entitlement discord.Entitlement) error {
	return HandleDiscordEntitlement(ctx, logger, queries, entitlement)
}

func OnDiscordEntitlementDeleted(ctx context.Context, logger zerolog.Logger, queries *database.Queries, entitlement discord.Entitlement) error {
	return HandleDiscordEntitlement(ctx, logger, queries, entitlement)
}
