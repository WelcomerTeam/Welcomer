package welcomer

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"github.com/WelcomerTeam/Discord/discord"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
)

type PushIngestVoiceChannelEventsHandler struct {
	sync.RWMutex

	limit  int
	Buffer []database.CreateManyIngestVoiceChannelEventsParams
}

func NewPushIngestVoiceChannelEventsHandler(limit int) *PushIngestVoiceChannelEventsHandler {
	return &PushIngestVoiceChannelEventsHandler{
		RWMutex: sync.RWMutex{},
		limit:   limit,
		Buffer:  make([]database.CreateManyIngestVoiceChannelEventsParams, 0, limit),
	}
}

func (h *PushIngestVoiceChannelEventsHandler) Run(ctx context.Context, interval time.Duration) {
	go func() {
		for {
			select {
			case <-time.After(interval):
				h.Flush(ctx)
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (h *PushIngestVoiceChannelEventsHandler) Push(ctx context.Context, guildID, channelID, userID discord.Snowflake, eventType IngestVoiceChannelEventType, occurredAt time.Time) {
	h.PushRaw(ctx, database.CreateManyIngestVoiceChannelEventsParams{
		GuildID:    int64(guildID),
		UserID:     int64(userID),
		ChannelID:  sql.NullInt64{Int64: int64(channelID), Valid: !channelID.IsNil()},
		EventType:  int16(eventType),
		OccurredAt: occurredAt,
	})
}

func (h *PushIngestVoiceChannelEventsHandler) PushRaw(ctx context.Context, event database.CreateManyIngestVoiceChannelEventsParams) {
	h.Lock()
	h.Buffer = append(h.Buffer, event)
	h.Unlock()

	if len(h.Buffer) >= h.limit {
		h.Flush(ctx)
	}
}

func (h *PushIngestVoiceChannelEventsHandler) Flush(ctx context.Context) {
	h.Lock()

	if len(h.Buffer) == 0 {
		h.Unlock()

		return
	}

	// Make a copy of the buffer to avoid holding the lock while flushing

	buf := make([]database.CreateManyIngestVoiceChannelEventsParams, len(h.Buffer))
	copy(buf, h.Buffer)
	h.Buffer = h.Buffer[:0]

	h.Unlock()

	_, err := Queries.CreateManyIngestVoiceChannelEvents(ctx, buf)
	if err != nil {
		Logger.Error().Err(err).Msg("failed to flush ingest voice channel events")

		return
	}
}
