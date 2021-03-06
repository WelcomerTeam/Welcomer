// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.14.0
// source: guild_settings_freeroles_query.sql

package database

import (
	"context"
	"database/sql"
)

const CreateFreeRolesGuildSettings = `-- name: CreateFreeRolesGuildSettings :one
INSERT INTO guild_settings_freeroles (guild_id)
    VALUES ($1)
RETURNING
    guild_id, toggle_enabled, roles
`

func (q *Queries) CreateFreeRolesGuildSettings(ctx context.Context, guildID int64) (*GuildSettingsFreeroles, error) {
	row := q.db.QueryRow(ctx, CreateFreeRolesGuildSettings, guildID)
	var i GuildSettingsFreeroles
	err := row.Scan(&i.GuildID, &i.ToggleEnabled, &i.Roles)
	return &i, err
}

const GetFreeRolesGuildSettings = `-- name: GetFreeRolesGuildSettings :one
SELECT
    guild_id, toggle_enabled, roles
FROM
    guild_settings_freeroles
WHERE
    guild_id = $1
`

func (q *Queries) GetFreeRolesGuildSettings(ctx context.Context, guildID int64) (*GuildSettingsFreeroles, error) {
	row := q.db.QueryRow(ctx, GetFreeRolesGuildSettings, guildID)
	var i GuildSettingsFreeroles
	err := row.Scan(&i.GuildID, &i.ToggleEnabled, &i.Roles)
	return &i, err
}

const UpdateFreeRolesGuildSettings = `-- name: UpdateFreeRolesGuildSettings :execrows
UPDATE
    guild_settings_freeroles
SET
    toggle_enabled = $2,
    roles = $3
WHERE
    guild_id = $1
`

type UpdateFreeRolesGuildSettingsParams struct {
	GuildID       int64        `json:"guild_id"`
	ToggleEnabled sql.NullBool `json:"toggle_enabled"`
	Roles         []int64      `json:"roles"`
}

func (q *Queries) UpdateFreeRolesGuildSettings(ctx context.Context, arg *UpdateFreeRolesGuildSettingsParams) (int64, error) {
	result, err := q.db.Exec(ctx, UpdateFreeRolesGuildSettings, arg.GuildID, arg.ToggleEnabled, arg.Roles)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}
