package welcomer

import (
	"context"

	subway "github.com/WelcomerTeam/Subway/subway"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	"github.com/jackc/pgx/v4/pgxpool"
)

type InteractionsContextKey int

const (
	PoolKey InteractionsContextKey = iota
	QueriesKey
	ManagerNameKey
	PushGuildScienceHandlerKey
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

// PushGuildScience context handler.
func AddPushGuildScienceToContext(ctx context.Context, v *PushGuildScienceHandler) context.Context {
	return context.WithValue(ctx, PushGuildScienceHandlerKey, v)
}

func GetPushGuildScienceFromContext(ctx context.Context) *PushGuildScienceHandler {
	value, _ := ctx.Value(PushGuildScienceHandlerKey).(*PushGuildScienceHandler)

	return value
}

// ManagerName context handler.
func AddManagerNameToContext(ctx context.Context, v string) context.Context {
	return context.WithValue(ctx, ManagerNameKey, v)
}

func GetManagerNameFromContext(ctx context.Context) string {
	url := subway.GetURLFromContext(ctx)
	query := url.Query()

	manager := query.Get("manager")
	if manager != "" {
		return manager
	}

	value, _ := ctx.Value(ManagerNameKey).(string)

	return value
}
