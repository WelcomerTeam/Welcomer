// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: guild_settings_timeroles_query.sql

package database

import (
	"context"

	"github.com/jackc/pgtype"
)

const CreateOrUpdateTimeRolesGuildSettings = `-- name: CreateOrUpdateTimeRolesGuildSettings :one
INSERT INTO guild_settings_timeroles (guild_id, toggle_enabled, timeroles)
    VALUES ($1, $2, $3)
ON CONFLICT(guild_id) DO UPDATE
    SET toggle_enabled = EXCLUDED.toggle_enabled,
        timeroles = EXCLUDED.timeroles
RETURNING
    guild_id, toggle_enabled, timeroles
`

type CreateOrUpdateTimeRolesGuildSettingsParams struct {
	GuildID       int64        `json:"guild_id"`
	ToggleEnabled bool         `json:"toggle_enabled"`
	Timeroles     pgtype.JSONB `json:"timeroles"`
}

func (q *Queries) CreateOrUpdateTimeRolesGuildSettings(ctx context.Context, arg *CreateOrUpdateTimeRolesGuildSettingsParams) (*GuildSettingsTimeroles, error) {
	row := q.db.QueryRow(ctx, CreateOrUpdateTimeRolesGuildSettings, arg.GuildID, arg.ToggleEnabled, arg.Timeroles)
	var i GuildSettingsTimeroles
	err := row.Scan(&i.GuildID, &i.ToggleEnabled, &i.Timeroles)
	return &i, err
}

const CreateTimeRolesGuildSettings = `-- name: CreateTimeRolesGuildSettings :one
INSERT INTO guild_settings_timeroles (guild_id, toggle_enabled, timeroles)
    VALUES ($1, $2, $3)
RETURNING
    guild_id, toggle_enabled, timeroles
`

type CreateTimeRolesGuildSettingsParams struct {
	GuildID       int64        `json:"guild_id"`
	ToggleEnabled bool         `json:"toggle_enabled"`
	Timeroles     pgtype.JSONB `json:"timeroles"`
}

func (q *Queries) CreateTimeRolesGuildSettings(ctx context.Context, arg *CreateTimeRolesGuildSettingsParams) (*GuildSettingsTimeroles, error) {
	row := q.db.QueryRow(ctx, CreateTimeRolesGuildSettings, arg.GuildID, arg.ToggleEnabled, arg.Timeroles)
	var i GuildSettingsTimeroles
	err := row.Scan(&i.GuildID, &i.ToggleEnabled, &i.Timeroles)
	return &i, err
}

const GetTimeRolesGuildSettings = `-- name: GetTimeRolesGuildSettings :one
SELECT
    guild_id, toggle_enabled, timeroles
FROM
    guild_settings_timeroles
WHERE
    guild_id = $1
`

func (q *Queries) GetTimeRolesGuildSettings(ctx context.Context, guildID int64) (*GuildSettingsTimeroles, error) {
	row := q.db.QueryRow(ctx, GetTimeRolesGuildSettings, guildID)
	var i GuildSettingsTimeroles
	err := row.Scan(&i.GuildID, &i.ToggleEnabled, &i.Timeroles)
	return &i, err
}

const UpdateTimeRolesGuildSettings = `-- name: UpdateTimeRolesGuildSettings :execrows
UPDATE
    guild_settings_timeroles
SET
    toggle_enabled = $2,
    timeroles = $3
WHERE
    guild_id = $1
`

type UpdateTimeRolesGuildSettingsParams struct {
	GuildID       int64        `json:"guild_id"`
	ToggleEnabled bool         `json:"toggle_enabled"`
	Timeroles     pgtype.JSONB `json:"timeroles"`
}

func (q *Queries) UpdateTimeRolesGuildSettings(ctx context.Context, arg *UpdateTimeRolesGuildSettingsParams) (int64, error) {
	result, err := q.db.Exec(ctx, UpdateTimeRolesGuildSettings, arg.GuildID, arg.ToggleEnabled, arg.Timeroles)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}
