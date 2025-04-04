package welcomer

import (
	"context"

	"github.com/WelcomerTeam/Discord/discord"
	pb "github.com/WelcomerTeam/Sandwich-Daemon/protobuf"
	sandwich "github.com/WelcomerTeam/Sandwich/sandwich"
	subway "github.com/WelcomerTeam/Subway/subway"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/rs/zerolog"
)

type InteractionsContextKey int

const (
	PoolKey InteractionsContextKey = iota
	QueriesKey
	ManagerNameKey
	PushGuildScienceHandlerKey
	SandwichClientKey
	GRPCInterfaceKey
	RESTInterfaceKey
	LoggerKey
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

// SandwichClient context handler.
func AddSandwichClientToContext(ctx context.Context, v pb.SandwichClient) context.Context {
	return context.WithValue(ctx, SandwichClientKey, v)
}

func GetSandwichClientFromContext(ctx context.Context) pb.SandwichClient {
	value, _ := ctx.Value(SandwichClientKey).(pb.SandwichClient)

	return value
}

// GRPCInterface context handler.

func AddGRPCInterfaceToContext(ctx context.Context, v sandwich.GRPC) context.Context {
	return context.WithValue(ctx, GRPCInterfaceKey, v)
}

func GetGRPCInterfaceFromContext(ctx context.Context) sandwich.GRPC {
	value, _ := ctx.Value(GRPCInterfaceKey).(sandwich.GRPC)

	return value
}

// RESTInterface context handler.

func AddRESTInterfaceToContext(ctx context.Context, v discord.RESTInterface) context.Context {
	return context.WithValue(ctx, RESTInterfaceKey, v)
}

func GetRESTInterfaceFromContext(ctx context.Context) discord.RESTInterface {
	value, _ := ctx.Value(RESTInterfaceKey).(discord.RESTInterface)

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

// Logger context handler.

func AddLoggerToContext(ctx context.Context, v zerolog.Logger) context.Context {
	return context.WithValue(ctx, LoggerKey, v)
}

func GetLoggerFromContext(ctx context.Context) zerolog.Logger {
	value, _ := ctx.Value(LoggerKey).(zerolog.Logger)

	return value
}
