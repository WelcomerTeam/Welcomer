package welcomer

import (
	"context"
	"fmt"
	"os"

	"github.com/WelcomerTeam/Discord/discord"
	sandwich "github.com/WelcomerTeam/Sandwich-Daemon/protobuf"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
)

func notifyMembershipCreated(ctx context.Context, session *discord.Session, membership database.UserMemberships) error {
	if membership.UserID == 0 {
		Logger.Error().Msg("Cannot notify membership created, user ID is 0")

		return nil
	}

	users, err := SandwichClient.FetchUsers(ctx, &sandwich.FetchUsersRequest{
		UserIDs:         []int64{membership.UserID},
		CreateDMChannel: true,
	})
	if err != nil {
		Logger.Error().Err(err).
			Int64("user_id", membership.UserID).
			Msg("Failed to fetch user")

		return err
	}

	if len(users.Users) == 0 {
		Logger.Error().
			Int64("user_id", membership.UserID).
			Msg("Failed to fetch user, no users found")

		return nil
	}

	pbUser := users.Users[0]

	user, err := sandwich.GRPCToUser(pbUser)
	if err != nil {
		Logger.Error().Err(err).
			Int64("user_id", pbUser.ID).
			Msg("Failed to convert user")

		return err
	}

	transactions, err := Queries.GetUserTransactionsByTransactionID(ctx, membership.TransactionUuid.String())
	if err != nil {
		Logger.Error().Err(err).
			Int64("user_id", int64(user.ID)).
			Str("transaction_uuid", membership.TransactionUuid.String()).
			Msg("Failed to get user transactions")

		return err
	}

	transaction := transactions[0]

	description := ""

	if membership.MembershipType == int32(database.MembershipTypeCustomBackgrounds) {
		description += "This membership is a custom backgrounds membership so you need to assign it to a server.\n\n"
	} else if membership.MembershipType == int32(database.MembershipTypeWelcomerPro) {
		description += "This membership is for Welcomer Pro so you need to make sure you assign it to a server before you can use it's features. This comes with custom backgrounds for the server.\n\n"
	}

	// Do not include for Discord payments, as they are automatically assigned to the server.
	if transaction.PlatformType != int32(database.PlatformTypeDiscord) {
		description += "If you have not assigned it to a server yet, you can run the `/membership add` command on the server you want to give it to, or use the **Memberships** tab on the website dashboard.\n\n"
	}

	description += "If you are having any issues, use the `/support` command to get access to our support server.\n\n"

	_, err = user.Send(ctx, session, discord.MessageParams{
		Content: fmt.Sprintf(
			"## Hey <@%d>, your **%s** subscription via **%s** has now been created!\nThank you for supporting Welcomer, it means a lot to us.",
			user.ID,
			database.MembershipType(membership.MembershipType).Label(),
			database.PlatformType(transaction.PlatformType).Label(),
		),
		Embeds: []discord.Embed{
			{
				Description: description,
				Color:       0x4CD787,
				Image: &discord.EmbedImage{
					URL: "https://" + os.Getenv("DOMAIN") + "/assets/membership_thank_you.png",
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

func notifyMembershipExpired(ctx context.Context, session *discord.Session, membership database.UserMemberships) error {
	if membership.UserID == 0 {
		Logger.Error().Msg("Cannot notify membership created, user ID is 0")

		return nil
	}

	return nil
}
