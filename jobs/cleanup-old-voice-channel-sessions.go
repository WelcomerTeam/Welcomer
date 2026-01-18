package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"runtime/debug"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"github.com/WelcomerTeam/Discord/discord"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	"github.com/jackc/pgx/v4"
)

func main() {
	var err error

	loggingLevel := flag.String("level", os.Getenv("LOGGING_LEVEL"), "Logging level")

	postgresURL := flag.String("postgresURL", os.Getenv("POSTGRES_URL"), "Postgres connection URL")

	webhookUrl := flag.String("webhookUrl", os.Getenv("JOB_CLEANUP_EXPIRED_WELCOME_MESSAGES_WEBHOOK_URL"), "Webhook URL for logging")

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
						Title:       "Cleanup Old Voice Channel Sessions Job",
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
	welcomer.SetupDatabase(ctx, *postgresURL)

	db, err := pgx.Connect(ctx, *postgresURL)
	if err != nil {
		panic(fmt.Sprintf(`pgx.Connect(%s): %v`, *postgresURL, err.Error()))
	}

	entrypoint(ctx, db)

	if err := welcomer.Queries.UpsertJobCheckpoint(ctx, database.UpsertJobCheckpointParams{
		JobName:         "cleanup-old-voice-channel-sessions",
		LastProcessedTs: time.Now().UTC(),
	}); err != nil {
		welcomer.Logger.Error().Err(err).Msg("Failed to upsert job checkpoint")
	}

	cancel()
}

func entrypoint(ctx context.Context, db *pgx.Conn) {
	expirationDate := time.Now().Add(-time.Minute * 5)

	rows, err := welcomer.Queries.DeleteAndGetGuildVoiceChannelOpenSessionsBefore(ctx, expirationDate)
	if err != nil {
		welcomer.Logger.Error().Err(err).Msg("Failed to get old voice channel sessions")
		return
	}

	for _, row := range rows {
		totalTime := row.LastSeenTs.Sub(row.StartTs).Milliseconds()

		if totalTime == 0 {
			continue
		}

		err := welcomer.Queries.CreateVoiceChannelStat(ctx, database.CreateVoiceChannelStatParams{
			GuildID:     row.GuildID,
			ChannelID:   row.ChannelID,
			UserID:      row.UserID,
			StartTs:     row.StartTs,
			EndTs:       row.LastSeenTs,
			TotalTimeMs: totalTime,
			Inferred:    true,
		})
		if err != nil {
			welcomer.Logger.Error().
				Err(err).
				Int64("guild_id", row.GuildID).
				Int64("channel_id", row.ChannelID).
				Int64("user_id", row.UserID).
				Time("start_ts", row.StartTs).
				Time("last_seen_ts", row.LastSeenTs).
				Msg("Failed to create voice channel stat from old open session")
		}
	}

	welcomer.Logger.Info().
		Int("deleted_voice_channel_sessions_count", len(rows)).
		Msg("Cleaned up old voice channel sessions")
}
