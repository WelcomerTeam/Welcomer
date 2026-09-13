package backend

import (
	"context"
	"fmt"
	"time"

	welcomer "github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

// rateLimiterScript atomically increments the request counter for the window
// and sets its expiry only on the first request, returning the new count.
var rateLimiterScript = redis.NewScript(`
local current = redis.call("INCR", KEYS[1])
if current == 1 then
	redis.call("PEXPIRE", KEYS[1], ARGV[1])
end
return current
`)

type RateLimiterRule struct {
	UUID     uuid.UUID // Unique identifier for the rate limiter rule
	Requests int
	Window   time.Duration
}

func NewRateLimiterRule(requests int, window time.Duration) *RateLimiterRule {
	return &RateLimiterRule{
		UUID:     uuid.New(),
		Requests: requests,
		Window:   window,
	}
}

// CanRequest reports whether a request for the given key is allowed under
// this rule, incrementing the underlying Redis counter as a side effect.
func (r *RateLimiterRule) CanRequest(ctx context.Context, key string) (bool, error) {
	redisKey := fmt.Sprintf("ratelimit:%s:%s", r.UUID.String(), key)

	count, err := rateLimiterScript.Run(ctx, welcomer.RedisClient, []string{redisKey}, r.Window.Milliseconds()).Int()
	if err != nil {
		return false, fmt.Errorf("failed to run rate limiter script: %w", err)
	}

	return count <= r.Requests, nil
}
