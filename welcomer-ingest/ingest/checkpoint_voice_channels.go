package ingest

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/WelcomerTeam/Discord/discord"
	sandwich "github.com/WelcomerTeam/Sandwich-Daemon/proto"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
)

func CheckpointVoiceChannels(ctx context.Context, waitGroup *sync.WaitGroup, interval time.Duration) {
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

				err := checkpointVoiceChannels(ctx)
				if err != nil {
					welcomer.Logger.Error().Err(err).Msg("error during checkpoint voice channels job")
				} else {
					if time.Since(startTime) > slowJobWarning {
						welcomer.Logger.Warn().Dur("duration", time.Since(startTime)).Msg("checkpoint voice channels too long to run")
					} else {
						welcomer.Logger.Info().Dur("duration", time.Since(startTime)).Msg("checkpoint voice channels completed")
					}
				}
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
}

func checkpointVoiceChannels(ctx context.Context) (err error) {
	welcomer.Logger.Info().Msg("starting checkpoint voice channels job")

	voiceStates, err := welcomer.SandwichClient.FetchVoiceStates(ctx, &sandwich.FetchVoiceStatesRequest{})
	if err != nil {
		return fmt.Errorf("error fetching voice states from sandwich: %w", err)
	}

	now := time.Now()

	for _, vs := range voiceStates.VoiceStates {
		if vs.Member != nil && vs.Member.User != nil && vs.Member.User.Bot {
			continue
		}

		welcomer.PusherIngestVoiceChannelEvents.Push(
			ctx,
			discord.Snowflake(vs.GuildID),
			discord.Snowflake(vs.ChannelID),
			discord.Snowflake(vs.UserID),
			welcomer.IngestVoiceChannelEventTypeCheckpoint,
			now,
		)
	}

	welcomer.PusherIngestVoiceChannelEvents.Flush(ctx)

	return nil
}
