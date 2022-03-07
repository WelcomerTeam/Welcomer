package database

import (
	"context"

	sandwich "github.com/WelcomerTeam/Sandwich/sandwich"
	uuid "github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	jsoniter "github.com/json-iterator/go"
)

type ScienceCommandErrors struct {
	*pgxpool.Pool
}

type ScienceCommandError struct {
	Command   *ScienceCommandUsage
	Trace     string
	Arguments jsoniter.RawMessage
}

func newScienceCommandErrors(pool *pgxpool.Pool) *ScienceCommandErrors {
	return &ScienceCommandErrors{pool}
}

func (t *ScienceCommandErrors) Schema() string {
	return `
CREATE TABLE IF NOT EXISTS "science_command_errors" (
	"command_uuid" uuid NOT NULL UNIQUE PRIMARY KEY,
	"trace" text NOT NULL,
	"arguments" jsonb NULL,
	FOREIGN KEY ("command_uuid") REFERENCES "science_command_usages" ("command_uuid") ON DELETE CASCADE ON UPDATE CASCADE
);
`
}

func (t *ScienceCommandErrors) Create(ctx context.Context, commandUUID uuid.UUID, trace string, arguments map[string]*sandwich.Argument) (err error) {
	query := `INSERT INTO "science_command_errors" ("command_uuid", "trace", "arguments") VALUES ($1, $2, $3);`
	argumentJSON, _ := jsoniter.Marshal(arguments)
	_, err = t.Exec(ctx, query, commandUUID, trace, argumentJSON)

	return
}

func (t *ScienceCommandErrors) Get(ctx context.Context, commandUUID uuid.UUID) (scienceCommandError *ScienceCommandError, err error) {
	scienceCommandError = &ScienceCommandError{}
	scienceCommandError.Command = &ScienceCommandUsage{}

	query := `SELECT command_uuid, datetime, guild_id, user_id, channel_id, command, errored, execution_time_ms, trace, arguments FROM "science_command_usages" LEFT JOIN science_command_errors ON ("science_command_errors"."command_uuid" = "science_command_usages"."command_uuid") WHERE "science_command_usages"."command_uuid" = $1`
	row := t.QueryRow(ctx, query, commandUUID)
	err = row.Scan(
		&scienceCommandError.Command.CommandUUID,
		&scienceCommandError.Command.DateTime,
		&scienceCommandError.Command.GuildID,
		&scienceCommandError.Command.UserID,
		&scienceCommandError.Command.ChannelID,
		&scienceCommandError.Command.Command,
		&scienceCommandError.Command.Errored,
		&scienceCommandError.Command.ExecutionTimeMs,
		&scienceCommandError.Trace,
		&scienceCommandError.Arguments,
	)

	return
}
