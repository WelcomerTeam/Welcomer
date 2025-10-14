package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"runtime/debug"
	"time"

	"github.com/WelcomerTeam/Discord/discord"
	sandwich_protobuf "github.com/WelcomerTeam/Sandwich-Daemon/proto"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	"github.com/jackc/pgx/v4"
	_ "github.com/joho/godotenv/autoload"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	var err error

	loggingLevel := flag.String("level", os.Getenv("LOGGING_LEVEL"), "Logging level")

	postgresURL := flag.String("postgresURL", os.Getenv("POSTGRES_URL"), "Postgres connection URL")
	sandwichGRPCHost := flag.String("sandwichGRPCHost", os.Getenv("SANDWICH_GRPC_HOST"), "GRPC Address for the Sandwich Daemon service")

	webhookUrl := flag.String("webhookUrl", os.Getenv("JOB_CLEANUP_EXPIRED_WELCOME_MESSAGES_WEBHOOK_URL"), "Webhook URL for logging")

	proxyAddress := flag.String("proxyAddress", os.Getenv("PROXY_ADDRESS"), "Address to proxy requests through. This can be 'https://discord.com', if one is not setup.")
	proxyDebug := flag.Bool("proxyDebug", false, "Enable debugging requests to the proxy")

	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from panic:", r)
			println(string(debug.Stack()))

			err = welcomer.SendWebhookMessage(ctx, *webhookUrl, discord.WebhookMessageParams{
				Content: "<@143090142360371200>",
				Embeds: []discord.Embed{
					{
						Title:       "Notify Expired Job",
						Description: fmt.Sprintf("Recovered from panic: %v", r),
						Color:       int32(16760839),
						Timestamp:   welcomer.ToPointer(time.Now()),
					},
				},
			})
			if err != nil {
				welcomer.Logger.Warn().Err(err).Msg("Failed to send webhook message")
			}
		}
	}()

	welcomer.SetupLogger(*loggingLevel)
	welcomer.SetupGRPCConnection(*sandwichGRPCHost,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(1024*1024*1024)), // Set max message size to 1GB
	)

	restInterface := welcomer.NewTwilightProxy(*proxyAddress)
	restInterface.SetDebug(*proxyDebug)
	welcomer.SetupRESTInterface(restInterface)

	welcomer.SetupSandwichClient()
	welcomer.SetupDatabase(ctx, *postgresURL)

	db, err := pgx.Connect(ctx, *postgresURL)
	if err != nil {
		panic(fmt.Sprintf(`pgx.Connect(%s): %v`, *postgresURL, err.Error()))
	}

	runPushGuildScience := welcomer.SetupPushGuildScience(1024)
	runPushGuildScience(ctx, time.Second*30)

	entrypoint(ctx, db)

	welcomer.PushGuildScience.Flush(ctx)

	cancel()
}

const FETCH_ENABLED_GUILDS = `-- name: FetchEnabledGuilds :many
SELECT
	guild_id,
	welcome_message_lifetime
FROM
	guild_settings_welcomer
WHERE
	auto_delete_welcome_messages = TRUE
	AND welcome_message_lifetime > 0`

type FetchEnabledGuildsRow struct {
	GuildID                int64 `json:"guild_id"`
	WelcomeMessageLifetime int32 `json:"welcome_message_lifetime"`
}

func entrypoint(ctx context.Context, db *pgx.Conn) {
	rows, err := db.Query(ctx, FETCH_ENABLED_GUILDS)
	if err != nil {
		panic(err)
	}

	defer rows.Close()

	var totalCountDeleted int

	for rows.Next() {
		var item FetchEnabledGuildsRow

		if err := rows.Scan(&item.GuildID, &item.WelcomeMessageLifetime); err != nil {
			panic(err)
		}

		welcomer.Logger.Info().Int64("guild_id", item.GuildID).
			Int32("welcome_message_lifetime", item.WelcomeMessageLifetime).
			Msg("Cleaning up expired welcome messages for guild")

		countDeleted, err := cleanupWelcomeMessagesForGuild(ctx, discord.Snowflake(item.GuildID), item.WelcomeMessageLifetime)
		if err != nil {
			welcomer.Logger.Error().Err(err).Int64("guild_id", item.GuildID).Msg("Failed to cleanup welcome messages for guild")
		}

		totalCountDeleted += countDeleted
	}

	welcomer.Logger.Info().
		Int("count_deleted", totalCountDeleted).
		Msg("Completed cleanup of expired welcome messages")
}

func cleanupWelcomeMessagesForGuild(ctx context.Context, guildID discord.Snowflake, welcomeMessageLifetime int32) (int, error) {
	var memberships []*database.GetUserMembershipsByGuildIDRow

	// Check if the guild has welcomer pro.
	memberships, err := welcomer.Queries.GetValidUserMembershipsByGuildID(ctx, guildID, time.Now())
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		welcomer.Logger.Warn().Err(err).
			Int64("guild_id", int64(guildID)).
			Msg("Failed to get welcomer memberships")

		return 0, err
	}

	hasWelcomerPro, _ := welcomer.CheckGuildMemberships(memberships)

	if !hasWelcomerPro {
		// Guild does not have welcomer pro, skip.
		return 0, nil
	}

	pbRoles, err := welcomer.SandwichClient.FetchGuildRole(ctx, &sandwich_protobuf.FetchGuildRoleRequest{
		GuildId: int64(guildID),
	})
	if err != nil {
		welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to fetch guild roles")

		return 0, fmt.Errorf("failed to fetch guild roles: %w", err)
	}

	locations, err := welcomer.SandwichClient.WhereIsGuild(ctx, &sandwich_protobuf.WhereIsGuildRequest{
		GuildId: int64(guildID),
	})
	if err != nil {
		return 0, fmt.Errorf("failed to find guild location: %w", err)
	}

	if len(locations.GetLocations()) == 0 {
		return 0, fmt.Errorf("no locations found for guild %d", guildID)
	}

	var botToken string

	for _, location := range locations.GetLocations() {

		guildMember := location.GetGuildMember()

		for _, roleID := range guildMember.Roles {
			for _, rolePb := range pbRoles.GetRoles() {
				role := sandwich_protobuf.PBToRole(rolePb)
				if role.ID == discord.Snowflake(roleID) {
					guildMember.Permissions |= int64(rolePb.Permissions)
				}
			}
		}

		if (guildMember.Permissions&discord.PermissionManageMessages != 0) ||
			(guildMember.Permissions&discord.PermissionAdministrator != 0) {
			applicationIdentifier := location.GetIdentifier()

			application, err := welcomer.SandwichClient.FetchApplication(ctx, &sandwich_protobuf.ApplicationIdentifier{
				ApplicationIdentifier: applicationIdentifier,
			})
			if err != nil {
				welcomer.Logger.Error().Err(err).Str("application_identifier", applicationIdentifier).
					Msg("Failed to fetch application for guild location")

				continue
			}

			if len(application.GetApplications()) == 0 {
				welcomer.Logger.Error().Str("application_identifier", applicationIdentifier).
					Msg("No application found for guild location")

				continue
			}

			for _, application := range application.GetApplications() {
				if application.GetBotToken() != "" {
					botToken = application.GetBotToken()
					break
				}
			}

		}
	}

	if botToken == "" {
		return 0, fmt.Errorf("no suitable bot token found for guild %d", guildID)
	}

	session := discord.NewSession("Bot "+botToken, welcomer.RESTInterface)

	messagesToDelete, err := welcomer.Queries.GetExpiredWelcomeMessageEvents(ctx, database.GetExpiredWelcomeMessageEventsParams{
		ScienceGuildEventTypeUserWelcomed:          int32(database.ScienceGuildEventTypeUserWelcomed),
		ScienceGuildEventTypeWelcomeMessageRemoved: int32(database.ScienceGuildEventTypeWelcomeMessageRemoved),
		GuildID:                int64(guildID),
		EventLimit:             50,
		WelcomeMessageLifetime: time.Now().Add(time.Second * time.Duration(-welcomeMessageLifetime)),
	})
	if err != nil {
		return 0, fmt.Errorf("failed to get expired welcome message events: %w", err)
	}

	channels := make(map[discord.Snowflake]bool)

	for _, message := range messagesToDelete {
		channels[discord.Snowflake(message.ChannelID)] = true
	}

	for channelID, _ := range channels {
		messageIDs := make([]discord.Snowflake, 0)
		messages := make([]database.GetExpiredWelcomeMessageEventsRow, 0)

		for _, message := range messagesToDelete {
			if discord.Snowflake(message.ChannelID) == channelID {
				messageIDs = append(messageIDs, discord.Snowflake(message.MessageID))
				messages = append(messages, *message)
			}
		}

		if len(messageIDs) > 2 {
			err = discord.BulkDeleteMessages(ctx, session, channelID, messageIDs, welcomer.ToPointer("Expired welcome message cleanup"))
			if err != nil {
				welcomer.Logger.Error().Err(err).Int64("guild_id", int64(guildID)).Int64("channel_id", int64(channelID)).
					Msg("Failed to bulk delete welcome messages for guild")
			}
		} else {
			message := discord.Message{ID: messageIDs[0], ChannelID: channelID}

			err = message.Delete(ctx, session, welcomer.ToPointer("Expired welcome message cleanup"))
			if err != nil {
				welcomer.Logger.Error().Err(err).Int64("guild_id", int64(guildID)).Int64("channel_id", int64(channelID)).
					Msg("Failed to delete welcome message for guild")
			}
		}

		for i, messageID := range messageIDs {
			// Push WelcomeMessageRemoved event to the buffer.
			welcomer.PushGuildScience.Push(
				ctx,
				guildID,
				discord.Snowflake(messages[i].UserID.Int64),
				database.ScienceGuildEventTypeWelcomeMessageRemoved,
				welcomer.GuildScienceWelcomeMessageRemoved{
					MessageID:        messageID,
					MessageChannelID: channelID,
					HasMessage:       true,
					Successful:       err == nil,
				},
			)
		}
	}

	return len(messagesToDelete), nil
}
