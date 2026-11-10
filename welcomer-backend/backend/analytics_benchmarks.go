package backend

import (
	"fmt"
	"sync"
	"time"

	discord "github.com/WelcomerTeam/Discord/discord"
	sandwich "github.com/WelcomerTeam/Sandwich-Daemon/proto"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	"github.com/gin-gonic/gin"
)

var guildAnalyticsBenchmarkSchemaState struct {
	sync.Mutex
	ready bool
}

type guildAnalyticsBenchmarkMetric struct {
	Key            string   `json:"key"`
	Label          string   `json:"label"`
	Unit           string   `json:"unit"`
	ServerValue    *float64 `json:"server_value"`
	BenchmarkValue *float64 `json:"benchmark_value"`
}

type guildAnalyticsBenchmarksResponse struct {
	Statistic string                          `json:"statistic"`
	UpdatedAt *time.Time                      `json:"updated_at"`
	Metrics   []guildAnalyticsBenchmarkMetric `json:"metrics"`
}

func ensureGuildAnalyticsBenchmarkSchema(ctx *gin.Context) error {
	guildAnalyticsBenchmarkSchemaState.Lock()
	defer guildAnalyticsBenchmarkSchemaState.Unlock()

	if guildAnalyticsBenchmarkSchemaState.ready {
		return nil
	}

	if err := welcomer.Queries.EnsureGuildAnalyticsBenchmarkSchema(ctx); err != nil {
		return fmt.Errorf("ensure guild analytics benchmark schema: %w", err)
	}

	guildAnalyticsBenchmarkSchemaState.ready = true
	return nil
}

func getGuildAnalyticsBenchmarks(
	ctx *gin.Context,
	guildID discord.Snowflake,
) (*guildAnalyticsBenchmarksResponse, error) {
	if err := ensureGuildAnalyticsBenchmarkSchema(ctx); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	window := welcomer.GuildAnalyticsBenchmarkWindowFor(now)
	inputs, err := welcomer.Queries.GetGuildAnalyticsBenchmarkInputs(
		ctx,
		database.GetGuildAnalyticsBenchmarkInputsParams{
			GuildID:  int64(guildID),
			WeekFrom: window.WeekFrom,
			WeekTo:   window.WeekTo,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("get guild benchmark activity inputs: %w", err)
	}

	values := make(map[string]*float64, 9)
	var bucket welcomer.GuildAnalyticsBenchmarkBucket
	if len(inputs) > 0 {
		input := inputs[0]
		bucket, _ = welcomer.GuildAnalyticsBenchmarkBucketForMembers(int64(input.MemberCount))

		dailyActivityEnd := window.WeekTo.Add(-time.Microsecond)
		activeUsers, err := welcomer.Queries.GetActiveUserCounts(
			ctx,
			database.GetActiveUserCountsParams{
				GuildID:  int64(guildID),
				DateFrom: window.WeekFrom,
				DateTo:   dailyActivityEnd,
			},
		)
		if err != nil {
			return nil, fmt.Errorf("get guild benchmark daily activity: %w", err)
		}
		values[welcomer.GuildAnalyticsBenchmarkMetricDailyActiveUsers] =
			floatPointer(activeUsers.AverageDailyActiveUsers)

		uniqueMessagingUsers := int64(0)
		if input.UniqueUsers.Valid {
			uniqueMessagingUsers = input.UniqueUsers.Int64
		}
		totalMessages := int64(0)
		if input.TotalMessages.Valid {
			totalMessages = input.TotalMessages.Int64
		}
		values[welcomer.GuildAnalyticsBenchmarkMetricWeeklyUniqueMessagingUsers] =
			floatPointer(float64(uniqueMessagingUsers))
		if uniqueMessagingUsers > 0 {
			values[welcomer.GuildAnalyticsBenchmarkMetricWeeklyMessagesPerUser] =
				floatPointer(float64(totalMessages) / float64(uniqueMessagingUsers))
		}
		if input.MemberCount > 0 {
			values[welcomer.GuildAnalyticsBenchmarkMetricActiveMembersPercentage] =
				floatPointer(float64(uniqueMessagingUsers) * 100 / float64(input.MemberCount))
		}

		uniqueVoiceUsers := int64(0)
		if input.UniqueUsers_2.Valid {
			uniqueVoiceUsers = input.UniqueUsers_2.Int64
		}
		values[welcomer.GuildAnalyticsBenchmarkMetricWeeklyUniqueVoiceUsers] =
			floatPointer(float64(uniqueVoiceUsers))
		if input.Sessions.Valid && input.Sessions.Int64 > 0 && input.TotalTimeMs.Valid {
			values[welcomer.GuildAnalyticsBenchmarkMetricAverageVoiceSessionMS] =
				floatPointer(float64(input.TotalTimeMs.Int64) / float64(input.Sessions.Int64))
		}

		joinMetrics, err := welcomer.Queries.GetGuildAnalyticsBenchmarkJoinMetrics(
			ctx,
			database.GetGuildAnalyticsBenchmarkJoinMetricsParams{
				GuildID:                         int64(guildID),
				JoinEventType:                   int32(database.ScienceGuildEventTypeUserJoin),
				CohortFrom:                      window.CohortFrom,
				CohortTo:                        window.CohortTo,
				LeaveEventType:                  int32(database.ScienceGuildEventTypeUserLeave),
				CohortTo30DayspgCataloginterval: window.CohortTo.AddDate(0, 0, 30),
			},
		)
		if err != nil {
			return nil, fmt.Errorf("get guild benchmark join metrics: %w", err)
		}
		if joinMetrics.CohortSize > 0 {
			values[welcomer.GuildAnalyticsBenchmarkMetricRetentionRate30D] =
				floatPointer(float64(joinMetrics.RetainedMembers) * 100 / float64(joinMetrics.CohortSize))
		}
		if joinMetrics.MessagedMembers > 0 {
			values[welcomer.GuildAnalyticsBenchmarkMetricAverageTimeToFirstMessageSeconds] =
				floatPointer(joinMetrics.AverageTimeToFirstMessageSeconds)
		}

		accountAgeDays, available, cacheErr := cachedGuildAverageAccountAgeDays(ctx, guildID, now)
		if cacheErr != nil {
			welcomer.Logger.Warn().
				Err(cacheErr).
				Int64("guild_id", int64(guildID)).
				Msg("guild account-age benchmark value unavailable because its member roster is not cached")
		} else if !available {
			welcomer.Logger.Warn().
				Int64("guild_id", int64(guildID)).
				Msg("guild account-age benchmark value unavailable because its cached member roster is empty")
		} else {
			values[welcomer.GuildAnalyticsBenchmarkMetricAverageAccountAgeDays] =
				floatPointer(accountAgeDays)
		}
	}

	response := &guildAnalyticsBenchmarksResponse{
		Statistic: "p90",
		Metrics: []guildAnalyticsBenchmarkMetric{
			{Key: welcomer.GuildAnalyticsBenchmarkMetricRetentionRate30D, Label: "30-day user retention", Unit: "percent", ServerValue: values[welcomer.GuildAnalyticsBenchmarkMetricRetentionRate30D]},
			{Key: welcomer.GuildAnalyticsBenchmarkMetricDailyActiveUsers, Label: "Average daily active users", Unit: "users", ServerValue: values[welcomer.GuildAnalyticsBenchmarkMetricDailyActiveUsers]},
			{Key: welcomer.GuildAnalyticsBenchmarkMetricWeeklyUniqueMessagingUsers, Label: "Unique messaging users (week)", Unit: "users", ServerValue: values[welcomer.GuildAnalyticsBenchmarkMetricWeeklyUniqueMessagingUsers]},
			{Key: welcomer.GuildAnalyticsBenchmarkMetricWeeklyMessagesPerUser, Label: "Messages per user", Unit: "messages_per_user", ServerValue: values[welcomer.GuildAnalyticsBenchmarkMetricWeeklyMessagesPerUser]},
			{Key: welcomer.GuildAnalyticsBenchmarkMetricWeeklyUniqueVoiceUsers, Label: "Unique voice activity users (week)", Unit: "users", ServerValue: values[welcomer.GuildAnalyticsBenchmarkMetricWeeklyUniqueVoiceUsers]},
			{Key: welcomer.GuildAnalyticsBenchmarkMetricAverageVoiceSessionMS, Label: "Average voice session", Unit: "milliseconds", ServerValue: values[welcomer.GuildAnalyticsBenchmarkMetricAverageVoiceSessionMS]},
			{Key: welcomer.GuildAnalyticsBenchmarkMetricActiveMembersPercentage, Label: "Active members (% of total)", Unit: "percent", ServerValue: values[welcomer.GuildAnalyticsBenchmarkMetricActiveMembersPercentage]},
			{Key: welcomer.GuildAnalyticsBenchmarkMetricAverageTimeToFirstMessageSeconds, Label: "Average time to first message", Unit: "seconds", ServerValue: values[welcomer.GuildAnalyticsBenchmarkMetricAverageTimeToFirstMessageSeconds]},
			{Key: welcomer.GuildAnalyticsBenchmarkMetricAverageAccountAgeDays, Label: "Average account age", Unit: "days", ServerValue: values[welcomer.GuildAnalyticsBenchmarkMetricAverageAccountAgeDays]},
		},
	}

	if bucket.Label != "" {
		benchmarks, err := welcomer.Queries.GetGuildAnalyticsBenchmarkBucket(ctx, bucket.ID)
		if err != nil {
			return nil, fmt.Errorf("get guild benchmark bucket values: %w", err)
		}

		for _, benchmark := range benchmarks {
			for index := range response.Metrics {
				if response.Metrics[index].Key != benchmark.MetricKey {
					continue
				}
				if benchmark.BenchmarkValue.Valid {
					response.Metrics[index].BenchmarkValue = floatPointer(benchmark.BenchmarkValue.Float64)
				}
				if response.UpdatedAt == nil || benchmark.UpdatedAt.After(*response.UpdatedAt) {
					updatedAt := benchmark.UpdatedAt
					response.UpdatedAt = &updatedAt
				}
				break
			}
		}
	}

	return response, nil
}

func cachedGuildAverageAccountAgeDays(
	ctx *gin.Context,
	guildID discord.Snowflake,
	now time.Time,
) (float64, bool, error) {
	response, err := welcomer.SandwichClient.FetchGuildMember(
		ctx,
		&sandwich.FetchGuildMemberRequest{GuildId: int64(guildID)},
	)
	if err != nil {
		return 0, false, err
	}

	userIDs := make([]int64, 0, len(response.GetGuildMembers()))
	for _, member := range response.GetGuildMembers() {
		if member.GetUser() != nil {
			userIDs = append(userIDs, member.GetUser().GetID())
		}
	}

	averageDays, ok := welcomer.AverageDiscordAccountAgeDays(userIDs, now)
	return averageDays, ok, nil
}

func floatPointer(value float64) *float64 {
	return &value
}
