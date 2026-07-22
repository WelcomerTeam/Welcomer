package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"runtime/debug"
	"time"

	"github.com/WelcomerTeam/Discord/discord"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	"github.com/jackc/pgtype"
	_ "github.com/joho/godotenv/autoload"
)

type ModerationRules struct {
	MaximumChangeScore       float64 `json:"maximum_change_score"`
	MinimumSafeScore         float64 `json:"minimum_safe_score"`
	MaximumQuestionableScore float64 `json:"maximum_questionable_score"`
	MaximumExplicitScore     float64 `json:"maximum_explicit_score"`
	MaximumInvites           int     `json:"maximum_invites"`
	MaximumUrls              int     `json:"maximum_urls"`
}

func main() {
	var err error

	loggingLevel := flag.String("level", os.Getenv("LOGGING_LEVEL"), "Logging level")

	postgresURL := flag.String("postgresURL", os.Getenv("POSTGRES_URL"), "Postgres connection URL")

	webhookUrl := flag.String("webhookUrl", os.Getenv("JOB_NOTIFY_EXPIRED_WEBHOOK_URL"), "Webhook URL for logging")

	modCoreUrl := flag.String("modCoreUrl", os.Getenv("MOD_CORE_URL"), "URL for the moderation core service")

	modCoreMaximumChangeScore := flag.Float64("maxChangeScore", 0.6, "Values above threshold for change will be blocked for moderation rules")
	modCoreMinimumSafeScore := flag.Float64("minSafeScore", -1, "Values below threshold for safe score will be blocked for moderation rules")
	modCoreMaximumQuestionableScore := flag.Float64("maxQuestionableScore", 0.9, "Values above threshold for questionable score will be blocked for moderation rules")
	modCoreMaximumExplicitScore := flag.Float64("maxExplicitScore", 0.6, "Values above threshold for explicit score will be blocked for moderation rules")
	modCoreMaximumInvites := flag.Int("maxInvites", 3, "Maximum invites for moderation rules")
	modCoreMaximumUrls := flag.Int("maxUrls", -1, "Maximum URLs for moderation rules")

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

	for {
		entrypoint(ctx, *webhookUrl, *modCoreUrl, ModerationRules{
			MaximumChangeScore:       *modCoreMaximumChangeScore,
			MinimumSafeScore:         *modCoreMinimumSafeScore,
			MaximumQuestionableScore: *modCoreMaximumQuestionableScore,
			MaximumExplicitScore:     *modCoreMaximumExplicitScore,
			MaximumInvites:           *modCoreMaximumInvites,
			MaximumUrls:              *modCoreMaximumUrls,
		})

		time.Sleep(time.Second * 10)
	}

	cancel()
}

type ModerationCoreRequest struct {
	Texts []string `json:"texts"`
}

type ModerationCoreResponse struct {
	Results []ModerationCoreResponseItems `json:"results`
}

type ModerationCoreResponseItems struct {
	URLs              []string `json:"urls"`
	Invites           []string `json:"invites"`
	ChangeScore       float64  `json:"change_score"`
	SafeScore         float64  `json:"safe_score"`
	QuestionableScore float64  `json:"questionable_score"`
	ExplicitScore     float64  `json:"explicit_score"`
}

func entrypoint(ctx context.Context, webhookUrl string, modCoreUrl string, modCoreRules ModerationRules) {
	chunkSize := 10

	for {
		queueValues, err := welcomer.Queries.FetchModerationCheckupQueue(ctx, int32(chunkSize))
		if err != nil {
			welcomer.Logger.Error().Err(err).Msg("Failed to fetch moderation checkup queue")

			panic(err)
		}

		if len(queueValues) > 30 {
			queueValues = queueValues[:30]
		}

		if len(queueValues) == 0 {
			welcomer.Logger.Info().Msg("No moderation checkup queue values found")

			return
		}

		chunk := make([]string, len(queueValues))

		for i, queue := range queueValues {
			chunk[i] = queue.Value
		}

		body, err := json.Marshal(ModerationCoreRequest{Texts: chunk})
		if err != nil {
			welcomer.Logger.Error().Err(err).Msg("Failed to marshal request body for moderation core")

			panic(err)
		}

		start := time.Now()

		welcomer.Logger.Info().Int("count", len(chunk)).Msg("Sending request to moderation core")

		req, err := http.Post(modCoreUrl, "application/json", bytes.NewBuffer(body))
		if err != nil {
			welcomer.Logger.Error().Err(err).Msg("Failed to send request to moderation core")

			panic(err)
		}

		welcomer.Logger.Info().Int("status", req.StatusCode).Dur("elapsed", time.Now().Sub(start)).Msg("Received response from moderation core")

		defer req.Body.Close()

		var moderationCoreResponse ModerationCoreResponse

		err = json.NewDecoder(req.Body).Decode(&moderationCoreResponse)
		if err != nil {
			welcomer.Logger.Error().Err(err).Msg("Failed to decode response from moderation core")

			panic(err)
		}

		for i, result := range moderationCoreResponse.Results {
			queue := queueValues[i]

			blocked := isBlocked(database.AuditType(queue.DataType), result, modCoreRules)

			welcomer.Logger.Info().
				Str("checkup_uuid", queue.CheckupQueueUuid.String()).
				Int64("guild_id", int64(queue.GuildID)).
				Float64("change_score", result.ChangeScore).
				Float64("safe_score", result.SafeScore).
				Float64("questionable_score", result.QuestionableScore).
				Float64("explicit_score", result.ExplicitScore).
				Bool("is_blocked", blocked).
				Msg("Updating moderation checkup")

			err = welcomer.Queries.UpdateModerationCheckup(ctx, database.UpdateModerationCheckupParams{
				CheckupUuid:   queue.CheckupQueueUuid,
				Dom:           marshalList(result.URLs),
				Inv:           marshalList(result.Invites),
				ScoreChange:   sql.NullFloat64{Float64: result.ChangeScore, Valid: true},
				ScoreSafe:     sql.NullFloat64{Float64: result.SafeScore, Valid: true},
				ScoreQuestion: sql.NullFloat64{Float64: result.QuestionableScore, Valid: true},
				ScoreExplicit: sql.NullFloat64{Float64: result.ExplicitScore, Valid: true},
				IsBlocked:     sql.NullBool{Bool: blocked, Valid: true},
			})
			if err != nil {
				welcomer.Logger.Error().Err(err).Msg("Failed to update moderation checkup")
			}

			err = welcomer.Queries.RemoveFromModerationCheckupQueue(ctx, queue.CheckupQueueUuid)
			if err != nil {
				welcomer.Logger.Error().Err(err).Msg("Failed to remove from moderation checkup queue")
			}
		}
	}
}

func isBlocked(dataType database.AuditType, result ModerationCoreResponseItems, rules ModerationRules) bool {
	if dataType != database.AuditTypeGuildSettingsRules {
		if rules.MinimumSafeScore != -1 && result.SafeScore < rules.MinimumSafeScore {
			return true
		}

		if rules.MaximumQuestionableScore != -1 && result.QuestionableScore > rules.MaximumQuestionableScore {
			return true
		}

		if rules.MaximumExplicitScore != -1 && result.ExplicitScore > rules.MaximumExplicitScore {
			return true
		}
	}

	if rules.MaximumChangeScore != -1 && result.ChangeScore > rules.MaximumChangeScore {
		return true
	}

	if rules.MaximumInvites != -1 && len(result.Invites) > rules.MaximumInvites {
		return true
	}

	if rules.MaximumUrls != -1 && len(result.URLs) > rules.MaximumUrls {
		return true
	}

	return false
}

func marshalList(list []string) pgtype.JSONB {
	if len(list) == 0 {
		return pgtype.JSONB{Status: pgtype.Null}
	}

	data, err := json.Marshal(list)
	if err != nil {
		welcomer.Logger.Error().Err(err).Msg("Failed to marshal list to JSONB")

		return pgtype.JSONB{Status: pgtype.Null}
	}

	return pgtype.JSONB{Bytes: data, Status: pgtype.Present}
}
