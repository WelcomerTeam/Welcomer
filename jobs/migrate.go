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
		"science_guild_events": "created_at",
	}

	for tableName, timestampColumn := range tableNames {
		var min time.Time
		var max time.Time

		welcomer.Logger.Info().Msgf("Migrating data for table %s based on column %s", tableName, timestampColumn)

		res, err := welcomer.Pool.Query(ctx, fmt.Sprintf("SELECT MIN(%s), MAX(%s) FROM %s", timestampColumn, timestampColumn, tableName+"_original"))
		if err != nil {
			welcomer.Logger.Error().Err(err).Msg("Failed to query min and max timestamps")
			continue
		}

		welcomer.Logger.Info().Msgf("Successfully queried min and max timestamps for table %s", tableName)

		if res.Next() {
			err = res.Scan(&min, &max)
			if err != nil {
				welcomer.Logger.Error().Err(err).Msg("Failed to scan min and max timestamps")
				continue
			}
		} else {
			welcomer.Logger.Warn().Msgf("No rows returned for table %s", tableName)
			continue
		}

		start_date := min.Truncate(24 * time.Hour)

		selector := "*"

		if tableName == "science_guild_events" {
			min = time.Date(2026, 4, 17, 0, 0, 0, 0, time.UTC)
		}

		for {
			if start_date.After(max) {
				break
			}

			to_date := start_date.AddDate(0, 0, 1)

			welcomer.Logger.Info().Msgf("Migrating data for table %s from %s to %s", tableName, start_date.Format("2006-01-02"), to_date.Format("2006-01-02"))

			_, err = welcomer.Pool.Exec(ctx,
				fmt.Sprintf("INSERT INTO %s SELECT %s FROM %s WHERE %s BETWEEN $1 AND $2 ON CONFLICT DO NOTHING", tableName, selector, tableName+"_original", timestampColumn),
				start_date,
				to_date,
			)
			if err != nil {
				welcomer.Logger.Error().Err(err).Msgf("Failed to insert data into table %s", tableName)
				if err != nil {
					welcomer.Logger.Warn().Err(err).Msg("Failed to send webhook message")
				}

				panic(err)
			}

			// print("Deleting...")

			// _, err = welcomer.Pool.Exec(ctx, fmt.Sprintf(
			// 	"DELETE FROM %s WHERE %s BETWEEN $1 AND $2", tableName+"_original", timestampColumn),
			// 	start_date,
			// 	to_date,
			// )
			// if err != nil {
			// 	welcomer.Logger.Error().Err(err).Msgf("Failed to select data from table %s", tableName)
			// }

			welcomer.Logger.Info().Msgf("Migrated data for table %s from %s to %s", tableName, start_date.Format("2006-01-02"), to_date.Format("2006-01-02"))
			start_date = to_date
		}
	}
}
