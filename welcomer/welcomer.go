package welcomer

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/WelcomerTeam/Discord/discord"
	sandwich "github.com/WelcomerTeam/Sandwich/sandwich"
	"github.com/WelcomerTeam/Welcomer/welcomer/database"
	plugins "github.com/WelcomerTeam/Welcomer/welcomer/plugins"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/rs/zerolog"
	"golang.org/x/xerrors"
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

func (w *Welcomer) Register() (err error) {
	w.Bot = sandwich.NewBot(sandwich.WhenMentionedOr(), w.Logger)

	// Register cogs

	w.Bot.MustRegisterCog(plugins.NewPogCog())

	// Register events

	w.Bot.RegisterOnInteractionCreateEvent(func(ctx *sandwich.EventContext, interaction discord.Interaction) (err error) {
		interactionCtx, err := w.Bot.GetInteractionContext(ctx, interaction)
		if err != nil {
			return xerrors.Errorf("Failed to get interaction context: %v", err)
		}

		if interactionCtx.InteractionCommand == nil {
			// TODO: This should not happen, this means a command registered does not have a handler.
			// Respond with some generic message that this command does not currently exist.

			return
		}

		interactionStart := time.Now()
		resp, err := w.Bot.InvokeInteraction(interactionCtx)
		// Capture interaction error
		if err != nil {
			ctx.Logger.Error().Err(err).Msg("Failed to process interaction")
		}

		if resp != nil {
			sendResponseErr := interaction.SendResponse(ctx.Session, resp.Type, resp.Data.WebhookMessageParams, resp.Data.Choices)
			if sendResponseErr != nil {
				ctx.Logger.Error().Err(sendResponseErr).Msg("Failed to send interaction response")
			}
		}

		// Capture interaction execution
		_, captureErr := w.CaptureInteractionExecution(interactionCtx, interactionStart, err)
		if captureErr != nil {
			ctx.Logger.Error().Err(captureErr).Msg("Failed to capture interaction response")
		}

		return nil
	})

	return nil
}

func (w *Welcomer) CaptureInteractionExecution(interactionCtx *sandwich.InteractionContext, interactionStart time.Time, interactionError error) (commandUUID uuid.UUID, err error) {
	var userID discord.Snowflake

	if interactionCtx.Member != nil {
		userID = interactionCtx.Member.User.ID
	} else {
		userID = interactionCtx.User.ID
	}

	context := context.Background()

	var guildID int64
	if interactionCtx.GuildID != nil {
		guildID = int64(*interactionCtx.GuildID)
	}

	var channelID sql.NullInt64
	_ = channelID.Scan(interactionCtx.ChannelID)

	usage, err := w.Database.CreateCommandUsage(context, &database.CreateCommandUsageParams{
		GuildID:         guildID,
		UserID:          int64(userID),
		ChannelID:       channelID,
		Command:         strings.Join(append([]string{interactionCtx.InteractionCommand.Name}, interactionCtx.CommandTree...), ","),
		Errored:         interactionError != nil,
		ExecutionTimeMs: time.Since(interactionStart).Milliseconds(),
	})
	if err != nil {
		return commandUUID, xerrors.Errorf("Failed to create command usage: %v", err.Error())
	}

	if interactionError != nil {
		var data pgtype.JSONB
		_ = data.Set(interactionCtx.Arguments)

		_, err = w.Database.CreateCommandError(context, &database.CreateCommandErrorParams{
			CommandUuid: usage.CommandUuid,
			Trace:       interactionError.Error(),
			Data:        data,
		})
		if err != nil {
			return commandUUID, xerrors.Errorf("Failed to create command error: %v", err.Error())
		}
	}

	return commandUUID, nil
}
