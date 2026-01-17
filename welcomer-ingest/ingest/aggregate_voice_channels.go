package ingest

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	"github.com/jackc/pgx/v4"
)

const (
	jobNameVoiceChannels = "guild_voice_channel_stats"
	backfillWindowVC     = time.Minute * 5  // allow late events
	maxSessionGap        = time.Minute * 10 // maximum gap before assuming session ended
)

// AggregateVoiceChannels runs on an interval and processes ingest_voice_channel_events
// into guild_voice_channel_stats, tracking user time in voice channels.
func AggregateVoiceChannels(ctx context.Context, waitGroup *sync.WaitGroup, interval time.Duration) {
	ticker := time.NewTicker(time.Millisecond)
	hasReset := false

	waitGroup.Add(1)

	go func() {
		defer waitGroup.Done()

		for {
			select {
			case <-ticker.C:
				if !hasReset {
					ticker.Reset(interval)
					hasReset = true
				}

				startTime := time.Now()

				if err := aggregateVoiceChannels(ctx); err != nil {
					welcomer.Logger.Error().Err(err).Msg("aggregate voice channels failed")
				} else {
					if time.Since(startTime) > slowJobWarning {
						welcomer.Logger.Warn().Dur("duration", time.Since(startTime)).Msg("aggregate voice channels too long to run")
					} else {
						welcomer.Logger.Info().Dur("duration", time.Since(startTime)).Msg("aggregate voice channels completed")
					}
				}
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
}

type voiceChannelEvent struct {
	guildID    int64
	userID     int64
	channelID  int64
	eventType  welcomer.IngestVoiceChannelEventType
	occurredAt time.Time
}

type userSession struct {
	guildID    int64
	userID     int64
	channelID  int64
	startTs    time.Time
	lastSeenTs time.Time
}

func aggregateVoiceChannels(ctx context.Context) (err error) {
	welcomer.Logger.Info().Msg("starting aggregate voice channels job")

	tx, err := welcomer.Pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("error beginning transaction: %w", err)
	}

	defer func() {
		if err != nil {
			welcomer.Logger.Info().Err(err).Msg("rolling back aggregate voice channels transaction")

			rollbackErr := tx.Rollback(ctx)
			if rollbackErr != nil {
				welcomer.Logger.Error().Err(rollbackErr).Msg("error rolling back aggregate voice channels transaction")
			}
		}
	}()

	queries := database.New(tx)

	lastProcessed := time.Time{}
	checkpoint, err := queries.GetJobCheckpointByName(ctx, jobNameVoiceChannels)

	if err == nil {
		lastProcessed = checkpoint.LastProcessedTs
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("error getting job checkpoint: %w", err)
	}

	upperBound := time.Now().UTC()
	lowerBound := lastProcessed.Add(-backfillWindowVC)

	welcomer.Logger.Info().
		Time("lower_bound", lowerBound).
		Time("upper_bound", upperBound).
		Msg("aggregating voice channels")

	// Fetch all recent events
	events, err := queries.FetchVoiceChannelEventsForAggregation(ctx, database.FetchVoiceChannelEventsForAggregationParams{
		OccurredAt:   lowerBound,
		OccurredAt_2: upperBound,
	})
	if err != nil {
		return fmt.Errorf("error querying voice channel events: %w", err)
	}

	// Convert to our working type
	voiceEvents := make([]voiceChannelEvent, len(events))
	var maxProcessed time.Time

	for i, event := range events {
		voiceEvents[i] = voiceChannelEvent{
			guildID:    event.GuildID,
			userID:     event.UserID,
			channelID:  event.ChannelID.Int64,
			eventType:  welcomer.IngestVoiceChannelEventType(event.EventType),
			occurredAt: event.OccurredAt,
		}

		if event.OccurredAt.After(maxProcessed) {
			maxProcessed = event.OccurredAt
		}
	}
	if err != nil {
		return fmt.Errorf("error processing voice channel events: %w", err)
	}

	err = processVoiceChannelEvents(ctx, queries, voiceEvents)
	if err != nil {
		return fmt.Errorf("error processing voice channel events: %w", err)
	}

	// Clean up stale sessions (users who haven't been seen in a long time)
	err = cleanupStaleSessions(ctx, queries, upperBound)
	if err != nil {
		return fmt.Errorf("error cleaning up stale sessions: %w", err)
	}

	// Update checkpoint
	err = queries.UpsertJobCheckpoint(ctx, database.UpsertJobCheckpointParams{
		JobName:         jobNameVoiceChannels,
		LastProcessedTs: maxProcessed,
	})
	if err != nil {
		return fmt.Errorf("error upserting job checkpoint: %w", err)
	}

	// Commit transaction
	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

func processVoiceChannelEvents(ctx context.Context, queries *database.Queries, events []voiceChannelEvent) error {
	// Group events by user
	userEvents := make(map[string][]voiceChannelEvent)

	for _, event := range events {
		key := fmt.Sprintf("%d-%d", event.guildID, event.userID)
		userEvents[key] = append(userEvents[key], event)
	}

	// Process each user's events in chronological order
	for _, userEventList := range userEvents {
		err := processUserVoiceEvents(ctx, queries, userEventList)
		if err != nil {
			return err
		}
	}

	return nil
}

func processUserVoiceEvents(ctx context.Context, queries *database.Queries, events []voiceChannelEvent) error {
	if len(events) == 0 {
		return nil
	}

	guildID := events[0].guildID
	userID := events[0].userID

	// Get current open session for this user
	var currentSession *userSession

	sessionRow, err := queries.GetOpenVoiceChannelSession(ctx, database.GetOpenVoiceChannelSessionParams{
		GuildID: guildID,
		UserID:  userID,
	})

	if err == nil {
		currentSession = &userSession{
			guildID:    sessionRow.GuildID,
			userID:     sessionRow.UserID,
			channelID:  sessionRow.ChannelID,
			startTs:    sessionRow.StartTs,
			lastSeenTs: sessionRow.LastSeenTs,
		}
	} else if err != pgx.ErrNoRows {
		return err
	}

	// Process events in order
	for _, event := range events {
		switch event.eventType {
		case welcomer.IngestVoiceChannelEventTypeJoin:
			// If user has an existing session in a different channel, close it
			if currentSession != nil && currentSession.channelID != event.channelID {
				err := closeSession(ctx, queries, currentSession, event.occurredAt)
				if err != nil {
					return err
				}
			}

			// Start new session or update existing
			currentSession = &userSession{
				guildID:    guildID,
				userID:     userID,
				channelID:  event.channelID,
				startTs:    event.occurredAt,
				lastSeenTs: event.occurredAt,
			}

			err := queries.UpsertOpenVoiceChannelSession(ctx, database.UpsertOpenVoiceChannelSessionParams{
				GuildID:    guildID,
				UserID:     userID,
				ChannelID:  event.channelID,
				StartTs:    event.occurredAt,
				LastSeenTs: event.occurredAt,
			})
			if err != nil {
				return err
			}

		case welcomer.IngestVoiceChannelEventTypeLeave:
			// Close session when user leaves
			if currentSession != nil {
				err := closeSession(ctx, queries, currentSession, event.occurredAt)
				if err != nil {
					return err
				}
				currentSession = nil
			}

			// Delete the session record
			err := queries.DeleteOpenVoiceChannelSession(ctx, database.DeleteOpenVoiceChannelSessionParams{
				GuildID: guildID,
				UserID:  userID,
			})
			if err != nil {
				return err
			}

		case welcomer.IngestVoiceChannelEventTypeCheckpoint:
			// Update last seen time
			if currentSession != nil {
				currentSession.lastSeenTs = event.occurredAt

				err := queries.UpsertOpenVoiceChannelSession(ctx, database.UpsertOpenVoiceChannelSessionParams{
					GuildID:    guildID,
					UserID:     userID,
					ChannelID:  currentSession.channelID,
					StartTs:    currentSession.startTs,
					LastSeenTs: event.occurredAt,
				})
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// closeSession closes a user's voice channel session and creates a stat entry
func closeSession(ctx context.Context, queries *database.Queries, session *userSession, endTime time.Time) error {
	totalTime := endTime.Sub(session.startTs)
	totalTimeMs := totalTime.Milliseconds()

	if totalTimeMs < 0 {
		totalTimeMs = 0
	}

	err := queries.CreateVoiceChannelStat(ctx, database.CreateVoiceChannelStatParams{
		GuildID:     session.guildID,
		ChannelID:   session.channelID,
		UserID:      session.userID,
		StartTs:     session.startTs,
		EndTs:       endTime,
		TotalTimeMs: totalTimeMs,
	})

	return err
}

// cleanupStaleSessions handles sessions where the last seen time is too old
// This catches cases where a leave event was missed
func cleanupStaleSessions(ctx context.Context, queries *database.Queries, now time.Time) error {
	sessions, err := queries.GetAllOpenVoiceChannelSessions(ctx)
	if err != nil {
		return err
	}

	for _, session := range sessions {
		timeSinceLastSeen := now.Sub(session.LastSeenTs)

		// If session hasn't been seen in more than maxSessionGap, close it
		if timeSinceLastSeen > maxSessionGap {
			welcomer.Logger.Debug().
				Int64("guild_id", session.GuildID).
				Int64("user_id", session.UserID).
				Dur("time_since_last_seen", timeSinceLastSeen).
				Msg("closing stale voice channel session")

			err := queries.CreateVoiceChannelStat(ctx, database.CreateVoiceChannelStatParams{
				GuildID:     session.GuildID,
				ChannelID:   session.ChannelID,
				UserID:      session.UserID,
				StartTs:     session.StartTs,
				EndTs:       session.LastSeenTs, // Use last seen as end
				TotalTimeMs: session.LastSeenTs.Sub(session.StartTs).Milliseconds(),
			})
			if err != nil {
				return err
			}

			err = queries.DeleteOpenVoiceChannelSession(ctx, database.DeleteOpenVoiceChannelSessionParams{
				GuildID: session.GuildID,
				UserID:  session.UserID,
			})
			if err != nil {
				return err
			}
		}
	}

	return nil
}
