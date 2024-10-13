// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: science_command_errors_query.sql

package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgtype"
)

const CreateCommandError = `-- name: CreateCommandError :one
INSERT INTO science_command_errors (command_uuid, created_at, trace, data)
    VALUES ($1, now(), $2, $3)
RETURNING
    command_uuid, created_at, trace, data
`

type CreateCommandErrorParams struct {
	CommandUuid uuid.UUID    `json:"command_uuid"`
	Trace       string       `json:"trace"`
	Data        pgtype.JSONB `json:"data"`
}

func (q *Queries) CreateCommandError(ctx context.Context, arg CreateCommandErrorParams) (*ScienceCommandErrors, error) {
	row := q.db.QueryRow(ctx, CreateCommandError, arg.CommandUuid, arg.Trace, arg.Data)
	var i ScienceCommandErrors
	err := row.Scan(
		&i.CommandUuid,
		&i.CreatedAt,
		&i.Trace,
		&i.Data,
	)
	return &i, err
}

const GetCommandError = `-- name: GetCommandError :one
SELECT
    science_command_usages.command_uuid, science_command_usages.created_at, guild_id, user_id, channel_id, command, errored, execution_time_ms, science_command_errors.command_uuid, science_command_errors.created_at, trace, data
FROM
    science_command_usages
    LEFT JOIN science_command_errors ON (science_command_errors.command_uuid = science_command_usages.command_uuid)
WHERE
    science_command_usages.command_uuid = $1
`

type GetCommandErrorRow struct {
	CommandUuid     uuid.UUID      `json:"command_uuid"`
	CreatedAt       time.Time      `json:"created_at"`
	GuildID         int64          `json:"guild_id"`
	UserID          int64          `json:"user_id"`
	ChannelID       sql.NullInt64  `json:"channel_id"`
	Command         string         `json:"command"`
	Errored         bool           `json:"errored"`
	ExecutionTimeMs int64          `json:"execution_time_ms"`
	CommandUuid_2   uuid.NullUUID  `json:"command_uuid_2"`
	CreatedAt_2     sql.NullTime   `json:"created_at_2"`
	Trace           sql.NullString `json:"trace"`
	Data            pgtype.JSONB   `json:"data"`
}

func (q *Queries) GetCommandError(ctx context.Context, commandUuid uuid.UUID) (*GetCommandErrorRow, error) {
	row := q.db.QueryRow(ctx, GetCommandError, commandUuid)
	var i GetCommandErrorRow
	err := row.Scan(
		&i.CommandUuid,
		&i.CreatedAt,
		&i.GuildID,
		&i.UserID,
		&i.ChannelID,
		&i.Command,
		&i.Errored,
		&i.ExecutionTimeMs,
		&i.CommandUuid_2,
		&i.CreatedAt_2,
		&i.Trace,
		&i.Data,
	)
	return &i, err
}
