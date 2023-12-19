package welcomer

import (
	"context"

	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	"github.com/jackc/pgx/v4/pgxpool"
)

type WelcomerInteractionsContextKey int

const (
	PoolKey WelcomerInteractionsContextKey = iota
	QueriesKey
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

// Queries context handler.
func AddQueriesToContext(ctx context.Context, v *database.Queries) context.Context {
	return context.WithValue(ctx, QueriesKey, v)
}

func GetQueriesFromContext(ctx context.Context) *database.Queries {
	value, _ := ctx.Value(QueriesKey).(*database.Queries)

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
