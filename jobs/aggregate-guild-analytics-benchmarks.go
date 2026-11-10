package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"math"
	"math/rand"
	"os"
	"runtime/debug"
	"sort"
	"sync"
	"time"

	sandwich "github.com/WelcomerTeam/Sandwich-Daemon/proto"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	_ "github.com/joho/godotenv/autoload"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var benchmarkMetricKeys = []string{
	welcomer.GuildAnalyticsBenchmarkMetricRetentionRate30D,
	welcomer.GuildAnalyticsBenchmarkMetricDailyActiveUsers,
	welcomer.GuildAnalyticsBenchmarkMetricWeeklyUniqueMessagingUsers,
	welcomer.GuildAnalyticsBenchmarkMetricWeeklyMessagesPerUser,
	welcomer.GuildAnalyticsBenchmarkMetricWeeklyUniqueVoiceUsers,
	welcomer.GuildAnalyticsBenchmarkMetricAverageVoiceSessionMS,
	welcomer.GuildAnalyticsBenchmarkMetricActiveMembersPercentage,
	welcomer.GuildAnalyticsBenchmarkMetricAverageTimeToFirstMessageSeconds,
	welcomer.GuildAnalyticsBenchmarkMetricAverageAccountAgeDays,
}

const (
	guildAnalyticsBenchmarkSampleLimit      = 5000
	guildAnalyticsBenchmarkProgressEvery    = 100
	guildAnalyticsBenchmarkProgressInterval = 30 * time.Second
)

func main() {
	loggingLevel := flag.String("level", os.Getenv("LOGGING_LEVEL"), "Logging level")
	postgresURL := flag.String("postgresURL", os.Getenv("POSTGRES_URL"), "Postgres connection URL")
	sandwichGRPCHost := flag.String(
		"sandwichGRPCHost",
		os.Getenv("SANDWICH_GRPC_HOST"),
		"GRPC address for the Sandwich Daemon service",
	)
	sandwichManagerName := flag.String(
		"sandwichManagerName",
		os.Getenv("SANDWICH_MANAGER_NAME"),
		"Sandwich manager identifier name",
	)
	flag.Parse()

	ctx := context.Background()
	defer func() {
		if recovered := recover(); recovered != nil {
			welcomer.Logger.Error().
				Interface("panic", recovered).
				Bytes("stack", debug.Stack()).
				Msg("guild analytics benchmark job panicked")
			os.Exit(1)
		}
	}()

	welcomer.SetupLogger(*loggingLevel)
	welcomer.SetupDefaultManagerName(*sandwichManagerName)
	welcomer.SetupGRPCConnection(
		*sandwichGRPCHost,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(1024*1024*1024)),
	)
	welcomer.SetupSandwichClient()
	welcomer.SetupDatabase(ctx, *postgresURL)

	if err := aggregateGuildAnalyticsBenchmarks(ctx); err != nil {
		welcomer.Logger.Fatal().Err(err).Msg("failed to aggregate guild analytics benchmarks")
	}
}

func aggregateGuildAnalyticsBenchmarks(ctx context.Context) error {
	jobStartedAt := time.Now()
	welcomer.Logger.Info().
		Int("sample_limit_per_bucket", guildAnalyticsBenchmarkSampleLimit).
		Msg("starting guild analytics benchmark aggregation")

	if err := welcomer.Queries.EnsureGuildAnalyticsBenchmarkSchema(ctx); err != nil {
		return fmt.Errorf("ensure benchmark table exists: %w", err)
	}
	welcomer.Logger.Info().Msg("benchmark storage table is ready")

	snapshotAt := time.Now().UTC()
	activeSince := snapshotAt.AddDate(0, 0, -30)
	sampleStartedAt := time.Now()
	welcomer.Logger.Info().
		Time("science_event_since", activeSince).
		Msg("loading guilds with at least one science event in the last 30 days")
	guilds, err := welcomer.Queries.GetGuildAnalyticsBenchmarkGuilds(
		ctx,
		database.GetGuildAnalyticsBenchmarkGuildsParams{
			ActiveSince: activeSince,
			ActiveUntil: snapshotAt,
		},
	)
	if err != nil {
		return fmt.Errorf("get eligible guilds: %w", err)
	}

	welcomer.Logger.Info().
		Int("eligible_guilds", len(guilds)).
		Dur("load_duration", time.Since(sampleStartedAt)).
		Msg("loaded eligible guilds")
	samplingStartedAt := time.Now()
	sampledGuilds, eligibleCounts := sampleGuildsPerBucket(
		guilds,
		guildAnalyticsBenchmarkSampleLimit,
		rand.New(rand.NewSource(time.Now().UnixNano())),
	)
	guilds = sampledGuilds
	sampledCounts := make(map[int16]int, len(eligibleCounts))
	for _, guild := range guilds {
		bucket, ok := welcomer.GuildAnalyticsBenchmarkBucketForMembers(int64(guild.MemberCount))
		if !ok {
			return fmt.Errorf("guild %d has no benchmark bucket", guild.GuildID)
		}
		sampledCounts[bucket.ID]++
	}
	welcomer.Logger.Info().
		Int("sampled_guilds", len(guilds)).
		Int("sample_limit_per_bucket", guildAnalyticsBenchmarkSampleLimit).
		Dur("sampling_duration", time.Since(samplingStartedAt)).
		Msg("random guild sample selected")
	for _, bucket := range welcomer.GuildAnalyticsBenchmarkBuckets() {
		welcomer.Logger.Info().
			Int("bucket_id", int(bucket.ID)).
			Str("bucket", bucket.Label).
			Int("eligible_guilds", eligibleCounts[bucket.ID]).
			Int("sampled_guilds", sampledCounts[bucket.ID]).
			Msg("selected guild sample for member-count bucket")
	}

	window := welcomer.GuildAnalyticsBenchmarkWindowFor(snapshotAt)
	welcomer.Logger.Info().
		Time("snapshot_at", snapshotAt).
		Time("science_event_since", activeSince).
		Time("activity_from", window.WeekFrom).
		Time("activity_to_exclusive", window.WeekTo).
		Time("cohort_from", window.CohortFrom).
		Time("cohort_to_exclusive", window.CohortTo).
		Msg("processing fixed benchmark metric windows")
	processingStartedAt := time.Now()
	lastProgressLog := processingStartedAt
	samples := make(map[int16]map[string][]float64)
	cacheFetchErrors := 0
	emptyCachedRosters := 0
	noActivityRows := 0
	firstCacheError := ""
	for i, guild := range guilds {
		bucket, ok := welcomer.GuildAnalyticsBenchmarkBucketForMembers(int64(guild.MemberCount))
		if !ok {
			return fmt.Errorf("guild %d has no benchmark bucket", guild.GuildID)
		}

		if i == 0 || i%guildAnalyticsBenchmarkProgressEvery == 0 ||
			time.Since(lastProgressLog) >= guildAnalyticsBenchmarkProgressInterval {
			elapsed := time.Since(processingStartedAt)
			estimatedRemaining := time.Duration(0)
			if i > 0 {
				estimatedRemaining = time.Duration(
					float64(elapsed) * float64(len(guilds)-i) / float64(i),
				)
			}
			progress := 0.0
			if len(guilds) > 0 {
				progress = float64(i) * 100 / float64(len(guilds))
			}

			welcomer.Logger.Info().
				Int("processed_guilds", i).
				Int("sampled_guilds", len(guilds)).
				Float64("progress_percent", progress).
				Int64("current_guild_id", guild.GuildID).
				Str("current_bucket", bucket.Label).
				Int("account_age_cache_errors", cacheFetchErrors).
				Int("empty_cached_rosters", emptyCachedRosters).
				Int("guilds_without_activity_rows", noActivityRows).
				Dur("elapsed", elapsed).
				Dur("estimated_remaining", estimatedRemaining).
				Msg("processing sampled guilds")
			lastProgressLog = time.Now()
		}

		queryResults, err := fetchGuildBenchmarkQueries(ctx, welcomer.Queries, guild.GuildID, window)
		if err != nil {
			return err
		}
		if queryResults.input == nil {
			noActivityRows++
			continue
		}

		input := queryResults.input
		addBenchmarkSample(samples, bucket.ID, welcomer.GuildAnalyticsBenchmarkMetricDailyActiveUsers, queryResults.activeUsers.AverageDailyActiveUsers)

		uniqueMessagingUsers := int64(0)
		if input.UniqueUsers.Valid {
			uniqueMessagingUsers = input.UniqueUsers.Int64
		}
		totalMessages := int64(0)
		if input.TotalMessages.Valid {
			totalMessages = input.TotalMessages.Int64
		}
		addBenchmarkSample(samples, bucket.ID, welcomer.GuildAnalyticsBenchmarkMetricWeeklyUniqueMessagingUsers, float64(uniqueMessagingUsers))
		if uniqueMessagingUsers > 0 {
			addBenchmarkSample(
				samples,
				bucket.ID,
				welcomer.GuildAnalyticsBenchmarkMetricWeeklyMessagesPerUser,
				float64(totalMessages)/float64(uniqueMessagingUsers),
			)
		}
		if input.MemberCount > 0 {
			addBenchmarkSample(
				samples,
				bucket.ID,
				welcomer.GuildAnalyticsBenchmarkMetricActiveMembersPercentage,
				float64(uniqueMessagingUsers)*100/float64(input.MemberCount),
			)
		}

		uniqueVoiceUsers := int64(0)
		if input.UniqueUsers_2.Valid {
			uniqueVoiceUsers = input.UniqueUsers_2.Int64
		}
		addBenchmarkSample(samples, bucket.ID, welcomer.GuildAnalyticsBenchmarkMetricWeeklyUniqueVoiceUsers, float64(uniqueVoiceUsers))
		if input.Sessions.Valid && input.Sessions.Int64 > 0 && input.TotalTimeMs.Valid {
			addBenchmarkSample(
				samples,
				bucket.ID,
				welcomer.GuildAnalyticsBenchmarkMetricAverageVoiceSessionMS,
				float64(input.TotalTimeMs.Int64)/float64(input.Sessions.Int64),
			)
		}

		joinMetrics := queryResults.joinMetrics
		if joinMetrics.CohortSize > 0 {
			addBenchmarkSample(
				samples,
				bucket.ID,
				welcomer.GuildAnalyticsBenchmarkMetricRetentionRate30D,
				float64(joinMetrics.RetainedMembers)*100/float64(joinMetrics.CohortSize),
			)
		}
		if joinMetrics.MessagedMembers > 0 {
			addBenchmarkSample(
				samples,
				bucket.ID,
				welcomer.GuildAnalyticsBenchmarkMetricAverageTimeToFirstMessageSeconds,
				joinMetrics.AverageTimeToFirstMessageSeconds,
			)
		}

		accountAgeDays, available, err := cachedGuildAverageAccountAgeDays(ctx, guild.GuildID, snapshotAt)
		if err != nil {
			cacheFetchErrors++
			if firstCacheError == "" {
				firstCacheError = err.Error()
			}
		} else if !available {
			emptyCachedRosters++
		} else {
			addBenchmarkSample(samples, bucket.ID, welcomer.GuildAnalyticsBenchmarkMetricAverageAccountAgeDays, accountAgeDays)
		}
	}

	welcomer.Logger.Info().
		Int("processed_guilds", len(guilds)).
		Int("sampled_guilds", len(guilds)).
		Int("account_age_cache_errors", cacheFetchErrors).
		Int("empty_cached_rosters", emptyCachedRosters).
		Int("guilds_without_activity_rows", noActivityRows).
		Dur("processing_duration", time.Since(processingStartedAt)).
		Msg("finished processing sampled guilds")
	if cacheFetchErrors > 0 {
		welcomer.Logger.Warn().
			Int("guilds", cacheFetchErrors).
			Str("first_error", firstCacheError).
			Msg("some account-age values were excluded because cached member rosters could not be fetched")
	}
	if emptyCachedRosters > 0 {
		welcomer.Logger.Warn().
			Int("guilds", emptyCachedRosters).
			Msg("some account-age values were excluded because cached member rosters were empty")
	}
	if noActivityRows > 0 {
		welcomer.Logger.Warn().
			Int("guilds", noActivityRows).
			Msg("some sampled guilds had no benchmark activity row")
	}

	if err := persistGuildAnalyticsBenchmarks(ctx, samples, snapshotAt); err != nil {
		return err
	}
	welcomer.Logger.Info().
		Int("sampled_guilds", len(guilds)).
		Dur("total_duration", time.Since(jobStartedAt)).
		Msg("guild analytics benchmark aggregation completed")
	return nil
}

type guildBenchmarkQuerier interface {
	GetGuildAnalyticsBenchmarkInputs(context.Context, database.GetGuildAnalyticsBenchmarkInputsParams) ([]*database.GetGuildAnalyticsBenchmarkInputsRow, error)
	GetActiveUserCounts(context.Context, database.GetActiveUserCountsParams) (*database.GetActiveUserCountsRow, error)
	GetGuildAnalyticsBenchmarkJoinMetrics(context.Context, database.GetGuildAnalyticsBenchmarkJoinMetricsParams) (*database.GetGuildAnalyticsBenchmarkJoinMetricsRow, error)
}

type guildBenchmarkQueryResults struct {
	input          *database.GetGuildAnalyticsBenchmarkInputsRow
	activeUsers    *database.GetActiveUserCountsRow
	joinMetrics    *database.GetGuildAnalyticsBenchmarkJoinMetricsRow
	inputDuration  time.Duration
	activeDuration time.Duration
	joinDuration   time.Duration
}

func fetchGuildBenchmarkQueries(
	ctx context.Context,
	queries guildBenchmarkQuerier,
	guildID int64,
	window welcomer.GuildAnalyticsBenchmarkWindow,
) (guildBenchmarkQueryResults, error) {
	var results guildBenchmarkQueryResults
	var inputRows []*database.GetGuildAnalyticsBenchmarkInputsRow
	var inputErr, activeErr, joinErr error
	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		start := time.Now()
		inputRows, inputErr = queries.GetGuildAnalyticsBenchmarkInputs(ctx, database.GetGuildAnalyticsBenchmarkInputsParams{
			GuildID: guildID, WeekFrom: window.WeekFrom, WeekTo: window.WeekTo,
		})
		results.inputDuration = time.Since(start)
	}()
	go func() {
		defer wg.Done()
		start := time.Now()
		results.activeUsers, activeErr = queries.GetActiveUserCounts(ctx, database.GetActiveUserCountsParams{
			GuildID: guildID, DateFrom: window.WeekFrom, DateTo: window.WeekTo.Add(-time.Microsecond),
		})
		results.activeDuration = time.Since(start)
	}()
	go func() {
		defer wg.Done()
		start := time.Now()
		results.joinMetrics, joinErr = queries.GetGuildAnalyticsBenchmarkJoinMetrics(ctx, database.GetGuildAnalyticsBenchmarkJoinMetricsParams{
			GuildID:                         guildID,
			JoinEventType:                   int32(database.ScienceGuildEventTypeUserJoin),
			CohortFrom:                      window.CohortFrom,
			CohortTo:                        window.CohortTo,
			LeaveEventType:                  int32(database.ScienceGuildEventTypeUserLeave),
			CohortTo30DayspgCataloginterval: window.CohortTo.AddDate(0, 0, 30),
		})
		results.joinDuration = time.Since(start)
	}()
	wg.Wait()

	if inputErr != nil {
		if activeErr != nil || joinErr != nil {
			welcomer.Logger.Warn().
				Int64("guild_id", guildID).
				AnErr("daily_activity_error", activeErr).
				AnErr("join_metrics_error", joinErr).
				Msg("additional benchmark queries failed")
		}
		return results, fmt.Errorf("get activity inputs for guild %d: %w", guildID, inputErr)
	}
	if len(inputRows) != 1 {
		if activeErr != nil || joinErr != nil {
			welcomer.Logger.Warn().
				Int64("guild_id", guildID).
				Err(activeErr).
				AnErr("join_metrics_error", joinErr).
				Msg("speculative benchmark queries failed for guild without an activity row")
		}
		return results, nil
	}
	results.input = inputRows[0]
	if activeErr != nil {
		if joinErr != nil {
			welcomer.Logger.Warn().
				Int64("guild_id", guildID).
				Err(joinErr).
				Msg("additional benchmark join metrics query failed")
		}
		return results, fmt.Errorf("get daily activity for guild %d: %w", guildID, activeErr)
	}
	if joinErr != nil {
		return results, fmt.Errorf("get retention and onboarding metrics for guild %d: %w", guildID, joinErr)
	}
	return results, nil
}

func cachedGuildAverageAccountAgeDays(
	ctx context.Context,
	guildID int64,
	now time.Time,
) (float64, bool, error) {
	response, err := welcomer.SandwichClient.FetchGuildMember(
		ctx,
		&sandwich.FetchGuildMemberRequest{GuildId: guildID},
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

// sampleGuildsPerBucket keeps a uniform reservoir sample independently for each member-count bucket.
func sampleGuildsPerBucket(
	guilds []*database.GetGuildAnalyticsBenchmarkGuildsRow,
	sampleLimit int,
	random *rand.Rand,
) ([]*database.GetGuildAnalyticsBenchmarkGuildsRow, map[int16]int) {
	samples := make(map[int16][]*database.GetGuildAnalyticsBenchmarkGuildsRow)
	eligibleCounts := make(map[int16]int)

	for _, guild := range guilds {
		bucket, ok := welcomer.GuildAnalyticsBenchmarkBucketForMembers(int64(guild.MemberCount))
		if !ok {
			continue
		}

		eligibleCounts[bucket.ID]++
		bucketSample := samples[bucket.ID]
		if len(bucketSample) < sampleLimit {
			samples[bucket.ID] = append(bucketSample, guild)
			continue
		}

		if sampleLimit == 0 {
			continue
		}

		replacementIndex := random.Intn(eligibleCounts[bucket.ID])
		if replacementIndex < sampleLimit {
			bucketSample[replacementIndex] = guild
		}
	}

	selected := make([]*database.GetGuildAnalyticsBenchmarkGuildsRow, 0)
	for _, bucket := range welcomer.GuildAnalyticsBenchmarkBuckets() {
		selected = append(selected, samples[bucket.ID]...)
	}

	return selected, eligibleCounts
}

func addBenchmarkSample(samples map[int16]map[string][]float64, bucketID int16, metricKey string, value float64) {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return
	}
	if samples[bucketID] == nil {
		samples[bucketID] = make(map[string][]float64)
	}
	samples[bucketID][metricKey] = append(samples[bucketID][metricKey], value)
}

func suppressedP90(values []float64) (float64, bool) {
	if len(values) < welcomer.GuildAnalyticsBenchmarkMinimumSampleCount {
		return 0, false
	}

	sorted := append([]float64(nil), values...)
	sort.Float64s(sorted)
	// Nearest-rank p90: ceil(0.9 * n), converted to a zero-based index.
	return sorted[len(sorted)-len(sorted)/10-1], true
}

func persistGuildAnalyticsBenchmarks(
	ctx context.Context,
	samples map[int16]map[string][]float64,
	updatedAt time.Time,
) error {
	persistStartedAt := time.Now()
	welcomer.Logger.Info().
		Time("snapshot_at", updatedAt).
		Msg("starting transactional benchmark snapshot replacement")

	tx, err := welcomer.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin benchmark snapshot transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	queries := database.New(tx)
	if err := queries.DeleteGuildAnalyticsBenchmarks(ctx); err != nil {
		return fmt.Errorf("remove old benchmark snapshot: %w", err)
	}

	for _, bucket := range welcomer.GuildAnalyticsBenchmarkBuckets() {
		publishedMetrics := 0
		suppressedMetrics := 0
		for _, metricKey := range benchmarkMetricKeys {
			values := samples[bucket.ID][metricKey]
			value, available := suppressedP90(values)
			benchmarkValue := sql.NullFloat64{}
			if available {
				benchmarkValue.Float64 = value
				benchmarkValue.Valid = true
				publishedMetrics++
			} else {
				suppressedMetrics++
			}

			if err := queries.UpsertGuildAnalyticsBenchmark(
				ctx,
				database.UpsertGuildAnalyticsBenchmarkParams{
					MemberCountBucket: bucket.ID,
					MetricKey:         metricKey,
					BenchmarkValue:    benchmarkValue,
					SampleCount:       int32(len(values)),
					UpdatedAt:         updatedAt,
				},
			); err != nil {
				return fmt.Errorf("write benchmark %s for bucket %d: %w", metricKey, bucket.ID, err)
			}
		}
		welcomer.Logger.Info().
			Str("bucket", bucket.Label).
			Int("metrics_published", publishedMetrics).
			Int("metrics_suppressed", suppressedMetrics).
			Msg("wrote benchmark metrics for member-count bucket")
	}

	welcomer.Logger.Info().Msg("committing benchmark snapshot")
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit benchmark snapshot: %w", err)
	}

	welcomer.Logger.Info().
		Time("updated_at", updatedAt).
		Dur("persistence_duration", time.Since(persistStartedAt)).
		Msg("stored guild analytics benchmark 90th percentiles")
	return nil
}
