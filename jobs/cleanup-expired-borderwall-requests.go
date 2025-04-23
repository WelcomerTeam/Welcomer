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

	webhookUrl := flag.String("webhookUrl", os.Getenv("JOB_CLEANUP_EXPIRED_BORDERWALL_REQUESTS_WEBHOOK_URL"), "Webhook URL for logging")

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
						Title:       "Cleanup Expired Borderwall Requests Job",
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

	entrypoint(ctx, *webhookUrl)

	cancel()
}

func entrypoint(ctx context.Context, webhookUrl string) {
	rows, err := welcomer.Pool.Query(ctx, "DELETE FROM borderwall_requests WHERE is_verified = false AND updated_at < $1", time.Now().Add(time.Hour*(-1*24*90)).UTC())
	if err != nil {
		welcomer.Logger.Error().Err(err).
			Msg("Failed to delete expired borderwall requests")

		panic(err)
	}

	rows.Close()

	welcomer.Logger.Info().
		Int64("deleted_count", rows.CommandTag().RowsAffected()).
		Msg("Expired borderwall requests deleted successfully")
}
