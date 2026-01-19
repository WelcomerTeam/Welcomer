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

	aggregateMessageCountsSQL = `
SELECT date_trunc('hour', occurred_at) AS hour_ts,
	guild_id,
	channel_id,
	user_id,
	COUNT(message_id) AS message_count,
	MIN(occurred_at) AS min_ts,
	MAX(occurred_at) AS max_ts
FROM ingest_message_events
WHERE occurred_at > $1 AND occurred_at <= $2 AND event_type IN ($3, $4)
GROUP BY 1, 2, 3, 4`

	upsertMessageCountsSQL = `
INSERT INTO guild_message_counts_hour (hour_ts, guild_id, channel_id, user_id, message_count, min_ts)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (hour_ts, guild_id, channel_id, user_id) DO UPDATE SET
	message_count = EXCLUDED.message_count,
	min_ts = LEAST(guild_message_counts_hour.min_ts, EXCLUDED.min_ts)`
)

// AggregateMessageCounts runs on an interval and rolls ingest_message_events into hourly aggregates.
func AggregateMessageCounts(ctx context.Context, waitGroup *sync.WaitGroup, interval time.Duration) {
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
	}()
}

type aggRow struct {
	hourTimestamp      time.Time
	guildID, channelID int64
	userID             int64
	messageCount       int64
	minTimestamp       time.Time
	maxTimestamp       time.Time
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

	rows, err := tx.Query(ctx, aggregateMessageCountsSQL, lowerBound, upperBound, welcomer.IngestMessageEventTypeCreate, welcomer.IngestMessageEventTypeEdit)
	if err != nil {
		return fmt.Errorf("error querying aggregate message counts: %w", err)
	}

	var maxProcessed time.Time

	results := make([]aggRow, 0)

	for rows.Next() {
		var aggregationRow aggRow

		if err := rows.Scan(&aggregationRow.hourTimestamp, &aggregationRow.guildID, &aggregationRow.channelID, &aggregationRow.userID, &aggregationRow.messageCount, &aggregationRow.minTimestamp, &aggregationRow.maxTimestamp); err != nil {
			rows.Close()

			return fmt.Errorf("error scanning aggregate message counts: %w", err)
		}

		results = append(results, aggregationRow)

		if aggregationRow.maxTimestamp.After(maxProcessed) {
			maxProcessed = aggregationRow.maxTimestamp
		}
	}

	if err := rows.Err(); err != nil {
		rows.Close()

		return fmt.Errorf("error iterating aggregate message counts: %w", err)
	}

	// Close the result set before issuing any Exec to avoid conn busy
	rows.Close()

	for _, r := range results {
		if _, err := tx.Exec(ctx, upsertMessageCountsSQL, r.hourTimestamp, r.guildID, r.channelID, r.userID, r.messageCount, r.minTimestamp); err != nil {
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
