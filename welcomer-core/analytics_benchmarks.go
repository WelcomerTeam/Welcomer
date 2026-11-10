package welcomer

import "time"

const (
	GuildAnalyticsBenchmarkMinimumSampleCount = 30

	GuildAnalyticsBenchmarkMetricRetentionRate30D                 = "retention_rate_30d"
	GuildAnalyticsBenchmarkMetricDailyActiveUsers                 = "daily_active_users"
	GuildAnalyticsBenchmarkMetricWeeklyUniqueMessagingUsers       = "weekly_unique_messaging_users"
	GuildAnalyticsBenchmarkMetricWeeklyMessagesPerUser            = "weekly_messages_per_user"
	GuildAnalyticsBenchmarkMetricWeeklyUniqueVoiceUsers           = "weekly_unique_voice_users"
	GuildAnalyticsBenchmarkMetricAverageVoiceSessionMS            = "average_voice_session_ms"
	GuildAnalyticsBenchmarkMetricActiveMembersPercentage          = "active_members_percentage"
	GuildAnalyticsBenchmarkMetricAverageTimeToFirstMessageSeconds = "average_time_to_first_message_seconds"
	GuildAnalyticsBenchmarkMetricAverageAccountAgeDays            = "average_account_age_days"
)

// GuildAnalyticsBenchmarkWindow contains the fixed periods shared by the job and API.
type GuildAnalyticsBenchmarkWindow struct {
	WeekFrom   time.Time
	WeekTo     time.Time
	CohortFrom time.Time
	CohortTo   time.Time
}

type GuildAnalyticsBenchmarkBucket struct {
	ID         int16
	Label      string
	MinMembers int64
	MaxMembers int64
}

var guildAnalyticsBenchmarkBuckets = [...]GuildAnalyticsBenchmarkBucket{
	{ID: 0, Label: "0–99", MinMembers: 0, MaxMembers: 100},
	{ID: 1, Label: "100–499", MinMembers: 100, MaxMembers: 500},
	{ID: 2, Label: "500–1,499", MinMembers: 500, MaxMembers: 1500},
	{ID: 3, Label: "1,500–3,499", MinMembers: 1500, MaxMembers: 3500},
	{ID: 4, Label: "3,500–9,999", MinMembers: 3500, MaxMembers: 10000},
	{ID: 5, Label: "10,000–99,999", MinMembers: 10000, MaxMembers: 100000},
	{ID: 6, Label: "100,000+", MinMembers: 100000},
}

// GuildAnalyticsBenchmarkWindowFor uses seven complete UTC activity days and a matured
// join cohort from 60 to 30 days ago, observed for an equal 30-day follow-up.
func GuildAnalyticsBenchmarkWindowFor(now time.Time) GuildAnalyticsBenchmarkWindow {
	today := time.Date(now.UTC().Year(), now.UTC().Month(), now.UTC().Day(), 0, 0, 0, 0, time.UTC)

	return GuildAnalyticsBenchmarkWindow{
		WeekFrom:   today.AddDate(0, 0, -7),
		WeekTo:     today,
		CohortFrom: today.AddDate(0, 0, -60),
		CohortTo:   today.AddDate(0, 0, -30),
	}
}

// GuildAnalyticsBenchmarkBuckets returns the fixed, lower-inclusive member-count buckets.
func GuildAnalyticsBenchmarkBuckets() []GuildAnalyticsBenchmarkBucket {
	return append([]GuildAnalyticsBenchmarkBucket(nil), guildAnalyticsBenchmarkBuckets[:]...)
}

// GuildAnalyticsBenchmarkBucketForMembers maps a guild size to its benchmark bucket.
func GuildAnalyticsBenchmarkBucketForMembers(memberCount int64) (GuildAnalyticsBenchmarkBucket, bool) {
	for _, bucket := range guildAnalyticsBenchmarkBuckets {
		if memberCount >= bucket.MinMembers && (bucket.MaxMembers == 0 || memberCount < bucket.MaxMembers) {
			return bucket, true
		}
	}

	return GuildAnalyticsBenchmarkBucket{}, false
}

// DiscordSnowflakeAccountCreatedAt extracts the creation timestamp encoded in a user ID.
func DiscordSnowflakeAccountCreatedAt(userID int64) time.Time {
	const discordEpochMilliseconds = int64(1420070400000)

	return time.UnixMilli(int64(uint64(userID)>>22) + discordEpochMilliseconds)
}

// AverageDiscordAccountAgeDays derives current-member account ages from Discord snowflake IDs.
// Invalid and future-dated IDs are ignored; false means there were no usable IDs.
func AverageDiscordAccountAgeDays(userIDs []int64, now time.Time) (float64, bool) {
	var totalAgeDays float64
	var validUsers int
	for _, userID := range userIDs {
		if userID <= 0 {
			continue
		}

		createdAt := DiscordSnowflakeAccountCreatedAt(userID)
		if createdAt.After(now) {
			continue
		}

		totalAgeDays += now.Sub(createdAt).Hours() / 24
		validUsers++
	}

	if validUsers == 0 {
		return 0, false
	}

	return totalAgeDays / float64(validUsers), true
}
