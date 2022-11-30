package internal

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
)

type WelcomerInteractionsContextKey int

const (
	PoolKey WelcomerInteractionsContextKey = iota
	ManagerNameKey
)

// Arguments context handler.
func AddPoolToContext(ctx context.Context, v *pgxpool.Pool) context.Context {
	return context.WithValue(ctx, PoolKey, v)
}

func GetPoolFromContext(ctx context.Context) *pgxpool.Pool {
	value, _ := ctx.Value(PoolKey).(*pgxpool.Pool)

	return value
}

// ManagerName context handler.
func AddManagerNameToContext(ctx context.Context, v string) context.Context {
	return context.WithValue(ctx, ManagerNameKey, v)
}

func GetManagerNameFromContext(ctx context.Context) string {
	value, _ := ctx.Value(ManagerNameKey).(string)

	return value
}
