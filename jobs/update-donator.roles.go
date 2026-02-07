package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"runtime/debug"
	"time"

	"github.com/WelcomerTeam/Discord/discord"
	sandwich "github.com/WelcomerTeam/Sandwich-Daemon/proto"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	_ "github.com/joho/godotenv/autoload"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	WelcomerGuild   discord.Snowflake = 341685098468343822
	DonatorRole     discord.Snowflake = 460443036825157633
	WelcomerProRole discord.Snowflake = 506219152097148938
)

func main() {
	var err error

	loggingLevel := flag.String("level", os.Getenv("LOGGING_LEVEL"), "Logging level")

	postgresURL := flag.String("postgresURL", os.Getenv("POSTGRES_URL"), "Postgres connection URL")
	sandwichGRPCHost := flag.String("sandwichGRPCHost", os.Getenv("SANDWICH_GRPC_HOST"), "GRPC Address for the Sandwich Daemon service")

	proxyAddress := flag.String("proxyAddress", os.Getenv("PROXY_ADDRESS"), "Address to proxy requests through. This can be 'https://discord.com', if one is not setup.")
	proxyDebug := flag.Bool("proxyDebug", false, "Enable debugging requests to the proxy")

	webhookUrl := flag.String("webhookUrl", os.Getenv("JOB_CLEANUP_CUSTOM_BOTS_WEBHOOK_URL"), "Webhook URL for logging")

	sandwichManagerName := flag.String("sandwichManagerName", os.Getenv("SANDWICH_MANAGER_NAME"), "Sandwich manager identifier name")

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
						Title:       "Update Donator Roles Job",
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

	restInterface := welcomer.NewTwilightProxy(*proxyAddress)
	restInterface.SetDebug(*proxyDebug)

	welcomer.SetupDefaultManagerName(*sandwichManagerName)
	welcomer.SetupLogger(*loggingLevel)
	welcomer.SetupGRPCConnection(*sandwichGRPCHost,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(1024*1024*1024)), // Set max message size to 1GB
	)
	welcomer.SetupRESTInterface(restInterface)
	welcomer.SetupSandwichClient()
	welcomer.SetupDatabase(ctx, *postgresURL)

	entrypoint(ctx, *webhookUrl)

	if err := welcomer.Queries.UpsertJobCheckpoint(ctx, database.UpsertJobCheckpointParams{
		JobName:         "update-donator-roles",
		LastProcessedTs: time.Now().UTC(),
	}); err != nil {
		welcomer.Logger.Error().Err(err).Msg("Failed to upsert job checkpoint")
	}

	cancel()
}

func entrypoint(ctx context.Context, webhookUrl string) {
	session, err := welcomer.AcquireSession(ctx, welcomer.DefaultManagerName)
	if err != nil {
		welcomer.Logger.Error().Err(err).Msg("Failed to acquire session")
		return
	}

	rows, err := welcomer.Pool.Query(ctx, `SELECT user_id, membership_type FROM "user_memberships" WHERE "expires_at" >= NOW() AND "status" IN (1,2)`)
	if err != nil {
		welcomer.Logger.Error().Err(err).Msg("Failed to query active donators")
		return
	}

	activeDonators := map[int64]bool{}
	activeWelcomerPro := map[int64]bool{}

	for rows.Next() {
		var userId int64
		var membershipType int32

		if err := rows.Scan(&userId, &membershipType); err != nil {
			welcomer.Logger.Error().Err(err).Msg("Failed to scan user ID")
			continue
		}

		activeDonators[userId] = true

		if database.MembershipType(membershipType) == database.MembershipTypeLegacyWelcomerPro ||
			database.MembershipType(membershipType) == database.MembershipTypeWelcomerPro {
			activeWelcomerPro[userId] = true
			activeDonators[userId] = true
		}
	}

	rows.Close()

	guildMembers, err := welcomer.SandwichClient.FetchGuildMember(ctx, &sandwich.FetchGuildMemberRequest{
		GuildId: int64(WelcomerGuild),
	})
	if err != nil {
		welcomer.Logger.Error().Err(err).Msg("Failed to fetch guild members")
		return
	}

	rolesAssigned := 0
	rolesUnassigned := 0

	for _, memberPB := range guildMembers.GetGuildMembers() {
		hasDonatorRole := false
		hasWelcomerProRole := false

		for _, roleId := range memberPB.GetRoles() {
			switch discord.Snowflake(roleId) {
			case DonatorRole:
				hasDonatorRole = true
			case WelcomerProRole:
				hasWelcomerProRole = true
			}
		}

		isActiveDonator := activeDonators[memberPB.GetUser().GetID()]
		isActiveWelcomerPro := activeWelcomerPro[memberPB.GetUser().GetID()]

		member := sandwich.PBToGuildMember(memberPB)
		member.GuildID = &WelcomerGuild

		if isActiveDonator && !hasDonatorRole {
			welcomer.Logger.Info().
				Int64("user_id", int64(member.User.ID)).
				Msg("Assigning donator role to user")

			err := member.AddRoles(ctx, session, []discord.Snowflake{DonatorRole}, welcomer.ToPointer("Active welcomer donator"), true)
			if err != nil {
				welcomer.Logger.Warn().Err(err).
					Int64("user_id", int64(member.User.ID)).
					Msg("Failed to add donator role to user")
			} else {
				rolesAssigned++
			}
		} else if !isActiveDonator && hasDonatorRole {
			welcomer.Logger.Info().
				Int64("user_id", int64(member.User.ID)).
				Msg("Removing donator role from user")

			err := member.RemoveRoles(ctx, session, []discord.Snowflake{DonatorRole}, welcomer.ToPointer("No longer active welcomer donator"), true)
			if err != nil {
				welcomer.Logger.Warn().Err(err).
					Int64("user_id", int64(member.User.ID)).
					Msg("Failed to remove donator role from user")
			} else {
				rolesUnassigned++
			}
		}

		if isActiveWelcomerPro && !hasWelcomerProRole {
			welcomer.Logger.Info().
				Int64("user_id", int64(member.User.ID)).
				Msg("Assigning welcomer pro role to user")

			err := member.AddRoles(ctx, session, []discord.Snowflake{WelcomerProRole}, welcomer.ToPointer("Active welcomer pro membership"), true)
			if err != nil {
				welcomer.Logger.Warn().Err(err).
					Int64("user_id", int64(member.User.ID)).
					Msg("Failed to add welcomer pro role to user")
			} else {
				rolesAssigned++
			}
		} else if !isActiveWelcomerPro && hasWelcomerProRole {
			welcomer.Logger.Info().
				Int64("user_id", int64(member.User.ID)).
				Msg("Removing welcomer pro role from user")

			err := member.RemoveRoles(ctx, session, []discord.Snowflake{WelcomerProRole}, welcomer.ToPointer("No longer active welcomer pro membership"), true)
			if err != nil {
				welcomer.Logger.Warn().Err(err).
					Int64("user_id", int64(member.User.ID)).
					Msg("Failed to remove welcomer pro role from user")
			} else {
				rolesUnassigned++
			}
		}
	}

	welcomer.Logger.Info().
		Int("roles_assigned", rolesAssigned).
		Int("roles_unassigned", rolesUnassigned).
		Msg("Finished updating donator roles")
}
