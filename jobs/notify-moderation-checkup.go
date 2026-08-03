package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"runtime/debug"
	"time"

	"github.com/WelcomerTeam/Discord/discord"
	sandwich "github.com/WelcomerTeam/Sandwich-Daemon/proto"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/jackc/pgx/v4"
	_ "github.com/joho/godotenv/autoload"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	var err error

	loggingLevel := flag.String("level", os.Getenv("LOGGING_LEVEL"), "Logging level")

	postgresURL := flag.String("postgresURL", os.Getenv("POSTGRES_URL"), "Postgres connection URL")

	webhookUrl := flag.String("webhookUrl", os.Getenv("JOB_CLEANUP_EXPIRED_WELCOME_MESSAGES_WEBHOOK_URL"), "Webhook URL for logging")

	sandwichGRPCHost := flag.String("sandwichGRPCHost", os.Getenv("SANDWICH_GRPC_HOST"), "GRPC Address for the Sandwich Daemon service")

	sandwichManagerName := flag.String("sandwichManagerName", os.Getenv("SANDWICH_MANAGER_NAME"), "Sandwich manager identifier name")

	proxyAddress := flag.String("proxyAddress", os.Getenv("PROXY_ADDRESS"), "Address to proxy requests through. This can be 'https://discord.com', if one is not setup.")
	proxyDebug := flag.Bool("proxyDebug", false, "Enable debugging requests to the proxy")

	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from panic:", r)
			println(string(debug.Stack()))

			err = welcomer.SendWebhookMessage(ctx, *webhookUrl, discord.WebhookMessageParams{
				Content: "<@143090142360371200>",
				Embeds: []discord.Embed{
					{
						Title:       "Notify Moderation Checkup Job",
						Description: fmt.Sprintf("Recovered from panic: %v", r),
						Color:       int32(16760839),
						Timestamp:   new(time.Now()),
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
	welcomer.SetupRESTInterface(restInterface)

	welcomer.SetupDefaultManagerName(*sandwichManagerName)
	welcomer.SetupLogger(*loggingLevel)
	welcomer.SetupDatabase(ctx, *postgresURL)
	welcomer.SetupGRPCConnection(*sandwichGRPCHost,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(1024*1024*1024)), // Set max message size to 1GB
	)
	welcomer.SetupSandwichClient()

	db, err := pgx.Connect(ctx, *postgresURL)
	if err != nil {
		panic(fmt.Sprintf(`pgx.Connect(%s): %v`, *postgresURL, err.Error()))
	}

	entrypoint(ctx, db)
}

const sqlQuery = `
SELECT
guilds.guild_id,
mc1.is_blocked rules_blocked,
mc2.is_blocked borderwall_blocked,
mc3.is_blocked text_blocked,
mc4.is_blocked dms_blocked
FROM

guilds
LEFT JOIN guild_settings_rules ON guild_settings_rules.guild_id = guilds.guild_id
LEFT JOIN moderation_checkup mc1 ON mc1.checkup_uuid = guild_settings_rules.moderation_checkup_uuid

LEFT JOIN guild_settings_borderwall ON guild_settings_borderwall.guild_id = guilds.guild_id
LEFT JOIN moderation_checkup mc2 ON mc2.checkup_uuid = guild_settings_borderwall.moderation_checkup_uuid

LEFT JOIN guild_settings_welcomer_text ON guild_settings_welcomer_text.guild_id = guilds.guild_id
LEFT JOIN moderation_checkup mc3 ON mc3.checkup_uuid = guild_settings_welcomer_text.moderation_checkup_uuid

LEFT JOIN guild_settings_welcomer_dms ON guild_settings_welcomer_dms.guild_id = guilds.guild_id
LEFT JOIN moderation_checkup mc4 ON mc4.checkup_uuid = guild_settings_welcomer_dms.moderation_checkup_uuid

WHERE mc1.is_blocked = TRUE or mc2.is_blocked = TRUE or mc3.is_blocked = TRUE or mc4.is_blocked
ORDER BY guilds.guild_id DESC
`

type ModerationCheckup struct {
	GuildID           discord.Snowflake `json:"guild_id"`
	RulesBlocked      sql.NullBool      `json:"rules_blocked"`
	BorderwallBlocked sql.NullBool      `json:"borderwall_blocked"`
	TextBlocked       sql.NullBool      `json:"text_blocked"`
	DmsBlocked        sql.NullBool      `json:"dms_blocked"`
}

func entrypoint(ctx context.Context, db *pgx.Conn) {
	rows, err := db.Query(ctx, sqlQuery)
	if err != nil {
		welcomer.Logger.Error().Err(err).Msg("Failed to query moderation checkup")
		return
	}
	defer rows.Close()

	var checkups []ModerationCheckup

	for rows.Next() {
		var checkup ModerationCheckup
		err := rows.Scan(&checkup.GuildID, &checkup.RulesBlocked, &checkup.BorderwallBlocked, &checkup.TextBlocked, &checkup.DmsBlocked)
		if err != nil {
			welcomer.Logger.Error().Err(err).Msg("Failed to scan moderation checkup row")
			continue
		}
		checkups = append(checkups, checkup)
	}

	session, err := welcomer.AcquireSession(ctx, welcomer.DefaultManagerName)
	if err != nil {
		welcomer.Logger.Panic().Err(err).
			Msg("Failed to acquire session")
	}

	for _, checkup := range checkups {
		if checkup.GuildID >= 613856329064513536 {
			continue
		}

		guilds, err := welcomer.SandwichClient.FetchGuild(ctx, &sandwich.FetchGuildRequest{
			GuildIds: []int64{int64(checkup.GuildID)},
		})
		if err != nil {
			welcomer.Logger.Info().Err(err).Str("guild_id", checkup.GuildID.String()).Msg("Failed to fetch guild from Sandwich")

			continue
		}

		var guild *sandwich.Guild

		if len(guilds.Guilds) == 0 {
			continue
		}

		guild = guilds.Guilds[int64(checkup.GuildID)]

		user := discord.User{ID: discord.Snowflake(guild.OwnerID)}

		_, err = user.Send(ctx, session, discord.MessageParams{
			Content: fmt.Sprintf(`Hello,

Your guild, **%s**, has had some settings flagged for violating our guidelines and we will no longer show these messages to your users until resolved. Please review your settings and ensure they comply with our guidelines.

Flagged content can include spam, inappropriate or explicit material, harassment or hate speech, scams, malicious content, impersonation, privacy/copyright violations, illegal activity, or attempts to disrupt the service. Our Acceptable Use policy can be found [here](https://www.welcomer.gg/terms). If you believe this is a mistake, please let us know on our [support server](https://discord.gg/welcomer).

Flagged settings:
%s
-# This is an automated message. Replies to this message will not be seen.`, guild.Name, flaggedSettings(guild.ID, checkup)),
		})
		if err != nil {
			welcomer.Logger.Info().Err(err).Str("guild_id", checkup.GuildID.String()).Msg("Failed to send DM to guild owner")
		} else {
			welcomer.Logger.Info().Str("guild_id", checkup.GuildID.String()).Msg("Sent DM to guild owner")
		}
	}
}

func flaggedSettings(guildID int64, checkup ModerationCheckup) string {
	var flagged string

	if checkup.RulesBlocked.Valid && checkup.RulesBlocked.Bool {
		flagged += fmt.Sprintf("Rules ([edit](https://welcomer.gg/dashboard/%d/rules))\n", guildID)
	}
	if checkup.BorderwallBlocked.Valid && checkup.BorderwallBlocked.Bool {
		flagged += fmt.Sprintf("Borderwall ([edit](https://welcomer.gg/dashboard/%d/borderwall))\n", guildID)
	}
	if checkup.TextBlocked.Valid && checkup.TextBlocked.Bool {
		flagged += fmt.Sprintf("Welcomer Text ([edit](https://welcomer.gg/dashboard/%d/welcomer))\n", guildID)
	}
	if checkup.DmsBlocked.Valid && checkup.DmsBlocked.Bool {
		flagged += fmt.Sprintf("Welcomer DMs ([edit](https://welcomer.gg/dashboard/%d/welcomer))\n", guildID)
	}

	print(flagged)

	return flagged
}
