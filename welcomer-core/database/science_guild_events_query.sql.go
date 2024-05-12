// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: science_guild_events_query.sql

package database

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgtype"
)

const CreateScienceGuildEvent = `-- name: CreateScienceGuildEvent :one
INSERT INTO science_guild_events (guild_event_uuid, guild_id, created_at, event_type, data)
    VALUES (uuid_generate_v7(), $1, now(), $2, $3)
RETURNING
    guild_event_uuid, guild_id, created_at, event_type, data
`

type CreateScienceGuildEventParams struct {
	GuildID   int64        `json:"guild_id"`
	EventType int32        `json:"event_type"`
	Data      pgtype.JSONB `json:"data"`
}

func (q *Queries) CreateScienceGuildEvent(ctx context.Context, arg CreateScienceGuildEventParams) (*ScienceGuildEvents, error) {
	row := q.db.QueryRow(ctx, CreateScienceGuildEvent, arg.GuildID, arg.EventType, arg.Data)
	var i ScienceGuildEvents
	err := row.Scan(
		&i.GuildEventUuid,
		&i.GuildID,
		&i.CreatedAt,
		&i.EventType,
		&i.Data,
	)
	return &i, err
}

const GetScienceGuildEvent = `-- name: GetScienceGuildEvent :one
SELECT
    guild_event_uuid, guild_id, created_at, event_type, data
FROM
    science_guild_events
WHERE
    guild_event_uuid = $1
`

func (q *Queries) GetScienceGuildEvent(ctx context.Context, guildEventUuid uuid.UUID) (*ScienceGuildEvents, error) {
	row := q.db.QueryRow(ctx, GetScienceGuildEvent, guildEventUuid)
	var i ScienceGuildEvents
	err := row.Scan(
		&i.GuildEventUuid,
		&i.GuildID,
		&i.CreatedAt,
		&i.EventType,
		&i.Data,
	)
	return &i, err
}
