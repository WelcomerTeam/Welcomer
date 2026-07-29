package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"runtime/debug"
	"time"

	"github.com/WelcomerTeam/Discord/discord"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	var err error

	loggingLevel := flag.String("level", os.Getenv("LOGGING_LEVEL"), "Logging level")

	postgresURL := flag.String("postgresURL", os.Getenv("POSTGRES_URL"), "Postgres connection URL")

	webhookUrl := flag.String("webhookUrl", os.Getenv("JOB_NOTIFY_EXPIRED_WEBHOOK_URL"), "Webhook URL for logging")

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
						Title:       "Process Queue",
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

	welcomer.SetupLogger(*loggingLevel)
	welcomer.SetupSandwichClient()
	welcomer.SetupDatabase(ctx, *postgresURL)

	// SELECT MIN(created_at), MAX(created_at) FROM science_guild_events_original
	// 2025-02-18 08:59:14.676735	2026-07-27 21:24:11.2347

	entrypoint(ctx, *webhookUrl)

	cancel()
}

func entrypoint(ctx context.Context, webhookUrl string) {
	tableNames := map[string]string{
		"guild_message_counts_hour": "hour_ts",
		"guild_voice_channel_stats": "start_ts",
		"science_guild_events":      "created_at",
	}

	for tableName, timestampColumn := range tableNames {
		now := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.Now().Location())

		err := createPartitions(ctx, tableName, timestampColumn, now)
		if err != nil {
			welcomer.Logger.Error().Err(err).Msgf("Failed to create partitions for table %s", tableName)
			err = welcomer.SendWebhookMessage(ctx, webhookUrl, discord.WebhookMessageParams{
				Content: "<@143090142360371200>",
				Embeds: []discord.Embed{
					{
						Title:       "Create Partitions",
						Description: fmt.Sprintf("Failed to create partitions for table %s: %v", tableName, err),
						Color:       int32(16760839),
						Timestamp:   new(time.Now()),
					},
				},
			})
			if err != nil {
				welcomer.Logger.Warn().Err(err).Msg("Failed to send webhook message")
			}
		}

		err = createPartitions(ctx, tableName, timestampColumn, now.AddDate(0, 1, 0))
		if err != nil {
			welcomer.Logger.Error().Err(err).Msgf("Failed to create partitions for table %s", tableName)
			err = welcomer.SendWebhookMessage(ctx, webhookUrl, discord.WebhookMessageParams{
				Content: "<@143090142360371200>",
				Embeds: []discord.Embed{
					{
						Title:       "Create Partitions",
						Description: fmt.Sprintf("Failed to create partitions for table %s: %v", tableName, err),
						Color:       int32(16760839),
						Timestamp:   new(time.Now()),
					},
				},
			})
			if err != nil {
				welcomer.Logger.Warn().Err(err).Msg("Failed to send webhook message")
			}
		}
	}
}

func createPartitions(ctx context.Context, tableName string, timestampColumn string, partitionTime time.Time) error {
	start := partitionTime
	end := partitionTime.AddDate(0, 1, 0)

	partitionName := fmt.Sprintf("%s_%d_%02d", tableName, start.Year(), start.Month())
	createPartitionQuery := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s PARTITION OF %s
		FOR VALUES FROM ('%s') TO ('%s');
	`, partitionName, tableName, start.Format("2006-01-02"), end.Format("2006-01-02"))

	println(createPartitionQuery)

	_, err := welcomer.Pool.Exec(ctx, createPartitionQuery)
	if err != nil {
		return fmt.Errorf("failed to create partition %s: %w", partitionName, err)
	}

	welcomer.Logger.Info().Msgf("Successfully created partition %s for table %s", partitionName, tableName)

	return nil
}
