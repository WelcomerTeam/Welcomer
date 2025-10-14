package welcomer

import (
	"context"
	"log/slog"
	"time"

	sandwich_daemon "github.com/WelcomerTeam/Sandwich-Daemon"
	"github.com/go-redis/redis/v8"
)

var _ sandwich_daemon.DedupeProvider = (*RedisDedupeProvider)(nil)

type RedisDedupeProvider struct {
	client *redis.Client
	Logger *slog.Logger
}

func NewRedisDedupeProvider(client *redis.Client, logger *slog.Logger) *RedisDedupeProvider {
	return &RedisDedupeProvider{client: client, Logger: logger}
}

func (r *RedisDedupeProvider) Deduplicate(ctx context.Context, key string, ttl time.Duration) bool {
	set, err := r.client.SetNX(ctx, key, "1", ttl).Result()
	if err != nil {
		r.Logger.Error("Failed to set key in Redis", "key", key, "error", err)

		return false
	}

	return set
}

func (r *RedisDedupeProvider) Release(ctx context.Context, key string) {
	r.client.Del(ctx, key)
}
