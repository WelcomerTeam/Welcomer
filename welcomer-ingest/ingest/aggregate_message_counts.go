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
	jobNameMessageCounts = "guild_message_counts_hour"
	backfillWindow       = time.Minute * 2 // allow late events
	slowJobWarning       = time.Second * 45
)

// AggregateMessageCounts runs on an interval and rolls ingest_message_events into hourly aggregates.
func AggregateMessageCounts(ctx context.Context, waitGroup *sync.WaitGroup, interval time.Duration) {
	ticker := time.NewTicker(time.Millisecond)
	hasReset := false

	waitGroup.Go(func() {
		for {
			select {
			case <-ticker.C:
				if !hasReset {
					ticker.Reset(interval)

					hasReset = true
				}

				startTime := time.Now()

				if err := aggregateMessageCounts(ctx); err != nil {
					welcomer.Logger.Error().Err(err).Msg("aggregate message counts failed")
				} else {
					if time.Since(startTime) > slowJobWarning {
						welcomer.Logger.Warn().Dur("duration", time.Since(startTime)).Msg("aggregate message counts too long to run")
					} else {
						welcomer.Logger.Info().Dur("duration", time.Since(startTime)).Msg("aggregate message counts completed")
					}
				}
			case <-ctx.Done():
				ticker.Stop()

				return
			}
		}
	})
}

func aggregateMessageCounts(ctx context.Context) (err error) {
	welcomer.Logger.Info().Msg("starting aggregate message counts job")

	tx, err := welcomer.Pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("error beginning transaction: %w", err)
	}

	defer func() {
		if err != nil {
			welcomer.Logger.Info().Err(err).Msg("rolling back aggregate message counts transaction")

			err = tx.Rollback(ctx)
			if err != nil {
				welcomer.Logger.Error().Err(err).Msg("error rolling back aggregate message counts transaction")
			}
		}
	}()

	queries := database.New(tx)

	lastProcessed := time.Time{}
	checkpoint, err := queries.GetJobCheckpointByName(ctx, jobNameMessageCounts)

	if err == nil {
		lastProcessed = checkpoint.LastProcessedTs
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("error getting job checkpoint: %w", err)
	}

	upperBound := time.Now().UTC()
	lowerBound := lastProcessed.Add(-backfillWindow)

	welcomer.Logger.Info().
		Time("lower_bound", lowerBound).
		Time("upper_bound", upperBound).
		Msg("aggregating message counts")

	rows, err := welcomer.Queries.AggregateMessageCounts(ctx, database.AggregateMessageCountsParams{
		DateFrom:        lowerBound,
		DateTo:          upperBound,
		EventTypeCreate: int16(welcomer.IngestMessageEventTypeCreate),
		EventTypeEdit:   int16(welcomer.IngestMessageEventTypeEdit),
	})
	if err != nil {
		return fmt.Errorf("error querying aggregate message counts: %w", err)
	}

	var maxProcessed time.Time

	for _, row := range rows {
		if row.MaxTs.After(maxProcessed) {
			maxProcessed = row.MaxTs
		}
	}

	for _, r := range rows {
		err := welcomer.Queries.UpsertMessageCounts(ctx, database.UpsertMessageCountsParams{
			HourTs:       r.HourTs,
			GuildID:      r.GuildID,
			ChannelID:    r.ChannelID,
			UserID:       r.UserID,
			MessageCount: r.MessageCount,
			MinTs:        r.MinTs,
		})
		if err != nil {
			return fmt.Errorf("error upserting aggregate message counts: %w", err)
		}
	}

	if maxProcessed.After(lastProcessed) {
		if err := queries.UpsertJobCheckpoint(ctx, database.UpsertJobCheckpointParams{
			JobName:         jobNameMessageCounts,
			LastProcessedTs: maxProcessed,
		}); err != nil {
			return fmt.Errorf("error upserting job checkpoint: %w", err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}
