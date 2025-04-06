package welcomer

import (
	"context"
	"fmt"
	"time"

	"github.com/WelcomerTeam/Discord/discord"
	sandwich "github.com/WelcomerTeam/Sandwich-Daemon/protobuf"
	internal "github.com/WelcomerTeam/Sandwich/sandwich"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	"github.com/jackc/pgx/v4/pgxpool"
	"google.golang.org/grpc"
)

var GRPCConnection *grpc.ClientConn

func SetupGRPCConnection(host string, opts ...grpc.DialOption) {
	var err error

	GRPCConnection, err = grpc.NewClient(host, opts...)
	if err != nil {
		panic(fmt.Sprintf(`grpc.NewClient(%s): %v`, host, err.Error()))
	}
}

var GRPCInterface internal.GRPC

func SetupGRPCInterface() {
	GRPCInterface = internal.NewDefaultGRPCClient()
}

var RESTInterface discord.RESTInterface

func SetupRESTInterface(restInterface discord.RESTInterface) {
	RESTInterface = restInterface
}

var SandwichClient sandwich.SandwichClient

func SetupSandwichClient() {
	SandwichClient = sandwich.NewSandwichClient(GRPCConnection)
}

var Pool *pgxpool.Pool

var Queries *database.Queries

func SetupDatabase(ctx context.Context, connectionString string) {
	var err error

	Pool, err = pgxpool.Connect(ctx, connectionString)
	if err != nil {
		panic(fmt.Sprintf(`pgxpool.Connect(%s): %v`, connectionString, err.Error()))
	}

	Queries = database.New(Pool)
}

var PushGuildScience *PushGuildScienceHandler

func SetupPushGuildScience(limit int) func(ctx context.Context, interval time.Duration) {
	PushGuildScience = NewPushGuildScienceHandler(limit)

	return PushGuildScience.Run
}
