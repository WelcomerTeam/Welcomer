package welcomer

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/WelcomerTeam/Discord/discord"
	sandwich_protobuf "github.com/WelcomerTeam/Sandwich-Daemon/proto"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
)

func NotifyMembershipCreated(ctx context.Context, session *discord.Session, membership database.UserMemberships) error {
	if membership.UserID == 0 {
		Logger.Error().Msg("Cannot notify membership created, user ID is 0")

		return nil
	}

	usersResponse, err := SandwichClient.FetchUser(ctx, &sandwich_protobuf.FetchUserRequest{
		UserIds: []int64{membership.UserID},
	})
	if err != nil {
		Logger.Error().Err(err).
			Int64("user_id", membership.UserID).
			Msg("Failed to fetch user")

		return err
	}

	users := usersResponse.GetUsers()

	if len(users) == 0 {
		Logger.Error().
			Int64("user_id", membership.UserID).
			Msg("Failed to fetch user, no users found")

		return nil
	}

	userPb := users[membership.UserID]
	user := sandwich_protobuf.PBToUser(userPb)

	transaction, err := Queries.GetUserTransaction(ctx, membership.TransactionUuid)
	if err != nil {
		Logger.Error().Err(err).
			Int64("user_id", int64(user.ID)).
			Str("transaction_uuid", membership.TransactionUuid.String()).
			Msg("Failed to get user transactions")

		return err
	}

	var descriptionBuilder strings.Builder

	if membership.MembershipType == int32(database.MembershipTypeCustomBackgrounds) {
		descriptionBuilder.WriteString("- This membership is for custom backgrounds, so you need to assign it to a server.\n")
	} else if membership.MembershipType == int32(database.MembershipTypeWelcomerPro) {
		descriptionBuilder.WriteString("- This membership is for Welcomer Pro so you need to make sure you assign it to a server before you can use the features. This also comes with custom backgrounds for the server.\n")
	}

	// Do not include for Discord payments, as they are automatically assigned to the server.
	if transaction.PlatformType != int32(database.PlatformTypeDiscord) {
		descriptionBuilder.WriteString("- If you have not assigned it to a server yet, you can run the `/membership add` command on the server you want to give it to, or use the **Memberships** tab on the website dashboard.\n")
	}

	descriptionBuilder.WriteString("- If you are having any issues, use the `/support` command to get access to our support server.\n")

	_, err = user.Send(ctx, session, discord.MessageParams{
		Content: fmt.Sprintf(
			"### Hey <@%d>, your **%s** %s via **%s** has now been created!\nThank you for supporting Welcomer, it means a lot to us.\n\n",
			user.ID,
			database.MembershipType(membership.MembershipType).Label(),
			If(membership.MembershipType == int32(database.MembershipTypeCustomBackgrounds), "benefit", "subscription"),
			database.PlatformType(transaction.PlatformType).Label(),
		),
		Embeds: []discord.Embed{
			{
				Description: descriptionBuilder.String(),
				Color:       EmbedColourInfo,
				Image: &discord.MediaItem{
					URL: WebsiteURL + "/assets/membership_thank_you.png",
				},
			},
		},
		Components: []discord.InteractionComponent{
			{
				Type: discord.InteractionComponentTypeActionRow,
				Components: []discord.InteractionComponent{
					{
						Type:  discord.InteractionComponentTypeButton,
						Style: discord.InteractionComponentStyleLink,
						Label: "Get Support",
						URL:   WebsiteURL + "/support",
					},
					{
						Type:  discord.InteractionComponentTypeButton,
						Style: discord.InteractionComponentStyleLink,
						Label: "Visit Dashboard",
						URL:   WebsiteURL + "/dashboard",
					},
				},
			},
		},
	})
	if err != nil {
		Logger.Error().Err(err).
			Int64("user_id", int64(user.ID)).
			Msg("Failed to send message")

		return err
	}

	Logger.Info().
		Int64("user_id", int64(user.ID)).
		Str("membership_id", membership.MembershipUuid.String()).
		Msg("Sent membership created message")

	return nil
}

func NotifyMembershipExpired(ctx context.Context, session *discord.Session, membership database.UserMemberships) error {
	if membership.UserID == 0 {
		Logger.Error().Msg("Cannot notify membership created, user ID is 0")

		return nil
	}

	usersResponse, err := SandwichClient.FetchUser(ctx, &sandwich_protobuf.FetchUserRequest{
		UserIds: []int64{membership.UserID},
	})
	if err != nil {
		Logger.Error().Err(err).
			Int64("user_id", membership.UserID).
			Msg("Failed to fetch user")

		return err
	}

	users := usersResponse.GetUsers()

	if len(users) == 0 {
		Logger.Error().
			Int64("user_id", membership.UserID).
			Msg("Failed to fetch user, no users found")

		return nil
	}

	userPb := users[membership.UserID]
	user := sandwich_protobuf.PBToUser(userPb)

	transaction, err := Queries.GetUserTransaction(ctx, membership.TransactionUuid)
	if err != nil {
		Logger.Error().Err(err).
			Int64("user_id", int64(user.ID)).
			Str("transaction_uuid", membership.TransactionUuid.String()).
			Msg("Failed to get user transactions")

		return err
	}

	var descriptionBuilder strings.Builder

	switch database.MembershipType(membership.MembershipType) {
	case database.MembershipTypeWelcomerPro:
		descriptionBuilder.WriteString("- This membership is for Welcomer Pro, so you will be missing out on any perks you got from it.\n")
	case database.MembershipTypeLegacyWelcomerPro:
		descriptionBuilder.WriteString("- This membership is a legacy membership for Welcomer Pro. You may have been receiving it for free, but it is now expired.\n")
		descriptionBuilder.WriteString("- This membership is for Welcomer Pro, so you will be missing out on any perks you got from it.\n")
	case database.MembershipTypeLegacyCustomBackgrounds:
		descriptionBuilder.WriteString("- This membership is a legacy membership for custom backgrounds. You may have been receiving it for free, but it is now expired.\n")
		descriptionBuilder.WriteString("- This membership is for custom backgrounds, so you will be missing out on any perks you got from it.\n")
	default:
		Logger.Warn().
			Str("membership_id", membership.MembershipUuid.String()).
			Int32("membership_type", membership.MembershipType).
			Int64("user_id", int64(user.ID)).
			Msg("Unexpected membership type is expiring")

		return nil
	}

	switch database.PlatformType(transaction.PlatformType) {
	case database.PlatformTypeDiscord:
		descriptionBuilder.WriteString("- This membership was purchased via Discord, so you will need to renew it below if you want to keep using it.\n")
	case database.PlatformTypePatreon:
		descriptionBuilder.WriteString("- This membership is a Patreon pledge, so you will need to renew it via Patreon or pledge again if you want to keep using it.\n")
	case database.PlatformTypePaypalSubscription:
		descriptionBuilder.WriteString("- This membership is a PayPal subscription, so you will need to renew it via PayPal or subscribe again if you want to keep using it.\n")
	case database.PlatformTypePaypal:
		descriptionBuilder.WriteString("- This membership was purchased via PayPal, so you will need to purchase a new plan if you want to keep using it.\n")
	default:
		Logger.Warn().
			Str("membership_id", membership.MembershipUuid.String()).
			Int32("platform_type", transaction.PlatformType).
			Int64("user_id", int64(user.ID)).
			Msg("Unexpected platform type is expiring")

		return nil
	}

	var content string

	if membership.ExpiresAt.Before(time.Now()) {
		content = fmt.Sprintf(
			"### Hey <@%d>, your **%s** %s via **%s** has expired!\n\n",
			user.ID,
			database.MembershipType(membership.MembershipType).Label(),
			If(membership.MembershipType == int32(database.MembershipTypeCustomBackgrounds), "benefit", "subscription"),
			database.PlatformType(transaction.PlatformType).Label(),
		)
	} else {
		content = fmt.Sprintf(
			"### Hey <@%d>, your **%s** %s via **%s** expires <t:%d:R>\nDon't miss out!\n\n",
			user.ID,
			database.MembershipType(membership.MembershipType).Label(),
			If(membership.MembershipType == int32(database.MembershipTypeCustomBackgrounds), "benefit", "subscription"),
			database.PlatformType(transaction.PlatformType).Label(),
			membership.ExpiresAt.Unix(),
		)
	}

	var guildID int64

	var guildName string

	if membership.GuildID != 0 {
		guild, err := FetchGuild(ctx, discord.Snowflake(membership.GuildID))
		if err != nil {
			Logger.Error().Err(err).
				Int64("guild_id", int64(membership.GuildID)).
				Msg("Failed to fetch guild from state cache")
		} else {
			guildName = guild.Name
		}
	}

	if guildName != "" {
		content += fmt.Sprintf("- This membership is assigned to the server **%s** `%d`.\n\n", guildName, guildID)
	}

	_, err = user.Send(ctx, session, discord.MessageParams{
		Content: content,
		Embeds: []discord.Embed{
			{
				Description: descriptionBuilder.String(),
				Color:       EmbedColourError,
			},
		},
		Components: []discord.InteractionComponent{
			{
				Type: discord.InteractionComponentTypeActionRow,
				Components: []discord.InteractionComponent{
					{
						Type:  discord.InteractionComponentTypeButton,
						Style: discord.InteractionComponentStylePremium,
						SKUID: discord.Snowflake(TryParseInt(os.Getenv("WELCOMER_PRO_DISCORD_SKU_ID"))),
					},
					{
						Type:  discord.InteractionComponentTypeButton,
						Style: discord.InteractionComponentStyleLink,
						Label: "Get Welcomer Pro",
						URL:   WebsiteURL + "/premium",
					},
				},
			},
			{
				Type: discord.InteractionComponentTypeActionRow,
				Components: []discord.InteractionComponent{
					{
						Type:  discord.InteractionComponentTypeButton,
						Style: discord.InteractionComponentStyleLink,
						Label: "Get Support",
						URL:   WebsiteURL + "/support",
					},
					{
						Type:  discord.InteractionComponentTypeButton,
						Style: discord.InteractionComponentStyleLink,
						Label: "Visit Dashboard",
						URL:   WebsiteURL + "/dashboard",
					},
				},
			},
		},
	})
	if err != nil {
		Logger.Error().Err(err).
			Int64("user_id", int64(user.ID)).
			Msg("Failed to send message")

		return err
	}

	Logger.Info().
		Int64("user_id", int64(user.ID)).
		Str("membership_id", membership.MembershipUuid.String()).
		Msg("Sent membership expiring message")

	return nil
}
