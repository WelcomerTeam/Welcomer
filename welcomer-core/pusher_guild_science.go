package welcomer

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/WelcomerTeam/Discord/discord"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	"github.com/jackc/pgtype"
)

type PushGuildScienceHandler struct {
	sync.RWMutex

	limit  int
	Buffer []database.CreateManyScienceGuildEventsParams
}

func NewPushGuildScienceHandler(limit int) *PushGuildScienceHandler {
	return &PushGuildScienceHandler{
		RWMutex: sync.RWMutex{},
		limit:   limit,
		Buffer:  make([]database.CreateManyScienceGuildEventsParams, 0, limit),
	}
}

func (h *PushGuildScienceHandler) Run(ctx context.Context, interval time.Duration) {
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

func (h *PushGuildScienceHandler) Push(ctx context.Context, guildID, userID discord.Snowflake, eventType database.ScienceGuildEventType, data any) {
	guildEventUUID, err := UUIDGen.NewV7()
	if err != nil {
		panic(fmt.Errorf("failed to generate UUID: %w", err))
	}

	guildEventData := pgtype.JSON{
		Status: pgtype.Null,
	}

	if data != nil {
		dataBytes, err := json.Marshal(data)
		if err != nil {
			panic(fmt.Errorf("failed to marshal data: %w", err))
		}

		guildEventData.Bytes = dataBytes
		guildEventData.Status = pgtype.Present
	}

	h.PushRaw(ctx, database.CreateManyScienceGuildEventsParams{
		GuildEventUuid: guildEventUUID,
		GuildID:        int64(guildID),
		UserID:         sql.NullInt64{Int64: int64(userID), Valid: !userID.IsNil()},
		CreatedAt:      time.Now(),
		EventType:      int32(eventType),
		Data:           guildEventData,
	})
}

func (h *PushGuildScienceHandler) PushRaw(ctx context.Context, event database.CreateManyScienceGuildEventsParams) {
	h.Lock()
	h.Buffer = append(h.Buffer, event)
	h.Unlock()

	if len(h.Buffer) >= h.limit {
		h.Flush(ctx)
	}
}

func (h *PushGuildScienceHandler) Flush(ctx context.Context) {
	h.Lock()

	if len(h.Buffer) == 0 {
		h.Unlock()

		return
	}

	// Make a copy of the buffer to avoid holding the lock while flushing

	buf := make([]database.CreateManyScienceGuildEventsParams, len(h.Buffer))
	copy(buf, h.Buffer)
	h.Buffer = h.Buffer[:0]

	h.Unlock()

	_, err := Queries.CreateManyScienceGuildEvents(ctx, buf)
	if err != nil {
		Logger.Error().Err(err).Msg("failed to flush guild science events")

		return
	}
}
