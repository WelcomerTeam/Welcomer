package backend

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"time"

	discord "github.com/WelcomerTeam/Discord/discord"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	"github.com/gin-gonic/gin"
)

var (
	defaultPeriod      = time.Hour * 24 * 7
	ratelimitAnalytics = NewRateLimiterRule(10, time.Second*5)
)

type datePeriod string

const (
	datePeriodHour  datePeriod = "hour"
	datePeriodDay   datePeriod = "day"
	datePeriodMonth datePeriod = "month"
)

func getPreviousDuration(from, to time.Time) (time.Time, time.Time) {
	duration := to.Sub(from)

	return from.Add(-duration), to.Add(-duration)
}

func getGroupingPeriod(from, to time.Time) datePeriod {
	if to.Sub(from) > time.Hour*24*30*6 {
		return datePeriodMonth
	}

	if to.Sub(from) > time.Hour*24*7 {
		return datePeriodDay
	}

	return datePeriodHour
}

type AnalyticsRequest struct {
	From time.Time `form:"from"`
	To   time.Time `form:"to"`

	PreviousFrom time.Time `form:"previous_from"`
	PreviousTo   time.Time `form:"previous_to"`
}

type DatasetItem struct {
	Key   string
	Value int
}

// DatasetItem is represented in json as ["Key", Value] instead of a dictionary.
func (d *DatasetItem) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("[\"%s\", %d]", d.Key, d.Value)), nil
}

func createTimeMap(from, to time.Time, p datePeriod) map[time.Time]int {
	var d time.Duration

	switch p {
	case datePeriodHour:
		d = time.Hour
	case datePeriodDay:
		d = time.Hour * 24
	case datePeriodMonth:
		d = time.Hour * 24
	}

	from = from.UTC().Truncate(d)
	to = to.UTC().Truncate(d).Add(d)

	memberMap := map[time.Time]int{}

	if p == datePeriodMonth {
		from = time.Date(from.Year(), from.Month(), 1, 0, 0, 0, 0, from.Location())
		to = time.Date(to.Year(), to.Month(), 1, 0, 0, 0, 0, to.Location()).AddDate(0, 1, 0).Add(-time.Second)

		for t := from; t.Before(to); t = t.AddDate(0, 1, 0) {
			memberMap[t] = 0
		}

		return memberMap
	}

	for t := from; t.Before(to); t = t.Add(d) {
		memberMap[t] = 0
	}

	return memberMap
}

func TimeMapToDatasetItems(timeMap map[time.Time]int) []DatasetItem {
	dataset := make([]DatasetItem, 0, len(timeMap))

	for k, v := range timeMap {
		dataset = append(dataset, DatasetItem{
			Key:   k.Format(time.RFC3339),
			Value: v,
		})
	}

	sort.Slice(dataset, func(i, j int) bool {
		return dataset[i].Key < dataset[j].Key
	})

	return dataset
}

type guildAnalyticsOverviewResponse struct {
	TotalGuildMembers int           `json:"total_guild_members"`
	MembersJoined     int           `json:"members_joined"`
	MembersLeft       int           `json:"members_left"`
	GuildMembersSum   []DatasetItem `json:"guild_members"`
	GuildNet          []DatasetItem `json:"guild_net"`
	GuildJoins        []DatasetItem `json:"guild_joins"`
	GuildLeaves       []DatasetItem `json:"guild_leaves"`
}

// Route GET /api/guild/:guildID/analytics/overview
func getGuildAnalyticsOverview(ctx *gin.Context) {
	requireOAuthAuthorization(ctx, func(ctx *gin.Context) {
		requireGuildElevation(ctx, func(ctx *gin.Context) {
			guildID := tryGetGuildID(ctx)

			var err error

			ok, err := ratelimitAnalytics.CanRequest(ctx, guildID.String())
			if err != nil {
				welcomer.Logger.Error().Err(err).Int64("guild_id", int64(guildID)).Msg("failed to check rate limit")

				ctx.JSON(http.StatusInternalServerError, NewBaseResponse(NewGenericErrorWithLineNumber(), nil))

				return
			}

			if !ok {
				welcomer.Logger.Info().Int64("guild_id", int64(guildID)).Msg("rate limit exceeded")

				ctx.JSON(http.StatusTooManyRequests, NewBaseResponse(ErrTooManyRequests, nil))

				return
			}

			// TODO: caching
			// TODO: premium date periods

			partial := &AnalyticsRequest{}

			err = ctx.BindQuery(partial)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, NewBaseResponse(NewGenericErrorWithLineNumber(), nil))

				return
			}

			if partial.From.IsZero() || partial.To.IsZero() {
				partial.From = time.Now().Add(-1 * defaultPeriod)
				partial.To = time.Now()
			}

			if partial.PreviousFrom.IsZero() || partial.PreviousTo.IsZero() {
				partial.PreviousFrom, partial.PreviousTo = getPreviousDuration(partial.From, partial.To)
			}

			data, err := getGuildAnalyticsOverviewData(ctx, guildID, partial.From, partial.To, 0)
			if err != nil {
				welcomer.Logger.Error().Err(err).Int64("guild_id", int64(guildID)).Msg("failed to get guild analytics overview data")

				ctx.JSON(http.StatusInternalServerError, NewBaseResponse(NewGenericErrorWithLineNumber(), nil))

				return
			}

			previousData, err := getGuildAnalyticsOverviewData(ctx, guildID, partial.PreviousFrom, partial.PreviousTo, int32(data.GuildMembersSum[0].Value))
			if err != nil {
				welcomer.Logger.Error().Err(err).Int64("guild_id", int64(guildID)).Msg("failed to get previous guild analytics overview data")

				ctx.JSON(http.StatusInternalServerError, NewBaseResponse(NewGenericErrorWithLineNumber(), nil))

				return
			}

			if len(previousData.GuildJoins) != len(data.GuildJoins) {
				welcomer.Logger.Info().Int("expected_length", len(data.GuildJoins)).Int("actual_length", len(previousData.GuildJoins)).Msg("mismatched guild joins length")

				// Keep datapoint lengths equal. Remove first N if not matching
				previousData.GuildJoins = previousData.GuildJoins[len(previousData.GuildJoins)-len(data.GuildJoins):]
				previousData.GuildLeaves = previousData.GuildLeaves[len(previousData.GuildLeaves)-len(data.GuildLeaves):]
				previousData.GuildMembersSum = previousData.GuildMembersSum[len(previousData.GuildMembersSum)-len(data.GuildMembersSum):]
			}

			ctx.JSON(http.StatusOK, BaseResponse{
				Ok: true,
				Data: map[string]any{
					"current":  data,
					"previous": previousData,
				},
			})
		})
	})
}

func getGuildAnalyticsOverviewData(ctx context.Context, guildID discord.Snowflake, from, to time.Time, overrideGuildMemberCount int32) (*guildAnalyticsOverviewResponse, error) {
	guildMemberCount := overrideGuildMemberCount

	groupingPeriod := getGroupingPeriod(from, to)

	if guildMemberCount == 0 {
		guild, err := welcomer.Queries.GetGuild(ctx, int64(guildID))
		if err != nil {
			return nil, fmt.Errorf("failed to get guild: %w", err)
		}

		guildMemberCount = guild.MemberCount
	}

	memberJoinEvents, err := welcomer.Queries.GetScienceGuildEventsForGuildGroupedByPeriod(ctx, database.GetScienceGuildEventsForGuildGroupedByPeriodParams{
		GuildID:   int64(guildID),
		EventType: int32(database.ScienceGuildEventTypeUserJoin),
		DateFrom:  from,
		DateTo:    to,
		Period:    string(groupingPeriod),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get member join events: %w", err)
	}

	memberLeftEvents, err := welcomer.Queries.GetScienceGuildEventsForGuildGroupedByPeriod(ctx, database.GetScienceGuildEventsForGuildGroupedByPeriodParams{
		GuildID:   int64(guildID),
		EventType: int32(database.ScienceGuildEventTypeUserLeave),
		DateFrom:  from,
		DateTo:    to,
		Period:    string(groupingPeriod),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get member left events: %w", err)
	}

	sumJoins := 0
	sumLeaves := 0

	memberJoins := createTimeMap(from, to, groupingPeriod)
	memberLeaves := createTimeMap(from, to, groupingPeriod)

	emptyTimeMap := createTimeMap(from, to, groupingPeriod)

	for _, e := range memberJoinEvents {
		memberJoins[e.Date] = int(e.EventCount)
		sumJoins += int(e.EventCount)
	}

	for _, e := range memberLeftEvents {
		memberLeaves[e.Date] = -int(e.EventCount)
		sumLeaves += int(e.EventCount)
	}

	memberCount := int(guildMemberCount) - sumJoins + sumLeaves

	guildMembersDataset := TimeMapToDatasetItems(emptyTimeMap)
	memberNetDataset := TimeMapToDatasetItems(emptyTimeMap)
	memberJoinsDataset := TimeMapToDatasetItems(memberJoins)
	memberLeavesDataset := TimeMapToDatasetItems(memberLeaves)

	for i := range guildMembersDataset {
		memberCount = memberCount + memberJoinsDataset[i].Value + memberLeavesDataset[i].Value
		guildMembersDataset[i].Value = memberCount
		memberNetDataset[i].Value = memberJoinsDataset[i].Value + memberLeavesDataset[i].Value
	}

	return &guildAnalyticsOverviewResponse{
		TotalGuildMembers: int(guildMemberCount),
		MembersJoined:     sumJoins,
		MembersLeft:       sumLeaves,
		GuildMembersSum:   guildMembersDataset,
		GuildNet:          memberNetDataset,
		GuildJoins:        memberJoinsDataset,
		GuildLeaves:       memberLeavesDataset,
	}, nil
}

func registerGuildAnalyticsRoutes(g *gin.Engine) {
	g.GET("/api/guild/:guildID/analytics/overview", getGuildAnalyticsOverview)
}

// Page Layouts

// Growth:

// Quick summaries:
// - Member joined percentage less than or more than 10%
// - Member left percentage less than or more than 10%
// - Retention percentage less than or more than 10%

// |-----------------------------------------------|
// | Guild Members | Members Left | Members Joined |
// | x             | 10           | 10             |
// |-----------------------------------------------|
// |Line chart of guild members (date filter)      |
// |-----------------------------------------------||
// |Line chart with net guild members (date filter) |
// |Overlay with bar chart for members joined (up)  |
// |Overlay with bar charts for members left (down) |
// |------------------------------------------------|

// Retention:

// | Members Left | Members Joined | Retention     |
// | x            | 10             | 10%           |
// |-----------------------------------------------|
// | Retention Cohorts                             |
// | x Joined 12 months ago .. joined this month   |
// | Left 12 months ago .. left this month         |
// |-----------------------------------------------|
// | Histogram showing how long members remain     |
// | before leaving (date filter)                  |
// |-----------------------------------------------|
// | Average messages sent before leaving (date filter) |
// |-----------------------------------------------|

// User Activity:

// | Daily Active | Weekly Active | Monthly Active |
// | x             | 10            | 10            |
// |-----------------------------------------------|
// | Unique Users (sent in time period)            |
// |-----------------------------------------------|

// |-----------------------------------------------|
// | Total Members | Active Members | Inactive Members | No Messages Sent |
// | x             | 10             | 10               | 10               |
// |-----------------------------------------------|
// | Active/Inactive/NoMsg account age distribution (date filter) |
// |-----------------------------------------------|

// |---------------------------------------------------------|
// | Time to First Message | Average | Median | Distribution |
// | x                     | 10      | 10     | 10           |
// |---------------------------------------------------------|

// Message Activity:

// |---------------|--------------|-------------------|
// | Messages Sent | Unique Users | Messages per User |
// | x             | 10           | 10                |
// |---------------|--------------|-------------------|

// |-----------------------------------------------------------|
// |Line chart of messages sent over time by day (date filter) |
// | (Show Hourly) limit to a week                             |
// |-----------------------------------------------------------|

// |-------------------------------------------------------|
// |Line chart of messages by channel by day (date filter) |
// |-------------------------------------------------------|

// |---------------------------------------------------------------|
// |Heatmap of messages sent by hour (globally or channels filter) |
// |---------------------------------------------------------------|

// |---------------------------------------------------------------|
// |Histogram of message counts by user (date filter)              |
// |---------------------------------------------------------------|

// |---------------------------------------------------|
// |Leaderboard of messages sent by user (date filter) |
// |---------------------------------------------------|

// |---------------------------------------------------------------------------------------|
// |Leaderboard of top contributors (1 tick per hour instead of per message) (date filter) |
// |---------------------------------------------------------------------------------------|

// Voice Activity:

// |--------------------------------------------------------------------------------|
// | Voice Session Count | Unique Users | Total Voice Time | Average Session Length |
// |--------------------------------------------------------------------------------|

// |--------------------------------------------------|
// |Line chart of voice sessions by day (date filter) |
// |--------------------------------------------------|

// |-----------------------------------------------------|
// |Line chart of time and duration by day (date filter) |
// |-----------------------------------------------------|

// |----------------------------------------------------------------|
// |Line chart of simultaneous users in voice by hour (date filter) |
// |----------------------------------------------------------------|

// |----------------------------------------------------------------|
// |Heatmap of voice activity by hour (globally or channels filter) |
// |----------------------------------------------------------------|

// |------------------------------------------------------------|
// |Leaderboard of voice time by user (date filter)             |
// |------------------------------------------------------------|

// |------------------------------------------------------------|
// |Leaderboard of voice session count by user (date filter)    |
// |------------------------------------------------------------|

// Demographics

// [Filter against current members only, or include former members.]

// |--------------------|--------------------|-----------------------------|
// |Average Account Age | Median Account Age | 90th Percentile Account Age |
// |--------------------|--------------------|-----------------------------|

// |----------------------------------------------------|
// |Histogram of account age distribution (date filter) |
// |----------------------------------------------------|

// |------------------------------------------------------------|
// |Histogram of account age distribution by role (date filter) |
// |------------------------------------------------------------|

// |------------------------------------------------------------|
// |Scatter plot of account age vs join date (date filter)      |
// |------------------------------------------------------------|

// Role Activity:

// |----------------------------------------|
// |Line chart of current role distribution |
// |----------------------------------------|

// |--------------------------------------------------------------------------------|
// |Line chart of invoked modules (total/auto/free/rr/time) over time (date filter) |
// |--------------------------------------------------------------------------------|

// |------------------------------------------------------|
// |Line chart of autoroles given over time (date filter) |
// |------------------------------------------------------|

// |------------------------------------------------------|
// |Line chart of freeroles given over time (date filter) |
// |------------------------------------------------------|

// |-----------------------------------------------------------|
// |Line chart of reaction roles given over time (date filter) |
// |-----------------------------------------------------------|

// |------------------------------------------------------|
// |Line chart of timeroles given over time (date filter) |
// |------------------------------------------------------|

// # Borderwall

// |------------------------------------------------|------------------------|
// | Total Requests | Verified Requests | Pass Rate | Average Time to Verify |
// | x              | 10                | 10%       | 10s                    |
// |-------------------------------------------------------------------------|

// |------------------------------------------------------------|
// |Line chart of requests and verified over time (date filter) |
// |------------------------------------------------------------|

// |------------------------------------------------|
// |Pie chart of connecting countries (date filter) |
// |------------------------------------------------|

// |------------------------------------------------|
// |Pie chart of device types (date filter)         |
// |------------------------------------------------|

// Invites

// |---------------------------------------------------| (inactive = no uses in 30 days)
// | Total Invites | Active Invites | Inactive Invites |
// | x             | 10             | 10               |
// |---------------------------------------------------|

// |----------------------------------------------------------| (date range)
// | Invite Joins | Invite Created | Average Joins per Invite |
// | x            | 10             | 10                       |
// |----------------------------------------------------------|

// |---------------------------------------------------------------------------| (date range)
// |Line chart of member joins and if it is tracked to an invite (date filter) |
// |---------------------------------------------------------------------------|

// |-------------------------------|
// |Leaderboard of invites by user |
// |-------------------------------|

// |---------------------------------|
// |Leaderboard of invites by invite |
// |---------------------------------|

// # Server Health

// ### Engagement Score
// Composite metric based on:
// - Active members
// - Messages
// - Voice activity
// - Retention

// ### Silent Members
// Count of members who:
// - Never messaged
// - Not active in 30 days
// - Not active in 90 days

// ### Member Lifetime
// Display:
// - Average
// - Median
// - 90th percentile

// ---

// # Filters

// Support filtering by:

// - Date range
// - Current members only
// - Include former members
// - Role
// - Channel
// - Join date
// - Account age
// - Minimum activity

// ---

// # Future Data Collection

// If additional data is collected in the future, add support for:

// - Message reactions
// - Message length
// - Thread activity
// - Invite usage
// - Role history
// - First voice session
// - Presence statistics

// These enable richer engagement, onboarding and conversion analytics.
