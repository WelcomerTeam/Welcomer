package welcomer

import (
	"context"
	"fmt"
	"os"

	sandwich "github.com/WelcomerTeam/Sandwich/sandwich"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	plugins "github.com/WelcomerTeam/Welcomer/welcomer-gateway/plugins"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/rs/zerolog"
)

type Welcomer struct {
	Logger zerolog.Logger
	Bot    *sandwich.Bot

	pool     *pgxpool.Pool
	Database *database.Queries
}

func NewWelcomer(identifierName string, sandwichClient *sandwich.Sandwich) (welcomer *Welcomer) {
	welcomer = &Welcomer{
		Logger: sandwichClient.Logger,
	}

	context := context.Background()

	// Setup Postgres pool

	poolConnectionString := os.Getenv("WELCOMER_DATABASE_URL")

	pool, err := pgxpool.Connect(context, poolConnectionString)
	if err != nil {
		panic(fmt.Sprintf("pgxpool.Connect(%s): %v", poolConnectionString, err.Error()))
	}

	welcomer.pool = pool
	welcomer.Database = database.New(welcomer.pool)

	// Register bot (cogs, events)

	if err := welcomer.Register(); err != nil {
		panic(fmt.Sprintf("welcomer.Register(): %v", err.Error()))
	}

	return welcomer
}

func (w *Welcomer) Register() error {
	w.Bot = sandwich.NewBot(w.Logger)

	// Register cogs

	w.Bot.MustRegisterCog(plugins.NewWelcomerCog())

	return nil
}

// func (w *Welcomer) CaptureInteractionExecution(interactionCtx *sandwich.InteractionContext, interactionStart time.Time, interactionError error) (commandUUID uuid.UUID, err error) {
// 	var userID discord.Snowflake

// 	if interactionCtx.Member != nil {
// 		userID = interactionCtx.Member.User.ID
// 	} else {
// 		userID = interactionCtx.User.ID
// 	}

// 	context := context.Background()

// 	var guildID int64
// 	if interactionCtx.GuildID != nil {
// 		guildID = int64(*interactionCtx.GuildID)

// 	var channelID sql.NullInt64
// 	_ = channelID.Scan(interactionCtx.ChannelID)

// 	usage, err := w.Database.CreateCommandUsage(context, &database.CreateCommandUsageParams{
// 		GuildID:         guildID,
// 		UserID:          int64(userID),
// 		ChannelID:       channelID,
// 		Command:         strings.Join(append([]string{interactionCtx.InteractionCommand.Name}, interactionCtx.CommandTree...), ","),
// 		Errored:         interactionError != nil,
// 		ExecutionTimeMs: time.Since(interactionStart).Milliseconds(),
// 	})
// 	if err != nil {
// 		return commandUUID, xerrors.Errorf("Failed to create command usage: %v", err.Error())
// 	}

// 	if interactionError != nil {
// 		var data pgtype.JSONB
// 		_ = data.Set(interactionCtx.Arguments)

// 		_, err = w.Database.CreateCommandError(context, &database.CreateCommandErrorParams{
// 			CommandUuid: usage.CommandUuid,
// 			Trace:       interactionError.Error(),
// 			Data:        data,
// 		})
// 		if err != nil {
// 			return commandUUID, xerrors.Errorf("Failed to create command error: %v", err.Error())
// 		}
// 	}

// 	return commandUUID, nil
// }
