package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"runtime/debug"
	"time"

	"github.com/jackc/pgx/v4"
	_ "github.com/joho/godotenv/autoload"

	"github.com/WelcomerTeam/Discord/discord"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
)

const (
	MaxRowRetention = time.Hour * 24 * 3 // 3 days
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
						Title:       "Cleanup Ingest Tables Job",
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
		JobName:         "cleanup-ingest-tables",
		LastProcessedTs: time.Now().UTC(),
	}); err != nil {
		welcomer.Logger.Error().Err(err).Msg("Failed to upsert job checkpoint")
	}

	cancel()
}

func entrypoint(ctx context.Context, db *pgx.Conn) {
	expirationDate := time.Now().Add(-MaxRowRetention)

	rows, err := welcomer.Pool.Query(ctx, "DELETE FROM ingest_message_events WHERE occurred_at < $1", expirationDate)
	if err != nil {
		welcomer.Logger.Error().Err(err).Msg("Failed to clean up ingest welcome messages")
		return
	}

	rows.Close()

	welcomer.Logger.Info().
		Int64("deleted_welcome_messages_count", rows.CommandTag().RowsAffected()).
		Msg("Ingest welcome messages cleaned up successfully")

	rows, err = welcomer.Pool.Query(ctx, "DELETE FROM ingest_voice_channel_events WHERE occurred_at < $1", expirationDate)
	if err != nil {
		welcomer.Logger.Error().Err(err).Msg("Failed to clean up ingest voice channel events")
		return
	}

	rows.Close()

	welcomer.Logger.Info().
		Int64("deleted_voice_channel_events_count", rows.CommandTag().RowsAffected()).
		Msg("Ingest voice channel events cleaned up successfully")

	rows, err = welcomer.Pool.Query(ctx, "DELETE FROM guild_voice_channel_open_sessions WHERE last_seen_ts < $1", expirationDate)
	if err != nil {
		welcomer.Logger.Error().Err(err).Msg("Failed to clean up guild voice channel open sessions")
		return
	}

	rows.Close()

	welcomer.Logger.Info().
		Int64("deleted_guild_voice_channel_open_sessions_count", rows.CommandTag().RowsAffected()).
		Msg("Guild voice channel open sessions cleaned up successfully")
}
