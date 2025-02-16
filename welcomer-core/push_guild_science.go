package welcomer

import (
	"context"
	"sync"

	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
)

type PushGuildScienceHandler struct {
	sync.RWMutex

	db *database.Queries

	limit  int
	buffer []database.CreateManyScienceGuildEventsParams
}

func NewPushGuildScienceHandler(db *database.Queries, limit int) *PushGuildScienceHandler {
	return &PushGuildScienceHandler{
		RWMutex: sync.RWMutex{},
		db:      db,
		limit:   limit,
		buffer:  make([]database.CreateManyScienceGuildEventsParams, 0, limit),
	}
}

func (h *PushGuildScienceHandler) Push(ctx context.Context, event database.CreateManyScienceGuildEventsParams) {
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

	h.db.CreateManyScienceGuildEvents(ctx, h.buffer)

	clear(h.buffer)
}
