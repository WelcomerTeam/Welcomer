package backend

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	discord "github.com/WelcomerTeam/Discord/discord"
	sandwich "github.com/WelcomerTeam/Sandwich-Daemon/proto"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	"github.com/gin-gonic/gin"
)

func withGuildAnalyticsRequest(
	ctx *gin.Context,
	handler func(*gin.Context, discord.Snowflake, *AnalyticsRequest),
) {
	requireOAuthAuthorization(ctx, func(ctx *gin.Context) {
		requireGuildElevation(ctx, func(ctx *gin.Context) {
			guildID := tryGetGuildID(ctx)

			ok, err := ratelimitAnalytics.CanRequest(ctx, guildID.String())
			if err != nil {
				welcomer.Logger.Error().
					Err(err).
					Int64("guild_id", int64(guildID)).
					Msg("failed to check rate limit")
				ctx.JSON(
					http.StatusInternalServerError,
					NewBaseResponse(NewGenericErrorWithLineNumber(), nil),
				)

				return
			}

			if !ok {
				ctx.JSON(http.StatusTooManyRequests, NewBaseResponse(ErrTooManyRequests, nil))
				return
			}

			request := &AnalyticsRequest{}
			if err := ctx.BindQuery(request); err != nil {
				ctx.JSON(
					http.StatusBadRequest,
					NewBaseResponse(NewGenericErrorWithLineNumber(), nil),
				)

				return
			}

			if request.From.IsZero() || request.To.IsZero() {
				request.To = time.Now()
				request.From = request.To.Add(-defaultPeriod)
			}

			if !request.From.Before(request.To) {
				ctx.JSON(
					http.StatusBadRequest,
					NewBaseResponse(NewGenericErrorWithLineNumber(), nil),
				)

				return
			}

			if request.PreviousFrom.IsZero() || request.PreviousTo.IsZero() {
				request.PreviousFrom, request.PreviousTo = getPreviousDuration(
					request.From,
					request.To,
				)
			}

			handler(ctx, guildID, request)
		})
	})
}

func getMessageActivityData(
	ctx *gin.Context,
	guildID discord.Snowflake,
	from, to time.Time,
) (map[string]any, error) {
	period := string(getGroupingPeriod(from, to))
	args := database.GetGuildMessageActivityParams{
		Period: period, GuildID: int64(guildID), DateFrom: from, DateTo: to,
	}

	timeline, err := welcomer.Queries.GetGuildMessageActivity(ctx, args)
	if err != nil {
		return nil, fmt.Errorf("get message activity: %w", err)
	}

	summary, err := welcomer.Queries.GetGuildMessageSummary(
		ctx,
		database.GetGuildMessageSummaryParams{
			GuildID: int64(guildID), DateFrom: from, DateTo: to,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("get message summary: %w", err)
	}

	channels, err := welcomer.Queries.GetGuildMessageChannels(
		ctx,
		database.GetGuildMessageChannelsParams{
			GuildID: int64(guildID), DateFrom: from, DateTo: to,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("get message channels: %w", err)
	}

	authors, err := welcomer.Queries.GetGuildMessageAuthors(
		ctx,
		database.GetGuildMessageAuthorsParams{
			GuildID: int64(guildID), DateFrom: from, DateTo: to,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("get message authors: %w", err)
	}

	contributors, err := welcomer.Queries.GetGuildMessageContributors(
		ctx,
		database.GetGuildMessageContributorsParams{
			GuildID: int64(guildID), DateFrom: from, DateTo: to,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("get message contributors: %w", err)
	}

	userIDs := make([]int64, 0, len(authors)+len(contributors))

	userIDsSeen := make(map[int64]struct{}, len(authors)+len(contributors))
	for _, row := range authors {
		var userID int64
		if _, scanErr := fmt.Sscan(row.Key, &userID); scanErr != nil {
			return nil, fmt.Errorf("parse message author ID %q: %w", row.Key, scanErr)
		}

		if _, exists := userIDsSeen[userID]; !exists {
			userIDsSeen[userID] = struct{}{}
			userIDs = append(userIDs, userID)
		}
	}

	for _, row := range contributors {
		var userID int64
		if _, scanErr := fmt.Sscan(row.Key, &userID); scanErr != nil {
			return nil, fmt.Errorf("parse message contributor ID %q: %w", row.Key, scanErr)
		}

		if _, exists := userIDsSeen[userID]; !exists {
			userIDsSeen[userID] = struct{}{}
			userIDs = append(userIDs, userID)
		}
	}

	userNames, err := fetchAnalyticsUserNames(ctx, guildID, userIDs)
	if err != nil {
		return nil, fmt.Errorf("get message analytics user names: %w", err)
	}

	channelResponse, err := welcomer.SandwichClient.FetchGuildChannel(
		ctx,
		&sandwich.FetchGuildChannelRequest{GuildId: int64(guildID)},
	)
	if err != nil {
		return nil, fmt.Errorf("get guild channels for analytics: %w", err)
	}

	channelNames := make(map[int64]string, len(channelResponse.GetChannels()))
	for channelID, channel := range channelResponse.GetChannels() {
		if channel.GetName() != "" {
			channelNames[channelID] = "#" + channel.GetName()
		}
	}

	heatmap, err := welcomer.Queries.GetGuildMessageHeatmap(
		ctx,
		database.GetGuildMessageHeatmapParams{
			GuildID: int64(guildID), DateFrom: from, DateTo: to,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("get message heatmap: %w", err)
	}

	totalMessages, uniqueUsers := summary.TotalMessages, summary.UniqueUsers
	messageSeries := make([]DatasetItem, 0, len(timeline))

	uniqueSeries := make([]DatasetItem, 0, len(timeline))
	for _, row := range timeline {
		messageSeries = append(
			messageSeries,
			DatasetItem{Key: row.Date.Format(time.RFC3339), Value: int(row.MessageCount)},
		)
		uniqueSeries = append(
			uniqueSeries,
			DatasetItem{Key: row.Date.Format(time.RFC3339), Value: int(row.UniqueUsers)},
		)
	}

	toDataset := func(rows []*database.GetGuildMessageChannelsRow) []DatasetItem {
		result := make([]DatasetItem, 0, len(rows))
		for _, row := range rows {
			var channelID int64
			if _, scanErr := fmt.Sscan(fmt.Sprint(row.Key), &channelID); scanErr != nil {
				continue
			}

			name := channelNames[channelID]
			if name == "" {
				name = "Deleted channel"
			}

			result = append(result, DatasetItem{Key: name, Value: int(row.Value)})
		}

		return result
	}

	authorDataset := make([]DatasetItem, 0, len(authors))
	for _, row := range authors {
		authorDataset = append(
			authorDataset,
			DatasetItem{Key: userNames[row.Key], Value: int(row.Value)},
		)
	}

	contributorDataset := make([]DatasetItem, 0, len(contributors))
	for _, row := range contributors {
		contributorDataset = append(
			contributorDataset,
			DatasetItem{Key: userNames[row.Key], Value: int(row.Value)},
		)
	}

	return map[string]any{
		"total_messages": totalMessages,
		"unique_users":   uniqueUsers,
		"messages_per_user": func() int64 {
			if uniqueUsers == 0 {
				return 0
			}

			return totalMessages / uniqueUsers
		}(),
		"messages_by_time":     messageSeries,
		"unique_users_by_time": uniqueSeries,
		"messages_by_channel":  toDataset(channels),
		"top_authors":          authorDataset,
		"top_contributors":     contributorDataset,
		"message_heatmap":      heatmap,
	}, nil
}

func fetchAnalyticsUserNames(
	ctx *gin.Context,
	guildID discord.Snowflake,
	userIDs []int64,
) (map[string]string, error) {
	names := make(map[string]string, len(userIDs))
	if len(userIDs) == 0 {
		return names, nil
	}

	// Request guild chunk to fetch user information for the given user IDs.
	_, err := welcomer.SandwichClient.RequestGuildChunk(ctx, &sandwich.RequestGuildChunkRequest{
		GuildId:     int64(guildID),
		AlwaysChunk: false,
	})
	if err != nil {
		return nil, err
	}

	response, err := welcomer.SandwichClient.FetchGuildMember(
		ctx,
		&sandwich.FetchGuildMemberRequest{
			GuildId: int64(guildID),
			UserIds: userIDs,
		},
	)
	if err != nil {
		return nil, err
	}

	guildMembers := response.GetGuildMembers()

	for _, userID := range userIDs {
		guildMember, _ := guildMembers[userID]

		name := getGuildMemberDisplayName(userID, guildMember)
		if name == "" {
			name = fmt.Sprintf("Unknown user (ID: %d)", userID)
		}

		names[fmt.Sprint(userID)] = name
	}

	return names, nil
}

func getGuildMemberDisplayName(userID int64, user *sandwich.GuildMember) string {
	name := ""

	if user != nil {
		name = user.Nick
		if name == "" {
			name = user.GetUser().GlobalName
		}

		if name == "" && user.GetUser().Username != "" {
			if user.GetUser().Discriminator != "0" {
				name = user.GetUser().Username + "#" + user.GetUser().Discriminator
			} else {
				name = user.GetUser().Username
			}
		}
	}

	return name
}

func discordAccountCreatedAt(userID int64) time.Time {
	return time.UnixMilli(int64(uint64(userID)>>22) + 1420070400000)
}

func getGuildAnalyticsMessageActivity(ctx *gin.Context) {
	withGuildAnalyticsRequest(
		ctx,
		func(ctx *gin.Context, guildID discord.Snowflake, request *AnalyticsRequest) {
			current, err := getMessageActivityData(ctx, guildID, request.From, request.To)
			if err == nil {
				var previous map[string]any

				previous, err = getMessageActivityData(
					ctx,
					guildID,
					request.PreviousFrom,
					request.PreviousTo,
				)
				if err == nil {
					ctx.JSON(
						http.StatusOK,
						BaseResponse{
							Ok:   true,
							Data: map[string]any{"current": current, "previous": previous},
						},
					)

					return
				}
			}

			welcomer.Logger.Error().
				Err(err).
				Int64("guild_id", int64(guildID)).
				Msg("failed to get guild message analytics")
			ctx.JSON(
				http.StatusInternalServerError,
				NewBaseResponse(NewGenericErrorWithLineNumber(), nil),
			)
		},
	)
}

func getVoiceActivityData(
	ctx *gin.Context,
	guildID discord.Snowflake,
	from, to time.Time,
) (map[string]any, error) {
	period := string(getGroupingPeriod(from, to))
	fromTime := time.Now()
	timeline, err := welcomer.Queries.GetGuildVoiceActivity(
		ctx,
		database.GetGuildVoiceActivityParams{
			Period: period, GuildID: int64(guildID), DateFrom: from, DateTo: to,
		},
	)

	println(time.Since(fromTime).String())

	if err != nil {
		return nil, fmt.Errorf("get voice activity: %w", err)
	}

	authors, err := welcomer.Queries.GetGuildVoiceAuthors(ctx, database.GetGuildVoiceAuthorsParams{
		GuildID: int64(guildID), DateFrom: from, DateTo: to,
	})

	println(time.Since(fromTime).String())

	if err != nil {
		return nil, fmt.Errorf("get voice authors: %w", err)
	}

	summary, err := welcomer.Queries.GetGuildVoiceSummary(ctx, database.GetGuildVoiceSummaryParams{
		GuildID: int64(guildID), DateFrom: from, DateTo: to,
	})
	if err != nil {
		return nil, fmt.Errorf("get voice summary: %w", err)
	}

	userIDs := make([]int64, 0, len(authors))
	for _, row := range authors {
		var userID int64
		if _, scanErr := fmt.Sscan(row.Key, &userID); scanErr != nil {
			return nil, fmt.Errorf("parse voice user ID %q: %w", row.Key, scanErr)
		}

		userIDs = append(userIDs, userID)
	}

	userNames, err := fetchAnalyticsUserNames(ctx, guildID, userIDs)
	if err != nil {
		return nil, fmt.Errorf("get voice analytics user names: %w", err)
	}

	sessions, uniqueUsers, totalTime := summary.Sessions, summary.UniqueUsers, summary.TotalTimeMs
	sessionSeries := make([]DatasetItem, 0, len(timeline))

	timeSeries := make([]DatasetItem, 0, len(timeline))
	for _, row := range timeline {
		key := row.Date.Format(time.RFC3339)
		sessionSeries = append(sessionSeries, DatasetItem{Key: key, Value: int(row.Sessions)})
		timeSeries = append(timeSeries, DatasetItem{Key: key, Value: int(row.TotalTimeMs / 1000)})
	}

	topUsers := make([]DatasetItem, 0, len(authors))
	for _, row := range authors {
		topUsers = append(
			topUsers,
			DatasetItem{Key: userNames[row.Key], Value: int(row.Value / 1000)},
		)
	}

	averageSession := int64(0)
	if sessions > 0 {
		averageSession = totalTime / sessions
	}

	return map[string]any{
		"voice_sessions":         sessions,
		"unique_users":           uniqueUsers,
		"total_voice_time_ms":    totalTime,
		"average_session_ms":     averageSession,
		"voice_sessions_by_time": sessionSeries,
		"voice_time_by_time":     timeSeries,
		"top_users":              topUsers,
	}, nil
}

func getVoiceActivitySummary(
	ctx *gin.Context,
	guildID discord.Snowflake,
	from, to time.Time,
) (map[string]any, error) {
	summary, err := welcomer.Queries.GetGuildVoiceSummary(ctx, database.GetGuildVoiceSummaryParams{
		GuildID: int64(guildID), DateFrom: from, DateTo: to,
	})
	if err != nil {
		return nil, fmt.Errorf("get voice summary: %w", err)
	}

	averageSession := int64(0)
	if summary.Sessions > 0 {
		averageSession = summary.TotalTimeMs / summary.Sessions
	}

	return map[string]any{
		"voice_sessions":      summary.Sessions,
		"unique_users":        summary.UniqueUsers,
		"total_voice_time_ms": summary.TotalTimeMs,
		"average_session_ms":  averageSession,
	}, nil
}

func getGuildAnalyticsVoiceActivity(ctx *gin.Context) {
	withGuildAnalyticsRequest(
		ctx,
		func(ctx *gin.Context, guildID discord.Snowflake, request *AnalyticsRequest) {
			current, err := getVoiceActivityData(ctx, guildID, request.From, request.To)
			if err == nil {
				var previous map[string]any

				previous, err = getVoiceActivitySummary(
					ctx,
					guildID,
					request.PreviousFrom,
					request.PreviousTo,
				)
				if err == nil {
					ctx.JSON(
						http.StatusOK,
						BaseResponse{
							Ok:   true,
							Data: map[string]any{"current": current, "previous": previous},
						},
					)

					return
				}
			}

			welcomer.Logger.Error().
				Err(err).
				Int64("guild_id", int64(guildID)).
				Msg("failed to get guild voice analytics")
			ctx.JSON(
				http.StatusInternalServerError,
				NewBaseResponse(NewGenericErrorWithLineNumber(), nil),
			)
		},
	)
}

func getGuildAnalyticsInvites(ctx *gin.Context) {
	withGuildAnalyticsRequest(
		ctx,
		func(ctx *gin.Context, guildID discord.Snowflake, request *AnalyticsRequest) {
			build := func(from, to time.Time) (map[string]any, error) {
				invites, err := welcomer.Queries.GetGuildInvites(ctx, int64(guildID))
				if err != nil {
					return nil, fmt.Errorf("get guild invites: %w", err)
				}

				activity, err := welcomer.Queries.GetGuildInviteActivity(
					ctx,
					database.GetGuildInviteActivityParams{
						Period:  string(getGroupingPeriod(from, to)),
						GuildID: int64(guildID),
						EventType: int32(
							database.ScienceGuildEventTypeUserJoin,
						),
						DateFrom: from,
						DateTo:   to,
					},
				)
				if err != nil {
					return nil, fmt.Errorf("get invite activity: %w", err)
				}

				joins, err := welcomer.Queries.GetGuildInviteJoinCounts(
					ctx,
					database.GetGuildInviteJoinCountsParams{
						GuildID: int64(
							guildID,
						),
						EventType: int32(database.ScienceGuildEventTypeUserJoin),
						DateFrom:  from,
						DateTo:    to,
					},
				)
				if err != nil {
					return nil, fmt.Errorf("get invite joins: %w", err)
				}

				var inviteJoins, trackedJoins int64

				joinSeries := make([]DatasetItem, 0, len(activity))

				trackedSeries := make([]DatasetItem, 0, len(activity))
				for _, row := range activity {
					inviteJoins += row.Joins
					key := row.Date.Format(time.RFC3339)
					joinSeries = append(joinSeries, DatasetItem{Key: key, Value: int(row.Joins)})
					trackedSeries = append(
						trackedSeries,
						DatasetItem{Key: key, Value: int(row.TrackedJoins)},
					)
				}

				inviteDataset := make([]DatasetItem, 0, len(joins))
				creatorCounts := make(map[int64]int64)

				for _, row := range joins {
					inviteDataset = append(
						inviteDataset,
						DatasetItem{Key: fmt.Sprint(row.Key), Value: int(row.Joins)},
					)
					creatorCounts[row.CreatedBy] += row.Joins
					trackedJoins += row.Joins
				}

				creatorDataset := make([]DatasetItem, 0, len(creatorCounts))
				for creator, count := range creatorCounts {
					creatorDataset = append(
						creatorDataset,
						DatasetItem{Key: fmt.Sprint(creator), Value: int(count)},
					)
				}

				sort.Slice(
					creatorDataset,
					func(i, j int) bool { return creatorDataset[i].Value > creatorDataset[j].Value },
				)

				if len(creatorDataset) > 10 {
					creatorDataset = creatorDataset[:10]
				}

				creatorIDs := make([]int64, 0, len(creatorDataset))
				for _, creator := range creatorDataset {
					var creatorID int64
					if _, scanErr := fmt.Sscan(creator.Key, &creatorID); scanErr != nil {
						return nil, fmt.Errorf(
							"parse invite creator ID %q: %w",
							creator.Key,
							scanErr,
						)
					}

					if creatorID != 0 {
						creatorIDs = append(creatorIDs, creatorID)
					}
				}

				creatorNames, err := fetchAnalyticsUserNames(ctx, guildID, creatorIDs)
				if err != nil {
					return nil, fmt.Errorf("get invite creator names: %w", err)
				}

				for index := range creatorDataset {
					var creatorID int64
					if _, scanErr := fmt.Sscan(
						creatorDataset[index].Key,
						&creatorID,
					); scanErr != nil {
						return nil, fmt.Errorf(
							"parse invite creator ID %q: %w",
							creatorDataset[index].Key,
							scanErr,
						)
					}

					if creatorID == 0 {
						creatorDataset[index].Key = "Unknown creator (ID: 0)"
					} else {
						creatorDataset[index].Key = creatorNames[fmt.Sprint(creatorID)]
					}
				}

				return map[string]any{
					"total_invites":   len(invites),
					"invite_joins":    inviteJoins,
					"tracked_joins":   trackedJoins,
					"untracked_joins": inviteJoins - trackedJoins,
					"average_joins_per_invite": func() float64 {
						if len(invites) == 0 {
							return 0
						}

						return float64(trackedJoins) / float64(len(invites))
					}(),
					"joins_by_time":         joinSeries,
					"tracked_joins_by_time": trackedSeries,
					"top_invites":           inviteDataset,
					"top_creators":          creatorDataset,
				}, nil
			}

			current, err := build(request.From, request.To)
			if err == nil {
				var previous map[string]any

				previous, err = build(request.PreviousFrom, request.PreviousTo)
				if err == nil {
					ctx.JSON(
						http.StatusOK,
						BaseResponse{
							Ok:   true,
							Data: map[string]any{"current": current, "previous": previous},
						},
					)

					return
				}
			}

			welcomer.Logger.Error().
				Err(err).
				Int64("guild_id", int64(guildID)).
				Msg("failed to get guild invite analytics")
			ctx.JSON(
				http.StatusInternalServerError,
				NewBaseResponse(NewGenericErrorWithLineNumber(), nil),
			)
		},
	)
}

func getGuildAnalyticsBorderwall(ctx *gin.Context) {
	withGuildAnalyticsRequest(
		ctx,
		func(ctx *gin.Context, guildID discord.Snowflake, request *AnalyticsRequest) {
			build := func(from, to time.Time) (map[string]any, error) {
				args := database.GetGuildBorderwallSummaryParams{
					GuildID:  int64(guildID),
					DateFrom: from,
					DateTo:   to,
				}

				summary, err := welcomer.Queries.GetGuildBorderwallSummary(ctx, args)
				if err != nil {
					return nil, fmt.Errorf("get borderwall summary: %w", err)
				}

				activity, err := welcomer.Queries.GetGuildBorderwallActivity(
					ctx,
					database.GetGuildBorderwallActivityParams{
						Period: string(
							getGroupingPeriod(from, to),
						),
						GuildID:  int64(guildID),
						DateFrom: from,
						DateTo:   to,
					},
				)
				if err != nil {
					return nil, fmt.Errorf("get borderwall activity: %w", err)
				}

				countries, err := welcomer.Queries.GetGuildBorderwallCountries(
					ctx,
					database.GetGuildBorderwallCountriesParams(args),
				)
				if err != nil {
					return nil, fmt.Errorf("get borderwall countries: %w", err)
				}

				devices, err := welcomer.Queries.GetGuildBorderwallDevices(
					ctx,
					database.GetGuildBorderwallDevicesParams(args),
				)
				if err != nil {
					return nil, fmt.Errorf("get borderwall devices: %w", err)
				}

				requestSeries := make([]DatasetItem, 0, len(activity))

				verifiedSeries := make([]DatasetItem, 0, len(activity))
				for _, row := range activity {
					key := row.Date.Format(time.RFC3339)
					requestSeries = append(
						requestSeries,
						DatasetItem{Key: key, Value: int(row.Requests)},
					)
					verifiedSeries = append(
						verifiedSeries,
						DatasetItem{Key: key, Value: int(row.Verified)},
					)
				}

				toDataset := func(rows []*database.GetGuildBorderwallCountriesRow) []DatasetItem {
					result := make([]DatasetItem, 0, len(rows))
					for _, row := range rows {
						result = append(
							result,
							DatasetItem{Key: fmt.Sprint(row.Key), Value: int(row.Value)},
						)
					}

					return result
				}

				deviceDataset := make([]DatasetItem, 0, len(devices))
				for _, row := range devices {
					deviceDataset = append(
						deviceDataset,
						DatasetItem{Key: fmt.Sprint(row.Key), Value: int(row.Value)},
					)
				}

				return map[string]any{
					"total_requests":    summary.Requests,
					"verified_requests": summary.Verified,
					"pass_rate": func() float64 {
						if summary.Requests == 0 {
							return 0
						}

						return float64(summary.Verified) * 100 / float64(summary.Requests)
					}(),
					"average_verification_time_ms": summary.AverageVerificationTimeMs,
					"requests_by_time":             requestSeries,
					"verified_by_time":             verifiedSeries,
					"top_countries":                toDataset(countries),
					"top_devices":                  deviceDataset,
				}, nil
			}

			current, err := build(request.From, request.To)
			if err == nil {
				var previous map[string]any

				previous, err = build(request.PreviousFrom, request.PreviousTo)
				if err == nil {
					ctx.JSON(
						http.StatusOK,
						BaseResponse{
							Ok:   true,
							Data: map[string]any{"current": current, "previous": previous},
						},
					)

					return
				}
			}

			welcomer.Logger.Error().
				Err(err).
				Int64("guild_id", int64(guildID)).
				Msg("failed to get guild borderwall analytics")
			ctx.JSON(
				http.StatusInternalServerError,
				NewBaseResponse(NewGenericErrorWithLineNumber(), nil),
			)
		},
	)
}

func getGuildAnalyticsRoleActivity(ctx *gin.Context) {
	withGuildAnalyticsRequest(
		ctx,
		func(ctx *gin.Context, guildID discord.Snowflake, request *AnalyticsRequest) {
			build := func(from, to time.Time) ([]DatasetItem, error) {
				rows, err := welcomer.Queries.GetGuildRoleActivity(
					ctx,
					database.GetGuildRoleActivityParams{
						GuildID:           int64(guildID),
						TimeroleEventType: int32(database.ScienceGuildEventTypeTimeRoleGiven),
						ReactionroleEventType: int32(
							database.ScienceGuildEventTypeReactionRoleGiven,
						),
						AutoroleEventType: int32(database.ScienceGuildEventTypeAutoRoleGiven),
						FreeroleEventType: int32(database.ScienceGuildEventTypeFreeRoleGiven),
						DateFrom:          from,
						DateTo:            to,
					},
				)
				if err != nil {
					return nil, fmt.Errorf("get guild role activity: %w", err)
				}

				counts := make(map[string]int)

				for _, row := range rows {
					roleID := fmt.Sprint(row.Key)
					if roleID == "" {
						continue
					}

					module := ""

					switch database.ScienceGuildEventType(row.EventType) {
					case database.ScienceGuildEventTypeAutoRoleGiven:
						module = "Autoroles"
					case database.ScienceGuildEventTypeFreeRoleGiven:
						module = "Freeroles"
					case database.ScienceGuildEventTypeReactionRoleGiven:
						module = "Reaction roles"
					case database.ScienceGuildEventTypeTimeRoleGiven:
						module = "Timeroles"
					default:
						continue
					}

					counts[module+"\x00"+roleID] += int(row.Value)
				}

				roles, err := welcomer.SandwichClient.FetchGuildRole(
					ctx,
					&sandwich.FetchGuildRoleRequest{GuildId: int64(guildID)},
				)
				if err != nil {
					return nil, fmt.Errorf("get guild roles for analytics: %w", err)
				}

				result := make([]DatasetItem, 0, len(counts))
				for key, count := range counts {
					module, roleID, _ := strings.Cut(key, "\x00")

					var id int64
					if _, scanErr := fmt.Sscan(roleID, &id); scanErr != nil {
						return nil, fmt.Errorf("parse role ID %q: %w", roleID, scanErr)
					}

					role := roles.GetRoles()[id]

					name := fmt.Sprintf("Deleted role (ID: %d)", id)
					if role != nil && role.GetName() != "" {
						name = role.GetName()
					}

					result = append(result, DatasetItem{Key: module + ": " + name, Value: count})
				}

				sort.Slice(result, func(i, j int) bool { return result[i].Value > result[j].Value })

				return result, nil
			}

			current, err := build(request.From, request.To)
			if err == nil {
				var previous []DatasetItem

				previous, err = build(request.PreviousFrom, request.PreviousTo)
				if err == nil {
					ctx.JSON(http.StatusOK, BaseResponse{Ok: true, Data: map[string]any{
						"current":  map[string]any{"role_changes": current},
						"previous": map[string]any{"role_changes": previous},
					}})

					return
				}
			}

			welcomer.Logger.Error().
				Err(err).
				Int64("guild_id", int64(guildID)).
				Msg("failed to get guild role analytics")
			ctx.JSON(
				http.StatusInternalServerError,
				NewBaseResponse(NewGenericErrorWithLineNumber(), nil),
			)
		},
	)
}

func getGuildAnalyticsUserDemographics(ctx *gin.Context) {
	withGuildAnalyticsRequest(
		ctx,
		func(ctx *gin.Context, guildID discord.Snowflake, _ *AnalyticsRequest) {
			_, err := welcomer.SandwichClient.RequestGuildChunk(
				ctx,
				&sandwich.RequestGuildChunkRequest{
					GuildId:     int64(guildID),
					AlwaysChunk: false,
				},
			)
			if err != nil {
				welcomer.Logger.Error().
					Err(err).
					Int64("guild_id", int64(guildID)).
					Msg("failed to request guild chunk for demographics")
				ctx.JSON(
					http.StatusInternalServerError,
					NewBaseResponse(NewGenericErrorWithLineNumber(), nil),
				)

				return
			}

			members, err := welcomer.SandwichClient.FetchGuildMember(
				ctx,
				&sandwich.FetchGuildMemberRequest{GuildId: int64(guildID)},
			)
			if err != nil {
				welcomer.Logger.Error().
					Err(err).
					Int64("guild_id", int64(guildID)).
					Msg("failed to fetch members for demographics")
				ctx.JSON(
					http.StatusInternalServerError,
					NewBaseResponse(NewGenericErrorWithLineNumber(), nil),
				)

				return
			}

			type memberDemographic struct {
				accountAgeDays   int
				accountAgeMonths int
				joinedAt         time.Time
				hasJoinedAt      bool
			}

			now := time.Now()

			ages := make([]memberDemographic, 0, len(members.GetGuildMembers()))
			for _, member := range members.GetGuildMembers() {
				user := member.GetUser()
				if user == nil || user.GetID() == 0 {
					continue
				}

				createdAt := discordAccountCreatedAt(user.GetID())
				if createdAt.After(now) {
					continue
				}

				joinedAt, err := time.Parse(time.RFC3339, member.GetJoinedAt())

				months := (now.Year()-createdAt.Year())*12 + int(now.Month()-createdAt.Month())
				if now.Day() < createdAt.Day() {
					months--
				}

				ageDays := int(now.Sub(createdAt).Hours() / 24)

				ages = append(ages, memberDemographic{
					accountAgeDays:   max(0, ageDays),
					accountAgeMonths: max(0, months),
					joinedAt:         joinedAt,
					hasJoinedAt:      err == nil,
				})
			}

			sort.Slice(
				ages,
				func(i, j int) bool { return ages[i].accountAgeDays < ages[j].accountAgeDays },
			)

			maxAgeMonths := 0

			var totalAgeDays int64

			scatter := make([]map[string]any, 0, min(len(ages), 1000))
			scatterStep := max(1, (len(ages)+999)/1000)

			for index, member := range ages {
				totalAgeDays += int64(member.accountAgeDays)

				maxAgeMonths = max(maxAgeMonths, member.accountAgeMonths)
				if index%scatterStep == 0 && member.hasJoinedAt {
					scatter = append(scatter, map[string]any{
						"account_age_days": member.accountAgeDays,
						"joined_at":        member.joinedAt.Format(time.RFC3339),
					})
				}
			}

			averageDays := int64(0)
			medianDays := 0
			percentile90Days := 0

			if len(ages) > 0 {
				averageDays = totalAgeDays / int64(len(ages))

				middle := len(ages) / 2
				if len(ages)%2 == 0 {
					medianDays = (ages[middle-1].accountAgeDays + ages[middle].accountAgeDays) / 2
				} else {
					medianDays = ages[middle].accountAgeDays
				}

				percentile90Days = ages[(len(ages)*90+99)/100-1].accountAgeDays
			}

			monthlyCounts := make([]int, maxAgeMonths+1)
			for _, member := range ages {
				monthlyCounts[member.accountAgeMonths]++
			}

			distribution := make([]DatasetItem, 0, len(monthlyCounts))
			for month, count := range monthlyCounts {
				distribution = append(distribution, DatasetItem{
					Key:   fmt.Sprintf("%d mo", month),
					Value: count,
				})
			}

			ctx.JSON(http.StatusOK, BaseResponse{Ok: true, Data: map[string]any{
				"current": map[string]any{
					"member_count":             len(ages),
					"average_account_age_days": max(0, int(averageDays)),
					"median_account_age_days":  max(0, medianDays),
					"p90_account_age_days":     max(0, percentile90Days),
					"account_age_distribution": distribution,
					"account_age_vs_join_date": scatter,
				},
			}})
		},
	)
}
