package welcomer

import (
	"context"
	"database/sql"
	"os"
	"time"

	"github.com/WelcomerTeam/Discord/discord"
	pb "github.com/WelcomerTeam/Sandwich-Daemon/protobuf"
	sandwich "github.com/WelcomerTeam/Sandwich-Daemon/protobuf"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
)

var (
	WelcomerProDiscordSKU               = discord.Snowflake(TryParseInt(os.Getenv("WELCOMER_PRO_DISCORD_SKU_ID")))
	WelcomerCustomBackgroundsDiscordSKU = discord.Snowflake(TryParseInt(os.Getenv("WELCOMER_CUSTOM_BACKGROUNDS_DISCORD_SKU_ID")))
)

func getSKUNameFromID(skuID discord.Snowflake) string {
	switch skuID {
	case WelcomerProDiscordSKU:
		return "Welcomer Pro"
	case WelcomerCustomBackgroundsDiscordSKU:
		return "Welcomer Custom Backgrounds"
	default:
		return "Unknown"
	}
}

func returnSnowflakeIfNotNull(i *discord.Snowflake) discord.Snowflake {
	if i == nil {
		return discord.Snowflake(0)
	}

	return *i
}

func FetchGuildName(ctx context.Context, guildID discord.Snowflake) string {
	guilds, err := SandwichClient.FetchGuild(ctx, &sandwich.FetchGuildRequest{
		GuildIDs: []int64{int64(guildID)},
	})
	if err != nil {
		Logger.Warn().Err(err).
			Int64("guild_id", int64(guildID)).
			Msg("Failed to fetch guild")

		return ""
	}

	var guild discord.Guild

	guildPb, ok := guilds.GetGuilds()[int64(guildID)]

	if !ok {
		return ""
	}

	guild, err = pb.GRPCToGuild(guildPb)
	if err != nil {
		Logger.Warn().Err(err).
			Int64("guild_id", int64(guildID)).
			Msg("Failed to convert guild")

		return ""
	}

	return guild.Name
}

func HandleDiscordEntitlement(ctx context.Context, entitlement discord.Entitlement) error {
	if entitlement.UserID == nil {
		Logger.Error().
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
		Logger.Warn().
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

	var guildName string

	if guildID.Valid {
		guildName = FetchGuildName(ctx, discord.Snowflake(guildID.Int64))
	}

	if guildName == "" {
		guildName = "Unknown"
	}

	if DiscordPaymentsWebhookURL != "" {
		err := SendWebhookMessage(ctx, DiscordPaymentsWebhookURL, discord.WebhookMessageParams{
			Embeds: []discord.Embed{
				{
					Title:     "Discord Sale",
					Color:     0x5865F2,
					Timestamp: ToPointer(time.Now()),
					Fields: []discord.EmbedField{
						{
							Name:   "ID",
							Value:  entitlement.ID.String(),
							Inline: true,
						},
						{
							Name:   "User",
							Value:  "<@" + entitlement.UserID.String() + "> " + entitlement.UserID.String(),
							Inline: true,
						},
						{
							Name:  "Application",
							Value: entitlement.ApplicationID.String(),
						},
						{
							Name:  "SKU",
							Value: getSKUNameFromID(entitlement.SkuID),
						},
						{
							Name:  "Starts At",
							Value: "<t:" + Itoa(startsAt.Time.Unix()) + ":R>",
						},
						{
							Name:  "Ends At",
							Value: "<t:" + Itoa(endsAt.Time.Unix()) + ":R>",
						},
						{
							Name:  "Gift Code Flags",
							Value: Itoa(giftCodeFlags.Int64),
						},
						{
							Name:  "Guild",
							Value: If(guildID.Valid, "`"+guildName+"` ", "") + Itoa(guildID.Int64),
						},
						{
							Name:  "Consumed",
							Value: If(entitlement.Consumed, "TRUE", "FALSE"),
						},
						{
							Name:  "Deleted",
							Value: If(entitlement.Deleted, "TRUE", "FALSE"),
						},
					},
				},
			},
		})
		if err != nil {
			Logger.Error().Err(err).Msg("Failed to send webhook message")
		}
	}

	var membershipType database.MembershipType

	switch entitlement.SkuID {
	case WelcomerProDiscordSKU:
		membershipType = database.MembershipTypeWelcomerPro
	case WelcomerCustomBackgroundsDiscordSKU:
		membershipType = database.MembershipTypeCustomBackgrounds
	default:
		Logger.Error().
			Int64("entitlement_id", int64(entitlement.ID)).
			Int64("sku_id", int64(entitlement.SkuID)).
			Msg("Unknown SKU")

		return nil
	}

	var quantity int = 1

	if endsAt.Time.Before(time.Now()) && !endsAt.Time.IsZero() {
		quantity = 0
	}

	txs, err := Queries.GetUserTransactionsByUserID(ctx, int64(*entitlement.UserID))
	if err != nil {
		Logger.Error().Err(err).
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
		Logger.Info().
			Str("entitlement_id", entitlement.ID.String()).
			Msg("no discord transactions found, creating one")

		discordTxs, err = CreateTransactionForUser(ctx, *entitlement.UserID, database.PlatformTypeDiscord, database.TransactionStatusCompleted, entitlement.ID.String(), "", "")
		if err != nil {
			Logger.Error().Err(err).
				Str("entitlement_id", entitlement.ID.String()).
				Msg("failed to create discord transaction")

			return err
		}
	}

	// Update discord subscription entity.
	_, err = Queries.CreateOrUpdateDiscordSubscription(ctx, database.CreateOrUpdateDiscordSubscriptionParams{
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
		Logger.Error().Err(err).
			Str("entitlement_id", entitlement.ID.String()).
			Msg("failed to create or update discord subscription")

		return err
	}

	memberships, err := Queries.GetUserMembershipsByTransactionID(ctx, entitlement.ID.String())
	if err != nil {
		Logger.Error().Err(err).
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
		Logger.Warn().
			Int64("user_id", int64(*entitlement.UserID)).
			Int("discord_memberships", len(discordMemberships)).
			Int("eligible_memberships", quantity).
			Msg("User has too many memberships")

		// Remove memberships. Let them expire naturally.
	} else if len(discordMemberships) < quantity {
		Logger.Info().
			Int64("user_id", int64(*entitlement.UserID)).
			Int("current_memberships", len(discordMemberships)).
			Int("eligible_memberships", quantity).
			Msg("User has too few memberships")

		// Add memberships.

		for i := len(discordMemberships); i < quantity; i++ {
			err = CreateMembershipForUser(ctx, *entitlement.UserID, discordTxs.TransactionUuid, membershipType, endsAt.Time, entitlement.GuildID)
			if err != nil {
				Logger.Error().Err(err).
					Int64("user_id", int64(*entitlement.UserID)).
					Msg("failed to create membership")

				return err
			}
		}

		for _, membership := range discordMemberships {
			membership.ExpiresAt = endsAt.Time
			membership.Status = If(
				membership.Status == int32(database.MembershipStatusIdle),
				int32(database.MembershipStatusActive),
				membership.Status)

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
					Int64("user_id", int64(*entitlement.UserID)).
					Msg("Failed to update membership")

				return err
			}
		}
	} else {
		Logger.Info().
			Int64("user_id", int64(*entitlement.UserID)).
			Int("current_memberships", len(discordMemberships)).
			Int("eligible_memberships", quantity).
			Msg("User has correct number of memberships")

		// No changes
	}

	return nil
}

func OnDiscordEntitlementCreated(ctx context.Context, entitlement discord.Entitlement) error {
	return HandleDiscordEntitlement(ctx, entitlement)
}

func OnDiscordEntitlementUpdated(ctx context.Context, entitlement discord.Entitlement) error {
	return HandleDiscordEntitlement(ctx, entitlement)
}

func OnDiscordEntitlementDeleted(ctx context.Context, entitlement discord.Entitlement) error {
	return HandleDiscordEntitlement(ctx, entitlement)
}
