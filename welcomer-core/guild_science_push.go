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
	utils "github.com/WelcomerTeam/Welcomer/welcomer-utils"
	"github.com/jackc/pgtype"
	"github.com/rs/zerolog"
)

type PushGuildScienceHandler struct {
	sync.RWMutex

	logger zerolog.Logger
	db     *database.Queries

	limit  int
	buffer []database.CreateManyScienceGuildEventsParams
}

func NewPushGuildScienceHandler(db *database.Queries, logger zerolog.Logger, limit int) *PushGuildScienceHandler {
	return &PushGuildScienceHandler{
		RWMutex: sync.RWMutex{},
		logger:  logger,
		db:      db,
		limit:   limit,
		buffer:  make([]database.CreateManyScienceGuildEventsParams, 0, limit),
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
	guildEventUUID, err := utils.UUIDGen.NewV7()
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
	h.buffer = append(h.buffer, event)
	defer h.Unlock()

	if len(h.buffer) >= h.limit {
		h.flushWithoutLock(ctx)
	}
}

func (h *PushGuildScienceHandler) Flush(ctx context.Context) {
	h.Lock()
	defer h.Unlock()

	h.flushWithoutLock(ctx)
}

func (h *PushGuildScienceHandler) flushWithoutLock(ctx context.Context) {
	if len(h.buffer) == 0 {
		return
	}

	_, err := h.db.CreateManyScienceGuildEvents(ctx, h.buffer)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to flush guild science events")

		return
	}

	h.buffer = h.buffer[:0]
}
