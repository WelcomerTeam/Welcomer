package welcomer

import (
	"context"
	"sync"
	"time"

	"github.com/WelcomerTeam/Discord/discord"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
)

type PushIngestMessageEventsHandler struct {
	sync.RWMutex

	limit  int
	Buffer []database.CreateManyIngestMessageEventsParams
}

func NewPushIngestMessageEventsHandler(limit int) *PushIngestMessageEventsHandler {
	return &PushIngestMessageEventsHandler{
		RWMutex: sync.RWMutex{},
		limit:   limit,
		Buffer:  make([]database.CreateManyIngestMessageEventsParams, 0, limit),
	}
}

func (h *PushIngestMessageEventsHandler) Run(ctx context.Context, interval time.Duration) {
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

func (h *PushIngestMessageEventsHandler) Push(ctx context.Context, messageID, guildID, channelID, userID discord.Snowflake, eventType IngestMessageEventType, occurredAt time.Time) {
	h.PushRaw(ctx, database.CreateManyIngestMessageEventsParams{
		MessageID:  int64(messageID),
		GuildID:    int64(guildID),
		ChannelID:  int64(channelID),
		UserID:     int64(userID),
		EventType:  int16(eventType),
		OccurredAt: occurredAt,
	})
}

func (h *PushIngestMessageEventsHandler) PushRaw(ctx context.Context, event database.CreateManyIngestMessageEventsParams) {
	h.Lock()
	h.Buffer = append(h.Buffer, event)
	h.Unlock()

	if len(h.Buffer) >= h.limit {
		h.Flush(ctx)
	}
}

func (h *PushIngestMessageEventsHandler) Flush(ctx context.Context) {
	h.Lock()

	if len(h.Buffer) == 0 {
		h.Unlock()

		return
	}

	// Make a copy of the buffer to avoid holding the lock while flushing

	buf := make([]database.CreateManyIngestMessageEventsParams, len(h.Buffer))
	copy(buf, h.Buffer)
	h.Buffer = h.Buffer[:0]

	h.Unlock()

	_, err := Queries.CreateManyIngestMessageEvents(ctx, buf)
	if err != nil {
		Logger.Error().Err(err).Msg("failed to flush ingest message events")

		return
	}
}
