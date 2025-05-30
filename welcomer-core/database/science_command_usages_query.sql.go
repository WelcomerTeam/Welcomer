// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: science_command_usages_query.sql

package database

import (
	"context"
	"database/sql"

	"github.com/gofrs/uuid"
)

const CreateCommandUsage = `-- name: CreateCommandUsage :one
INSERT INTO science_command_usages (command_uuid, created_at, guild_id, user_id, channel_id, command, errored, execution_time_ms)
    VALUES (uuid_generate_v7(), now(), $1, $2, $3, $4, $5, $6)
RETURNING
    command_uuid, created_at, guild_id, user_id, channel_id, command, errored, execution_time_ms
`

type CreateCommandUsageParams struct {
	GuildID         int64         `json:"guild_id"`
	UserID          int64         `json:"user_id"`
	ChannelID       sql.NullInt64 `json:"channel_id"`
	Command         string        `json:"command"`
	Errored         bool          `json:"errored"`
	ExecutionTimeMs int64         `json:"execution_time_ms"`
}

func (q *Queries) CreateCommandUsage(ctx context.Context, arg CreateCommandUsageParams) (*ScienceCommandUsages, error) {
	row := q.db.QueryRow(ctx, CreateCommandUsage,
		arg.GuildID,
		arg.UserID,
		arg.ChannelID,
		arg.Command,
		arg.Errored,
		arg.ExecutionTimeMs,
	)
	var i ScienceCommandUsages
	err := row.Scan(
		&i.CommandUuid,
		&i.CreatedAt,
		&i.GuildID,
		&i.UserID,
		&i.ChannelID,
		&i.Command,
		&i.Errored,
		&i.ExecutionTimeMs,
	)
	return &i, err
}

const GetCommandUsage = `-- name: GetCommandUsage :one
SELECT
    command_uuid, created_at, guild_id, user_id, channel_id, command, errored, execution_time_ms
FROM
    science_command_usages
WHERE
    command_uuid = $1
`

func (q *Queries) GetCommandUsage(ctx context.Context, commandUuid uuid.UUID) (*ScienceCommandUsages, error) {
	row := q.db.QueryRow(ctx, GetCommandUsage, commandUuid)
	var i ScienceCommandUsages
	err := row.Scan(
		&i.CommandUuid,
		&i.CreatedAt,
		&i.GuildID,
		&i.UserID,
		&i.ChannelID,
		&i.Command,
		&i.Errored,
		&i.ExecutionTimeMs,
	)
	return &i, err
}
